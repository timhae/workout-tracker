package main

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/donseba/go-htmx"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	mock   sqlmock.Sqlmock
	exCols = []string{"ID", "CreatedAt", "UpdatedAt", "Name", "Force", "Level",
		"Mechanic", "Category", "PrimaryMuscle", "SecondaryMuscles",
		"Equipment", "Instructions", "Images"}
	t1, _ = time.Parse("2025-10-11 15:04:09.152093+00", "2006-01-02 15:04:05.999999999Z07:00")
	ex1   = []driver.Value{1, t1, t1, "fff", "0", "0", "0", "0", "0", "[5]",
		"[8]", "asf", `["fff_0", "fff_1"]`}
	t2, _ = time.Parse("2025-10-11 15:08:09.152093+00", "2006-01-02 15:04:05.999999999Z07:00")
	ex2   = []driver.Value{2, t2, t2, "bla", "1", "1", "1", "1", "8", "[1, 5]",
		"[2, 8]", "ddd", "[]"}
)

func SetupApp() (*App, *gin.Engine) {
	var mockDb *sql.DB
	mockDb, mock, _ = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	db, _ := gorm.Open(postgres.New(postgres.Config{
		Conn:       mockDb,
		DriverName: "postgres",
	}), &gorm.Config{})
	ctx := context.Background()
	app := &App{
		htmx: htmx.New(),
		db:   db,
		ctx:  &ctx,
	}
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	return app, router
}

func TestListExercises(t *testing.T) {
	app, router := SetupApp()
	router.GET("/exercise/list", app.ListExercises)

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
			mock.ExpectQuery(`SELECT * FROM "exercises" ORDER BY id`).WillReturnRows(tt.rows)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/exercise/list", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			f, _ := os.ReadFile(tt.fixture)
			assert.Equal(t, string(f), w.Body.String())
		})
	}
}

func TestCreateExercise(t *testing.T) {
	app, router := SetupApp()
	router.GET("/exercise", app.CreateExercise)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/exercise", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	f, _ := os.ReadFile("./fixtures/exercise/form.html")
	assert.Equal(t, string(f), w.Body.String())
}

func TestValidateExercise(t *testing.T) {
	app, router := SetupApp()
	router.POST("/exercise/validate", app.ValidateExercise)
	router.POST("/exercise/:id/validate", app.ValidateExercise)

	var tests = []struct {
		form    map[string][]string
		fixture string
	}{
		{map[string][]string{}, "./fixtures/exercise/validate_empty.html"},
		{map[string][]string{
			"name": {"test"},
		}, "./fixtures/exercise/validate_single_value.html"},
		{map[string][]string{
			"name":      {"test"},
			"secondary": {"1", "2"},
		}, "./fixtures/exercise/validate_w_secondary.html"},
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
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			f, _ := os.ReadFile(tt.fixture)
			assert.Equal(t, string(f), w.Body.String())
		})
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
			req, _ := http.NewRequest("POST", "/exercise/validate/42", &b)
			req.Header.Set("Content-Type", wr.FormDataContentType())
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			f, _ := os.ReadFile(tt.fixture)
			assert.Equal(t, string(f), w.Body.String())
		})
	}
}
