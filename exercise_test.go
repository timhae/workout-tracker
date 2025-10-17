package main

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

var (
	mocksql sqlmock.Sqlmock
	exCols  = []string{"ID", "CreatedAt", "UpdatedAt", "Name", "Force", "Level",
		"Mechanic", "Category", "PrimaryMuscle", "SecondaryMuscles",
		"Equipment", "Instructions", "Images"}
	t1, _ = time.Parse("2025-10-11 15:04:09.152093+00", "2006-01-02 15:04:05.999999999Z07:00")
	ex1   = []driver.Value{1, t1, t1, "fff", "0", "0", "0", "0", "0", "[5]",
		"[8]", "asf", `["fff_0", "fff_1"]`}
	t2, _ = time.Parse("2025-10-11 15:08:09.152093+00", "2006-01-02 15:04:05.999999999Z07:00")
	ex2   = []driver.Value{2, t2, t2, "bla", "1", "1", "1", "1", "8", "[1, 5]",
		"[2, 8]", "ddd", "[]"}
)

func TestListExercises(t *testing.T) {
	router := SetupTestApp()

	var tests = []struct {
		rows    *sqlmock.Rows
		fixture string
	}{
		{sqlmock.NewRows(exCols), "./fixtures/exercise/list_empty.html"},
		{sqlmock.NewRows(exCols).AddRow(ex1...), "./fixtures/exercise/list_single.html"},
		{sqlmock.NewRows(exCols).AddRow(ex1...).AddRow(ex2...), "./fixtures/exercise/list_multiple.html"},
	}

	for _, tt := range tests {
		testname := filepath.Base(tt.fixture)
		t.Run(testname, func(t *testing.T) {
			mocksql.ExpectQuery(`SELECT * FROM "exercises" ORDER BY id`).WillReturnRows(tt.rows)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/exercise/list", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			f, _ := os.ReadFile(tt.fixture)
			assert.Equal(t, string(f), w.Body.String())
		})
	}
}

func TestListExercisesError(t *testing.T) {
	router := SetupTestApp()

	mocksql.ExpectQuery(`SELECT * FROM "exercises" ORDER BY id`).WillReturnError(fmt.Errorf("test error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/exercise/list", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	f, _ := os.ReadFile("./fixtures/exercise/list_empty.html")
	assert.Equal(t, string(f), w.Body.String())
}

func TestCreateExercise(t *testing.T) {
	router := SetupTestApp()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/exercise", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	f, _ := os.ReadFile("./fixtures/exercise/form.html")
	assert.Equal(t, string(f), w.Body.String())
}

func TestValidateExercise(t *testing.T) {
	router := SetupTestApp()
	emptyFunc := func() {}
	validateFixture := func(t *testing.T, fixture string, w *httptest.ResponseRecorder) {
		f, _ := os.ReadFile(fixture)
		assert.Equal(t, string(f), w.Body.String())
		// os.WriteFile(tt.fixture, w.Body.Bytes(), 0o644)
	}

	var tests = []struct {
		dbmocks  func()
		form     map[string][]string
		fixture  string
		validate func(*testing.T, string, *httptest.ResponseRecorder)
	}{
		{
			emptyFunc,
			map[string][]string{},
			"./fixtures/exercise/validate_empty.html",
			validateFixture,
		},
		{
			emptyFunc,
			map[string][]string{
				"name": {"test"},
			},
			"./fixtures/exercise/validate_single_value.html",
			validateFixture,
		},
		{
			emptyFunc,
			map[string][]string{
				"name":      {"test"},
				"secondary": {"1", "2"},
			},
			"./fixtures/exercise/validate_w_secondary.html",
			validateFixture,
		},
		{
			func() {
				mocksql.ExpectQuery(`SELECT COUNT("name") FROM "exercises" WHERE name = $1`).
					WithArgs("bla").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			map[string][]string{
				"name":         {"bla"},
				"secondary":    {"1", "2"},
				"equipment":    {"1"},
				"instructions": {"test"},
			},
			"./fixtures/exercise/validate_existing_name.html",
			validateFixture,
		},
		{
			func() {
				mocksql.ExpectQuery(`SELECT COUNT("name") FROM "exercises" WHERE name = $1`).
					WithArgs("test").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mocksql.ExpectBegin()
				mocksql.ExpectQuery(`INSERT INTO "exercises" ("created_at","updated_at","name","force","level","mechanic","category","primary_muscle","secondary_muscles","equipment","instructions","images") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING "id"`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "test", 0, 0, 0, 0, 0, "[1,2]", "[1]", "test", "[]").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mocksql.ExpectCommit()
			},
			map[string][]string{
				"name":         {"test"},
				"secondary":    {"1", "2"},
				"equipment":    {"1"},
				"instructions": {"test"},
			},
			"./nonexistent/validate_valid.html",
			func(t *testing.T, _ string, w *httptest.ResponseRecorder) {
				assert.Equal(t, `{"path":"/exercise/list", "target":"#content"}`, w.Result().Header.Get("HX-Location"))
			},
		},
		// {
		// 	func() {
		// 		mocksql.ExpectQuery(`SELECT COUNT("name") FROM "exercises" WHERE name = $1`).
		// 			WithArgs("test").
		// 			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		// 		mocksql.ExpectBegin()
		// 		mocksql.ExpectQuery(`INSERT INTO "exercises" ("created_at","updated_at","name","force","level","mechanic","category","primary_muscle","secondary_muscles","equipment","instructions","images") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING "id"`).
		// 			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "test", 0, 0, 0, 0, 0, "[1,2]", "[1]", "test", "[]").
		// 			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		// 		mocksql.ExpectCommit()
		// 	},
		// 	map[string][]string{
		// 		"name":         {"test"},
		// 		"secondary":    {"1", "2"},
		// 		"equipment":    {"1"},
		// 		"instructions": {"test"},
		// 	},
		// 	"./nonexistent/validate_valid_with_files.html",
		// 	func(t *testing.T, _ string, w *httptest.ResponseRecorder) {
		// 		assert.Equal(t, `{"path":"/exercise/list", "target":"#content"}`, w.Result().Header.Get("HX-Location"))
		// 	},
		// },
	}

	for _, tt := range tests {
		testname := filepath.Base(tt.fixture)
		t.Run(testname, func(t *testing.T) {
			w := httptest.NewRecorder()
			var b bytes.Buffer
			wr := multipart.NewWriter(&b)
			for key, vals := range tt.form {
				for _, v := range vals {
					_ = wr.WriteField(key, v)
				}
			}
			wr.Close()
			req, _ := http.NewRequest("POST", "/exercise/validate", &b)
			req.Header.Set("Content-Type", wr.FormDataContentType())
			tt.dbmocks()
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			tt.validate(t, tt.fixture, w)
		})
	}
}
