package main

import (
	"context"
	"log"
	"net/http"

	"github.com/donseba/go-htmx"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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
		c.Redirect(http.StatusMovedPermanently, "/workouts")
	})
	router.GET("/workouts", app.Workouts)
	router.GET("/plans", app.Plans)
	router.GET("/measurements", app.Measurements)

	router.GET("/exercises", app.Exercises)
	router.GET("/exercise", app.CreateExercise)
	router.POST("/exercise", app.CreateExercise)
	router.GET("/exercise/:id", app.ReadExercise)
	router.PUT("/exercise/:id", app.ReadExercise)
	router.DELETE("/exercise/:id", app.DeleteExercise)

	err = router.Run(":8080")
	log.Fatal(err)
}

func (a *App) Workouts(c *gin.Context) {
	data := map[string]any{
		"Text": "Welcome to the workouts page",
	}

	a.render(c, &data, "pages/workouts.html")
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

func (a *App) Exercises(c *gin.Context) {
	var exercises []Exercise
	exercises, err := gorm.G[Exercise](a.db).Find(*a.ctx)
	if err != nil {
		log.Printf("error fetching exercises: %v", err)
	}

	data := map[string]any{
		"Exercises": exercises,
		"Columns": []string{
			"ID", "Delete", "Name", "Force", "Level", "Mechanic", "Category", "Primary", "Secondary", "Equipment", "Instructions", "Images",
		},
	}
	a.render(c, &data, "pages/exercises.html")
}

func (a *App) Exercise(c *gin.Context, exercise Exercise, errorMsg string) {
	validationRequest := c.Request.Header.Get("X-Validation") == "true"

	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		if err := c.ShouldBindWith(&exercise, binding.Form); err != nil {
			log.Printf("input err: %v+", err)
			errorMsg = err.Error()
		}

		if errorMsg == "" && !validationRequest {
			form, err := c.MultipartForm()
			if err != nil {
				log.Printf("form error: %v+", err)
				errorMsg = err.Error()
			} else {
				files := form.File["images"]
				fileNames := ""

				for _, file := range files {
					log.Printf("saving file: %v+", file.Filename)
					err := c.SaveUploadedFile(file, "./static/images/"+file.Filename)
					fileNames += file.Filename + ";"
					if err != nil {
						log.Printf("upload error: %v+", err)
						errorMsg += err.Error()
					}
				}
				exercise.Images = fileNames
			}
		}

		if errorMsg == "" && !validationRequest {
			switch c.Request.Method {
			case "POST":
				err := gorm.G[Exercise](a.db).Create(*a.ctx, &exercise)
				if err != nil {
					log.Printf("db err: %v+", err)
					errorMsg = err.Error()
				}
			case "PUT":
				_, err := gorm.G[Exercise](a.db).Where("id = ?", c.Param("id")).Updates(*a.ctx, exercise)
				if err != nil {
					log.Printf("db err: %v+", err)
					errorMsg = err.Error()
				}
			}
		}

		if errorMsg == "" && !validationRequest {
			c.Header("HX-Location", `{"path":"/exercises", "target":"#content"}`)
			return
		}
	}

	data := map[string]any{
		"PossibleValues": map[string]any{
			"Forces":     AllValues[Force](uint(_ForceCount)),
			"Levels":     AllValues[Level](uint(_LevelCount)),
			"Mechanics":  AllValues[Mechanic](uint(_MechanicCount)),
			"Categories": AllValues[Category](uint(_CategoryCount)),
			"Muscles":    AllValues[Muscle](uint(_MuscleCount)),
			"Equipment":  AllValues[Equipment](uint(_EquipmentCount)),
		},
		"Input":  exercise,
		"ID":     c.Param("id"),
		"Error":  errorMsg,
	}
	a.render(c, &data, "pages/exercise.html")
}

func (a *App) CreateExercise(c *gin.Context) {
	exercise := Exercise{}
	a.Exercise(c, exercise, "")
}

func (a *App) ReadExercise(c *gin.Context) {
	validationRequest := c.Request.Header.Get("X-Validation") == "true"
	var errorMsg = ""

	if !validationRequest {
		id := c.Param("id")
		exercise, err := gorm.G[Exercise](a.db).Where("id = ?", id).First(*a.ctx)
		if err != nil {
			log.Printf("error reading exercise: %v", err)
			errorMsg = err.Error()
		}
		a.Exercise(c, exercise, errorMsg)
	} else {
		a.Exercise(c, Exercise{}, errorMsg)
	}
}

func (a *App) DeleteExercise(c *gin.Context) {
	_, err := gorm.G[Exercise](a.db).Where("id = ?", c.Param("id")).Delete(*a.ctx)
	if err != nil {
		log.Printf("error deleting exercise: %v", err)
	}
	a.Exercises(c)
}

func (a *App) render(c *gin.Context, data *map[string]any, template string) {
	h := a.htmx.NewHandler(c.Writer, c.Request)
	page := htmx.NewComponent("templates/"+template).SetData(*data).Wrap(mainContent(), "Content")
	_, err := h.Render(c.Request.Context(), page)
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
			{"Workouts", "/workouts"},
			{"Plans", "/plans"},
			{"Measurements", "/measurements"},
			{"Exercises", "/exercises"},
		},
	}

	navbar := htmx.NewComponent("templates/components/navbar.html")
	return htmx.NewComponent("templates/index.html").SetData(data).With(navbar, "Navbar")
}

func AllValues[T ~uint](count uint) []T {
	out := make([]T, count)
	for i := range out {
		out[i] = T(i)
	}
	return out
}
