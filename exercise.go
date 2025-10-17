package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"strconv"
	"time"

	"github.com/donseba/go-htmx"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type Exercise struct {
	ID               uint
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Name             string      `form:"name" binding:"required"`
	Force            Force       `form:"force" binding:"number,gte=0"`
	Level            Level       `form:"level" binding:"number,gte=0"`
	Mechanic         Mechanic    `form:"mechanic" binding:"number,gte=0"`
	Category         Category    `form:"category" binding:"number,gte=0"`
	PrimaryMuscle    Muscle      `form:"primary" binding:"number,gte=0"`
	SecondaryMuscles []Muscle    `form:"secondary" binding:"required" gorm:"type:jsonb;serializer:json"`
	Equipment        []Equipment `form:"equipment" binding:"required" gorm:"type:jsonb;serializer:json"`
	Instructions     string      `form:"instructions" binding:"required" gorm:"type:text"`
	Images           []string    `gorm:"type:jsonb;serializer:json"`
}

type Force uint

const (
	Pull Force = iota
	Push
	Static

	_ForceCount
)

var forceName = map[Force]string{
	Pull:   "Pull",
	Push:   "Push",
	Static: "Static",
}

func (f Force) String() string {
	return forceName[f]
}

type Level uint

const (
	Easy Level = iota
	Middle
	Hard

	_LevelCount
)

var levelName = map[Level]string{
	Easy:   "Easy",
	Middle: "Middle",
	Hard:   "Hard",
}

func (l Level) String() string {
	return levelName[l]
}

type Mechanic uint

const (
	Compound Mechanic = iota
	Isolation

	_MechanicCount
)

var mechanicName = map[Mechanic]string{
	Compound:  "Compound",
	Isolation: "Isolation",
}

func (m Mechanic) String() string {
	return mechanicName[m]
}

type Category uint

const (
	Endurance Category = iota
	Strength
	Stretching

	_CategoryCount
)

var categoryName = map[Category]string{
	Endurance:  "Endurance",
	Strength:   "Strength",
	Stretching: "Stretching",
}

func (c Category) String() string {
	return categoryName[c]
}

type Equipment uint

const (
	Bands Equipment = iota
	Barbell
	Bench
	Body
	Cable
	Dumbbells
	Kettlebells
	Machine
	Other

	_EquipmentCount
)

var equipmentName = map[Equipment]string{
	Bands:       "Bands",
	Barbell:     "Barbell",
	Bench:       "Bench",
	Body:        "Body",
	Cable:       "Cable",
	Dumbbells:   "Dumbbells",
	Kettlebells: "Kettlebells",
	Machine:     "Machine",
	Other:       "Other",
}

func (e Equipment) String() string {
	return equipmentName[e]
}

func (m Equipment) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint(m))
}

func (m *Equipment) UnmarshalJSON(data []byte) error {
	var v uint
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*m = Equipment(v)
	return nil
}

type Muscle uint

const (
	Abdominals Muscle = iota
	Abductors
	Adductors
	Biceps
	Calves
	Chest
	Forearms
	Glutes
	Hamstrings
	Lats
	LowerBack
	Neck
	Quadriceps
	Shoulders
	Traps
	Triceps

	_MuscleCount
)

var muscleName = map[Muscle]string{
	Abdominals: "Abdominals",
	Abductors:  "Abductors",
	Adductors:  "Adductors",
	Biceps:     "Biceps",
	Calves:     "Calves",
	Chest:      "Chest",
	Forearms:   "Forearms",
	Glutes:     "Glutes",
	Hamstrings: "Hamstrings",
	Lats:       "Lats",
	LowerBack:  "LowerBack",
	Neck:       "Neck",
	Quadriceps: "Quadriceps",
	Shoulders:  "Shoulders",
	Traps:      "Traps",
	Triceps:    "Triceps",
}

func (m Muscle) String() string {
	return muscleName[m]
}

func (m Muscle) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint(m))
}

func (m *Muscle) UnmarshalJSON(data []byte) error {
	var v uint
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*m = Muscle(v)
	return nil
}

var (
	possibleValues = map[string]any{
		"Forces":     allValues[Force](uint(_ForceCount)),
		"Levels":     allValues[Level](uint(_LevelCount)),
		"Mechanics":  allValues[Mechanic](uint(_MechanicCount)),
		"Categories": allValues[Category](uint(_CategoryCount)),
		"Muscles":    allValues[Muscle](uint(_MuscleCount)),
		"Equipment":  allValues[Equipment](uint(_EquipmentCount)),
	}
)

func (a *App) ListExercises(c *gin.Context) {
	var exercises []Exercise
	exercises, err := gorm.G[Exercise](a.db).Order("id").Find(*a.ctx)
	if err != nil {
		log.Printf("error fetching exercises: %v", err)
	}

	data := map[string]any{
		"Exercises": exercises,
		"Columns": []string{
			"Action", "Name", "Force", "Level", "Mechanic", "Category", "Primary", "Secondary", "Equipment", "Instructions", "Images",
		},
		"Actions": []string{"Del", "Edit"},
	}
	page := htmx.NewComponent("templates/pages/exercises.html").SetData(data).Wrap(mainContent(), "Content").AddTemplateFunction("exerciseAction", exerciseAction)
	a.render(c, &page)
}

