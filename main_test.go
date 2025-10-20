package main

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/donseba/go-htmx"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupTestApp() (*gin.Engine, *App) {
	var mockDb *sql.DB
	mockDb, mocksql, _ = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	db, _ := gorm.Open(postgres.New(postgres.Config{
		Conn:       mockDb,
		DriverName: "postgres",
	}), &gorm.Config{})
	ctx := context.Background()
	app := &App{
		htmx:   htmx.New(),
		db:     db,
		ctx:    &ctx,
		mockFS: &mockFS{},
	}
	router := app.setupRouter(gin.TestMode)
	return router, app
}

func TestHomeRedirect(t *testing.T) {
	router, _ := SetupTestApp()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 301, w.Code)
	assert.Equal(t, "/workout/list", w.Result().Header.Get("Location"))
}
