package main

import (
	"context"
	"log"
	"net/http"

	"github.com/donseba/go-htmx"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	htmx *htmx.HTMX
	db   *gorm.DB
	ctx  *context.Context
}

func main() {
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

	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Static("/static", "./static")
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/workout/list")
	})
	router.GET("/workout/list", app.ListWorkouts)
	router.GET("/plan/list", app.ListPlans)
	router.GET("/measurement/list", app.ListMeasurements)

	ex := router.Group("/exercise")
	ex.GET("/list", app.ListExercises)
	ex.GET("", app.CreateExercise)
	ex.POST("/validate", app.ValidateExercise)
	ex.GET("/:id", app.ReadExercise)
	ex.DELETE("/:id", app.DeleteExercise)
	ex.POST("/:id/validate", app.ValidateExercise)

	err = router.Run(":8080")
	log.Fatal(err)
}

func (a *App) ListWorkouts(c *gin.Context) {
	data := map[string]any{
		"Text": "Welcome to the workouts page",
	}
	page := htmx.NewComponent("templates/pages/workouts.html").SetData(data).Wrap(mainContent(), "Content")
	a.render(c, &page)
}

func (a *App) ListPlans(c *gin.Context) {
	data := map[string]any{
		"Text": "Welcome to the plans page",
	}
	page := htmx.NewComponent("templates/pages/plans.html").SetData(data).Wrap(mainContent(), "Content")
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
		log.Printf("error rendering page: %v", err.Error())
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
