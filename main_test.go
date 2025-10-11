package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
	"path/filepath"
	"testing"

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
	ex1 = []driver.Value{1, t1, t1, "fff", "0", "0", "0", "0", "0", "[5]",
		"[8]", "asf", "PXL_20221129_202949801.jpg;PXL_20221129_203004190.jpg;"}
	t2, _ = time.Parse("2025-10-11 15:08:09.152093+00", "2006-01-02 15:04:05.999999999Z07:00")
	ex2 = []driver.Value{2, t2, t2, "bla", "1", "1", "1", "1", "8", "[1, 5]",
		"[2, 8]", "ddd", "PXL_20221129_203004190.jpg;"}
)

func SetupApp() *App {
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
	return app
}

func TestExercises(t *testing.T) {
	app := SetupApp()
	router := gin.Default()
	router.GET("/exercises", app.Exercises)

	var tests = []struct {
		rows    *sqlmock.Rows
		fixture string
	}{
		{sqlmock.NewRows(exCols), "./fixtures/exercises_empty.html"},
		{sqlmock.NewRows(exCols).AddRow(ex1...), "./fixtures/exercises_single.html"},
		{sqlmock.NewRows(exCols).AddRow(ex1...).AddRow(ex2...), "./fixtures/exercises_multiple.html"},
	}

	for _, tt := range tests {
		testname := filepath.Base(tt.fixture)
		t.Run(testname, func(t *testing.T) {
			mock.ExpectQuery(`SELECT * FROM "exercises" ORDER BY id`).WillReturnRows(tt.rows)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/exercises", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			f, _ := os.ReadFile(tt.fixture)
			assert.Equal(t, string(f), w.Body.String())
		})
	}
}
