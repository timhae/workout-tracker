package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/donseba/go-htmx"
	"github.com/gin-gonic/gin"
)

type App struct {
	htmx *htmx.HTMX
	// TODO: add gorm for db
}

func main() {
	app := &App{
		htmx: htmx.New(),
	}

	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/workouts")
	})
	router.GET("/workouts", app.Workouts)
	router.GET("/exercises", app.Exercises)
	router.GET("/exercises/create", app.CreateExercise)
	router.GET("/exercises/:id/read", app.ReadExercise)
	router.GET("/exercises/:id/update", app.UpdateExercise)
	router.GET("/exercises/:id/delete", app.DeleteExercise)
	router.GET("/plans", app.Plans)
	router.GET("/measurements", app.Measurements)
	err := router.Run(":8080")
	log.Fatal(err)
}

func (a *App) Workouts(c *gin.Context) {
	data := map[string]any{
		"Text": "Welcome to the workouts page",
	}
	a.render(c, &data, "pages/workouts.html")
}

func (a *App) Exercises(c *gin.Context) {
	data := map[string]any{
		"Exercises": []Exercise{
			{ID: 0, Name: "ex1", Force: Pull, Level: Easy, Mechanic: Compound, Category: Strength},
			{ID: 1, Name: "ex2", Force: Push, Level: Easy, Mechanic: Compound, Category: Strength},
		},
		"Columns": []string{
			"ID", "Name", "Force", "Level", "Mechanic", "Category",
		},
	}
	a.render(c, &data, "pages/exercises.html")
}

func (a *App) CreateExercise(c *gin.Context) {
}

func (a *App) ReadExercise(c *gin.Context) {
	//     id:= c.Param("id")
}

func (a *App) UpdateExercise(c *gin.Context) {

}

func (a *App) DeleteExercise(c *gin.Context) {

}

func (a *App) Plans(c *gin.Context) {
	data := map[string]any{
		"Text": "Welcome to the plans page",
	}
	a.render(c, &data, "pages/plans.html")
}

func (a *App) Measurements(c *gin.Context) {
	data := map[string]any{
		"Text": "Welcome to the measurements page",
	}
	a.render(c, &data, "pages/measurements.html")
}

func (a *App) render(c *gin.Context, data *map[string]any, template string) {
	h := a.htmx.NewHandler(c.Writer, c.Request)
	page := htmx.NewComponent("templates/"+template).SetData(*data).Wrap(mainContent(), "Content")
	_, err := h.Render(c.Request.Context(), page)
	if err != nil {
		fmt.Printf("error rendering page: %v", err.Error())
	}
}

func mainContent() htmx.RenderableComponent {
	menuItems := []struct {
		Name string
		Link string
	}{
		{"Workouts", "/workouts"},
		{"Exercises", "/exercises"},
		{"Plans", "/plans"},
		{"Measurements", "/measurements"},
	}

	data := map[string]any{
		"Title":     "Home",
		"MenuItems": menuItems,
	}

	navbar := htmx.NewComponent("templates/components/navbar.html")
	return htmx.NewComponent("templates/index.html").SetData(data).With(navbar, "Navbar")
}
