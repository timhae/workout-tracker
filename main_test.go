package main

import (
	"context"
	"database/sql"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/donseba/go-htmx"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type mockFileSaver struct {
	mock.Mock
}

func (m *mockFileSaver) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	args := m.Called(file, dst)
	return args.Error(0)
}

func SetupTestApp() *gin.Engine {
	var mockDb *sql.DB
	mockDb, mocksql, _ = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	db, _ := gorm.Open(postgres.New(postgres.Config{
		Conn:       mockDb,
		DriverName: "postgres",
	}), &gorm.Config{})
	ctx := context.Background()
	app := &App{
		htmx:      htmx.New(),
		db:        db,
		ctx:       &ctx,
		fileSaver: &mockFileSaver{},
	}
	router := app.setupRouter(gin.TestMode)
	return router
}

func TestHomeRedirect(t *testing.T) {
	router := SetupTestApp()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 301, w.Code)
	assert.Equal(t, "/workout/list", w.Result().Header.Get("Location"))
}
