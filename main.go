package main

import (
	"context"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/donseba/go-htmx"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type mockRM struct {
	mock.Mock
}

func (m *mockRM) Remove(file string) error {
	log.Printf("mock removing file %s", file)
	args := m.Called(file)
	return args.Error(0)
}

type mockFS struct {
	mock.Mock
}

func (m *mockFS) SaveUploadedFile(file *multipart.FileHeader, dst string, perm ...fs.FileMode) error {
	log.Printf("mock saving file %s as %s", file.Filename, dst)
	args := m.Called(file, dst)
	return args.Error(0)
}

type App struct {
	htmx   *htmx.HTMX
	db     *gorm.DB
	ctx    *context.Context
	mockFS *mockFS
	mockRM *mockRM
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Europe/Berlin",
	}), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.AutoMigrate(&Exercise{})
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	app := &App{
		htmx: htmx.New(),
		db:   db,
		ctx:  &ctx,
	}

	router := app.setupRouter(gin.DebugMode)
	err = router.Run(":8080")
	log.Fatal(err)
}

func (a *App) setupRouter(mode string) *gin.Engine {
	gin.SetMode(mode)

	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Static("/static", "./static")
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/workout/list")
	})
	router.GET("/workout/list", a.ListWorkouts)
	router.GET("/measurement/list", a.ListMeasurements)

	ex := router.Group("/exercise")
	ex.GET("/list", a.ListExercises)
	ex.POST("/list", a.ListExercisesWithFilter)
	ex.GET("", a.CreateExercise)
	ex.POST("/validate", a.ValidateExercise)
	ex.GET("/:id", a.ReadExercise)
	ex.DELETE("/:id", a.DeleteExercise)
	ex.POST("/:id/validate", a.ValidateExercise)

	plan := router.Group("/plan")
	plan.GET("/list", a.ListPlans)
	plan.GET("", a.CreatePlan)

	return router
}

func (a *App) ListWorkouts(c *gin.Context) {
	data := map[string]any{
		"Text": "Welcome to the workouts page",
	}
	page := htmx.NewComponent("templates/pages/workouts.html").SetData(data).Wrap(mainContent(), "Content")
	a.render(c, &page)
}

func (a *App) ListMeasurements(c *gin.Context) {
	data := map[string]any{
		"Text": "Welcome to the measurements page",
	}
	page := htmx.NewComponent("templates/pages/measurements.html").SetData(data).Wrap(mainContent(), "Content")
	a.render(c, &page)
}

func (a *App) render(c *gin.Context, page *htmx.RenderableComponent) {
	htmx := a.htmx.NewHandler(c.Writer, c.Request)
	_, err := htmx.Render(c.Request.Context(), *page)
	if err != nil {
		log.Printf("render error: %v", err.Error())
	}
}

func mainContent() htmx.RenderableComponent {
	data := map[string]any{
		"MenuItems": []struct {
			Name string
			Link string
		}{
			{"Workouts", "/workout/list"},
			{"Plans", "/plan/list"},
			{"Measurements", "/measurement/list"},
			{"Exercises", "/exercise/list"},
		},
	}
	navbar := htmx.NewComponent("templates/components/navbar.html")
	return htmx.NewComponent("templates/index.html").SetData(data).With(navbar, "Navbar")
}

func allValues[T ~uint](count uint) []T {
	out := make([]T, count)
	for i := range out {
		out[i] = T(i)
	}
	return out
}
