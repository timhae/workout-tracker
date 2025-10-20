package main

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	validateFixture = func(t *testing.T, fixture string, w *httptest.ResponseRecorder, a *App) {
		f, _ := os.ReadFile(fixture)
		assert.Equal(t, string(f), w.Body.String())
		// os.WriteFile(fixture, w.Body.Bytes(), 0o644)
	}
)

func TestListExercises(t *testing.T) {
	router, app := SetupTestApp()

	tests := []struct {
		dbmocks func()
		fixture string
	}{
		{
			func() {
				mocksql.ExpectQuery(`SELECT * FROM "exercises" ORDER BY id`).
					WillReturnRows(sqlmock.NewRows(exCols))
			},
			"./fixtures/exercise/list_empty.html",
		},
		{
			func() {
				mocksql.ExpectQuery(`SELECT * FROM "exercises" ORDER BY id`).
					WillReturnRows(sqlmock.NewRows(exCols).AddRow(ex1...))
			},
			"./fixtures/exercise/list_single.html",
		},
		{
			func() {
				mocksql.ExpectQuery(`SELECT * FROM "exercises" ORDER BY id`).
					WillReturnRows(sqlmock.NewRows(exCols).AddRow(ex1...).AddRow(ex2...))
			},
			"./fixtures/exercise/list_multiple.html",
		},
		{
			func() {
				mocksql.ExpectQuery(`SELECT * FROM "exercises" ORDER BY id`).
					WillReturnError(fmt.Errorf("test list error"))
			},
			"./fixtures/exercise/list_empty.html",
		},
	}

	for _, tt := range tests {
		testname := filepath.Base(tt.fixture)
		t.Run(testname, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/exercise/list", nil)
			tt.dbmocks()
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			validateFixture(t, tt.fixture, w, app)
			if err := mocksql.ExpectationsWereMet(); err != nil {
				t.Fatalf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestCreateExercise(t *testing.T) {
	router, app := SetupTestApp()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/exercise", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	validateFixture(t, "./fixtures/exercise/form.html", w, app)
}

func TestValidateExercise(t *testing.T) {
	router, app := SetupTestApp()

	tests := []struct {
		dbmocks  func(*App)
		form     map[string][]string
		fixture  string
		validate func(*testing.T, string, *httptest.ResponseRecorder, *App)
	}{
		{
			func(a *App) {},
			map[string][]string{},
			"./fixtures/exercise/validate_empty.html",
			validateFixture,
		},
		{
			func(a *App) {},
			map[string][]string{
				"name": {"test"},
			},
			"./fixtures/exercise/validate_single_value.html",
			validateFixture,
		},
		{
			func(a *App) {},
			map[string][]string{
				"name":      {"test"},
				"secondary": {"1", "2"},
			},
			"./fixtures/exercise/validate_w_secondary.html",
			validateFixture,
		},
		{
			func(a *App) {
				mocksql.ExpectQuery(`SELECT COUNT("name") FROM "exercises" WHERE name = $1`).
					WithArgs("test").
					WillReturnError(fmt.Errorf("test count error"))
			},
			map[string][]string{
				"name":         {"test"},
				"secondary":    {"1", "2"},
				"equipment":    {"1"},
				"instructions": {"test"},
			},
			"./fixtures/exercise/validate_count_error.html",
			validateFixture,
		},
		{
			func(a *App) {
				mocksql.ExpectQuery(`SELECT COUNT("name") FROM "exercises" WHERE name = $1`).
					WithArgs("test").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mocksql.ExpectBegin()
				mocksql.ExpectQuery(`INSERT INTO "exercises" ("created_at","updated_at","name","force","level","mechanic","category","primary_muscle","secondary_muscles","equipment","instructions","images") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING "id"`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "test", 0, 0, 0, 0, 0, "[1,2]", "[1]", "test", "[]").
					WillReturnError(fmt.Errorf("test insert error"))
				mocksql.ExpectRollback()
			},
			map[string][]string{
				"name":         {"test"},
				"secondary":    {"1", "2"},
				"equipment":    {"1"},
				"instructions": {"test"},
			},
			"./fixtures/exercise/validate_insert_error.html",
			validateFixture,
		},
		{
			func(a *App) {
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
			func(a *App) {},
			map[string][]string{
				"name":         {"test"},
				"secondary":    {"1", "2"},
				"equipment":    {"1"},
				"instructions": {"test"},
			},
			"./fixtures/exercise/validate_validation_request.html",
			validateFixture,
		},
		{
			func(a *App) {
				mocksql.ExpectQuery(`SELECT COUNT("name") FROM "exercises" WHERE name = $1`).
					WithArgs("test").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mockFS := &mockFS{}
				mockFS.On("SaveUploadedFile", mock.Anything, "./static/images/test_0").Return(nil)
				mockFS.On("SaveUploadedFile", mock.Anything, "./static/images/test_1").Return(fmt.Errorf("save file error"))
				a.mockFS = mockFS
			},
			map[string][]string{
				"name":         {"test"},
				"secondary":    {"1", "2"},
				"equipment":    {"1"},
				"instructions": {"test"},
				"images":       {"img1", "img2"},
			},
			"./fixtures/exercise/validate_valid_with_files_upload_error.html",
			validateFixture,
		},
		{
			func(a *App) {
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
			func(t *testing.T, _ string, w *httptest.ResponseRecorder, a *App) {
				assert.Equal(t, `{"path":"/exercise/list", "target":"#content"}`, w.Result().Header.Get("HX-Location"))
			},
		},
		{
			func(a *App) {
				mocksql.ExpectQuery(`SELECT COUNT("name") FROM "exercises" WHERE name = $1`).
					WithArgs("test").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mocksql.ExpectBegin()
				mocksql.ExpectQuery(`INSERT INTO "exercises" ("created_at","updated_at","name","force","level","mechanic","category","primary_muscle","secondary_muscles","equipment","instructions","images") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING "id"`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "test", 0, 0, 0, 0, 0, "[1,2]", "[1]", "test", `["test_0","test_1"]`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mocksql.ExpectCommit()
				mockFS := &mockFS{}
				mockFS.On("SaveUploadedFile", mock.Anything, "./static/images/test_0").Return(nil)
				mockFS.On("SaveUploadedFile", mock.Anything, "./static/images/test_1").Return(nil)
				a.mockFS = mockFS
			},
			map[string][]string{
				"name":         {"test"},
				"secondary":    {"1", "2"},
				"equipment":    {"1"},
				"instructions": {"test"},
				"images":       {"img1", "img2"},
			},
			"./nonexistent/validate_valid_with_files.html",
			func(t *testing.T, _ string, w *httptest.ResponseRecorder, a *App) {
				assert.Equal(t, `{"path":"/exercise/list", "target":"#content"}`, w.Result().Header.Get("HX-Location"))
				a.mockFS.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		testname := filepath.Base(tt.fixture)
		t.Run(testname, func(t *testing.T) {
			w := httptest.NewRecorder()
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			for key, vals := range tt.form {
				for _, val := range vals {
					if key == "images" {
						part, _ := writer.CreateFormFile(key, val)
						empty := make([]byte, 32)
						file := bytes.NewReader(empty)
						io.Copy(part, file)
					} else {
						writer.WriteField(key, val)
					}
				}
			}
			writer.Close()
			req, _ := http.NewRequest("POST", "/exercise/validate", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			if strings.Contains(testname, "validation_request") {
				req.Header.Set("X-Validation-Only", "true")
			}
			tt.dbmocks(app)
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			tt.validate(t, tt.fixture, w, app)
			if err := mocksql.ExpectationsWereMet(); err != nil {
				t.Fatalf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestValidateExerciseWithID(t *testing.T) {
	router, app := SetupTestApp()

	var tests = []bool{false, true}

	for _, validationOnly := range tests {
		testname := fmt.Sprintf("validate_with_id_validation_only_%t.html", validationOnly)
		t.Run(testname, func(t *testing.T) {
			w := httptest.NewRecorder()
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			writer.WriteField("name", "test")
			writer.WriteField("force", "2")
			writer.WriteField("secondary", "2")
			writer.WriteField("equipment", "1")
			writer.WriteField("instructions", "test")
			writer.Close()
			req, _ := http.NewRequest("POST", "/exercise/42/validate", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			if validationOnly {
				req.Header.Set("X-Validation-Only", "true")
			} else {
				mocksql.ExpectBegin()
				mocksql.ExpectExec(`UPDATE "exercises" SET "updated_at"=$1,"name"=$2,"force"=$3,"secondary_muscles"=$4,"equipment"=$5,"instructions"=$6,"images"=$7 WHERE id = $8`).
					WithArgs(sqlmock.AnyArg(), "test", 2, "[2]", "[1]", "test", "[]", "42").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mocksql.ExpectCommit()
			}
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			if validationOnly {
				validateFixture(t, "./fixtures/exercise/validate_valid_with_id.html", w, app)
			} else {
				assert.Equal(t, `{"path":"/exercise/list", "target":"#content"}`, w.Result().Header.Get("HX-Location"))
			}
			if err := mocksql.ExpectationsWereMet(); err != nil {
				t.Fatalf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestReadExercises(t *testing.T) {
	router, app := SetupTestApp()

	tests := []struct {
		dbmocks func()
		fixture string
	}{
		{
			func() {
				mocksql.ExpectQuery(`SELECT * FROM "exercises" WHERE id = $1 ORDER BY "exercises"."id" LIMIT $2`).
					WithArgs("2", 1).
					WillReturnRows(sqlmock.NewRows(exCols).AddRow(ex2...))
			},
			"./fixtures/exercise/read.html",
		},
		{
			func() {
				mocksql.ExpectQuery(`SELECT * FROM "exercises" WHERE id = $1 ORDER BY "exercises"."id" LIMIT $2`).
					WithArgs("2", 1).
					WillReturnError(fmt.Errorf("test read error"))
			},
			"./fixtures/exercise/read_error.html",
		},
	}

	for _, tt := range tests {
		testname := filepath.Base(tt.fixture)
		t.Run(testname, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/exercise/2", nil)
			tt.dbmocks()
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			validateFixture(t, tt.fixture, w, app)
			if err := mocksql.ExpectationsWereMet(); err != nil {
				t.Fatalf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestDeleteExercises(t *testing.T) {
	router, app := SetupTestApp()

	tests := []struct {
		dbmocks func()
		fixture string
	}{
		{
			func() {
				mocksql.ExpectBegin()
				mocksql.ExpectExec(`DELETE FROM "exercises" WHERE id = $1`).
					WithArgs("2").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mocksql.ExpectCommit()
				mocksql.ExpectQuery(`SELECT * FROM "exercises" ORDER BY id`).
					WillReturnRows(sqlmock.NewRows(exCols).AddRow(ex1...))
			},
			"./fixtures/exercise/list_single.html",
		},
		{
			func() {
				mocksql.ExpectBegin()
				mocksql.ExpectExec(`DELETE FROM "exercises" WHERE id = $1`).
					WithArgs("2").
					WillReturnError(fmt.Errorf("test delete error"))
				mocksql.ExpectRollback()
				mocksql.ExpectQuery(`SELECT * FROM "exercises" ORDER BY id`).
					WillReturnRows(sqlmock.NewRows(exCols).AddRow(ex1...).AddRow(ex2...))
			},
			"./fixtures/exercise/list_multiple.html",
		},
	}

	for _, tt := range tests {
		testname := filepath.Base(tt.fixture)
		t.Run(testname, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", "/exercise/2", nil)
			tt.dbmocks()
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			validateFixture(t, tt.fixture, w, app)
			if err := mocksql.ExpectationsWereMet(); err != nil {
				t.Fatalf("unfulfilled expectations: %v", err)
			}
		})
	}
}