func (a *App) CreateExercise(c *gin.Context) {
	data := map[string]any{
		"PossibleValues": possibleValues,
		"ValidationLink": template.HTMLAttr(`hx-post="/exercise/validate"`),
		"Input":          Exercise{},
		"Button":         "Create",
	}
	page := htmx.NewComponent("templates/components/exercise_form.html").SetData(data).Wrap(mainContent(), "Content")
	a.render(c, &page)
}

func (a *App) ReadExercise(c *gin.Context) {
	id := c.Param("id")
	exercise, err := gorm.G[Exercise](a.db).Where("id = ?", id).First(*a.ctx)
	if err != nil {
		log.Printf("error reading exercise: %v", err)
	} else {
		err = errors.New("")
	}

	data := map[string]any{
		"PossibleValues": possibleValues,
		"ValidationLink": template.HTMLAttr(`hx-post="/exercise/` + id + `/validate"`),
		"Input":          exercise,
		"Error":          err.Error(),
		"Button":         "Update",
	}
	page := htmx.NewComponent("templates/components/exercise_form.html").SetData(data).Wrap(mainContent(), "Content")
	a.render(c, &page)
}

func (a *App) DeleteExercise(c *gin.Context) {
	_, err := gorm.G[Exercise](a.db).Where("id = ?", c.Param("id")).Delete(*a.ctx)
	if err != nil {
		log.Printf("error deleting exercise: %v", err)
	}
	a.ListExercises(c)
}

func (a *App) ValidateExercise(c *gin.Context) {
	var err error
	var exercise Exercise
	button := "Create"
	validationLink := `hx-post="/exercise/validate"`
	validationRequest := c.Request.Header.Get("X-Validation-Only") == "true"
	id := c.Param("id")

	if err = c.ShouldBindWith(&exercise, binding.FormMultipart); err != nil {
		log.Printf("input err: %v+", err)
	} else if !validationRequest {
		err = a.insertExercise(id, c, &exercise)
		if err != nil {
			log.Printf("insert err: %v+", err)
		}
	}

	if err == nil {
		if !validationRequest {
			c.Header("HX-Location", `{"path":"/exercise/list", "target":"#content"}`)
			return
		} else {
			err = errors.New("")
		}
	}

	if id != "" {
		button = "Update"
		validationLink = `hx-post="/exercise/` + id + `/validate"`
	}

	data := map[string]any{
		"PossibleValues": possibleValues,
		"ValidationLink": template.HTMLAttr(validationLink),
		"Input":          exercise,
		"Error":          err.Error(),
		"Button":         button,
	}
	page := htmx.NewComponent("templates/components/exercise_form.html").SetData(data).Wrap(mainContent(), "Content")
	a.render(c, &page)
}

func (a *App) insertExercise(id string, c *gin.Context, exercise *Exercise) error {
	var fileNames []string

	if id == "" {
		count, err := gorm.G[Exercise](a.db).Where("name = ?", exercise.Name).Count(*a.ctx, "name")
		if err != nil {
			return err
		}
		if count > 0 {
			return errors.New("exercise with name '" + exercise.Name + "' already exists")
		}
	}

	fileNames, err := a.saveImages(exercise.Name, c)
	if err != nil {
		log.Printf("file upload err: %v+", err)
		return err
	}
	exercise.Images = fileNames

	if id != "" {
		_, err = gorm.G[Exercise](a.db).Where("id = ?", c.Param("id")).Updates(*a.ctx, *exercise)
	} else {
		err = gorm.G[Exercise](a.db).Create(*a.ctx, exercise)
	}
	if err != nil {
		log.Printf("db err: %v+", err)
		return err
	}

	return nil
}

func (a *App) saveImages(name string, c *gin.Context) ([]string, error) {
	var err error
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}
	files := form.File["images"]
	fileNames := []string{}

	for idx, file := range files {
		fileName := name + "_" + strconv.Itoa(idx)
		log.Printf("saving file %s as %s", file.Filename, fileName)
		if a.fileSaver != nil {
			err = a.fileSaver.SaveUploadedFile(file, "./static/images/"+fileName)
		} else {
			err = c.SaveUploadedFile(file, "./static/images/"+fileName)
		}
		if err != nil {
			return nil, err
		}
		fileNames = append(fileNames, fileName)
	}

	return fileNames, nil
}

func exerciseAction(action string, id uint) any {
	switch action {
	case "Del":
		return template.HTML(`<button hx-delete="/exercise/` + strconv.FormatUint(uint64(id), 10) + `" hx-confirm="Delete exercise?">Del</button>`)
	case "Edit":
		return template.HTML(`<button hx-get="/exercise/` + strconv.FormatUint(uint64(id), 10) + `">Edit</button>`)
	default:
		return ""
	}
}
