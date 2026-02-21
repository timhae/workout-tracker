package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"mime/multipart"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/donseba/go-htmx"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type ExerciseFilter struct {
	Name            string      `form:"name"`
	Force           []Force     `form:"force" binding:"required"`
	Level           []Level     `form:"level" binding:"required"`
	Mechanic        []Mechanic  `form:"mechanic" binding:"required"`
	Category        []Category  `form:"category" binding:"required"`
	PrimaryMuscle   []Muscle    `form:"primary" binding:"required"`
	SecondaryMuscle []Muscle    `form:"secondary" binding:"required"`
	Equipment       []Equipment `form:"equipment" binding:"required"`
}

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
		log.Printf("db error: %v", err)
	}

	data := map[string]any{
		"Exercises": exercises,
		"Columns": []string{
			"Action", "Name", "Force", "Level", "Mechanic", "Category", "Primary", "Secondary", "Equipment", "Instructions", "Images",
		},
		"Actions":        []string{"Del", "Edit"},
		"PossibleValues": possibleValues,
	}
	table := htmx.NewComponent("templates/components/exercise_table.html").
		AddTemplateFunction("exerciseAction", exerciseAction).
		AddTemplateFunction("join", join)
	filter := htmx.NewComponent("templates/components/exercise_filter.html")
	page := htmx.NewComponent("templates/pages/exercises.html").
		With(table, "Table").
		With(filter, "Filter").
		SetData(data).
		Wrap(mainContent(), "Content")
	a.render(c, &page)
}

func (a *App) ListExercisesWithFilter(c *gin.Context) {
	var filter ExerciseFilter
	var exercises []Exercise
	if err := c.MustBindWith(&filter, binding.FormMultipart); err != nil {
		log.Printf("bind error: %v", err)
		return
	}
	exercises, err := gorm.G[Exercise](a.db).Order("id").
		Where("name LIKE ?", "%"+filter.Name+"%").
		Where(map[string]any{
			"force":          filter.Force,
			"level":          filter.Level,
			"mechanic":       filter.Mechanic,
			"category":       filter.Category,
			"primary_muscle": filter.PrimaryMuscle,
		}).
		Where(`NOT EXISTS (
			SELECT 1 FROM jsonb_array_elements(secondary_muscles) elem
			WHERE (elem::int) NOT IN (SELECT unnest(?::int[]))
		)`, pq.Array(filter.SecondaryMuscle)).
		Where(`NOT EXISTS (
			SELECT 1 FROM jsonb_array_elements(equipment) elem
			WHERE (elem::int) NOT IN (SELECT unnest(?::int[]))
		)`, pq.Array(filter.Equipment)).
		Find(c)
	if err != nil {
		log.Printf("db error: %v", err)
	}

	data := map[string]any{
		"Exercises": exercises,
		"Columns": []string{
			"Action", "Name", "Force", "Level", "Mechanic", "Category", "Primary", "Secondary", "Equipment", "Instructions", "Images",
		},
		"Actions": []string{"Del", "Edit"},
	}
	page := htmx.NewComponent("templates/components/exercise_table.html").
		SetData(data).
		AddTemplateFunction("exerciseAction", exerciseAction).
		AddTemplateFunction("join", join)
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
	log.Printf("id: %v+", id)
	exercise, err := gorm.G[Exercise](a.db).Where("id = ?", id).First(*a.ctx)
	if err != nil {
		log.Printf("db error: %v", err)
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
	id := c.Param("id")
	exercise, err := gorm.G[Exercise](a.db).Where("id = ?", id).First(*a.ctx)
	if err != nil {
		log.Printf("db error: %v", err)
	}
	a.deleteImages(exercise.Images)
	_, err = gorm.G[Exercise](a.db).Where("id = ?", id).Delete(*a.ctx)
	if err != nil {
		log.Printf("db error: %v", err)
	}
	a.ListExercises(c)
}

func (a *App) ValidateExercise(c *gin.Context) {
	var err error
	var exercise Exercise
	var button, validationLink string
	validationRequest := c.Request.Header.Get("X-Validation-Only") == "true"
	id := c.Param("id")

	if id == "" {
		button = "Create"
		validationLink = `hx-post="/exercise/validate"`
	} else {
		button = "Update"
		validationLink = `hx-post="/exercise/` + id + `/validate"`
	}

	err = c.ShouldBindWith(&exercise, binding.FormMultipart)

	switch {
	case err != nil:
		log.Printf("bind error: %v+", err)
	case validationRequest:
		err = errors.New("")
	case id == "":
		err = a.insertExercise(c, &exercise)
	case id != "":
		err = a.updateExercise(c, &exercise, id)
	}

	if err != nil {
		data := map[string]any{
			"PossibleValues": possibleValues,
			"ValidationLink": template.HTMLAttr(validationLink),
			"Input":          exercise,
			"Error":          err.Error(),
			"Button":         button,
		}
		page := htmx.NewComponent("templates/components/exercise_form.html").SetData(data).Wrap(mainContent(), "Content")
		a.render(c, &page)
		return
	}

	c.Header("HX-Location", `{"path":"/exercise/list", "target":"#content"}`)
}

func (a *App) updateExercise(c *gin.Context, exercise *Exercise, id string) error {
	var err error
	var fileNames []string

	dbExercise, err := gorm.G[Exercise](a.db).Where("id = ?", id).First(*a.ctx)
	if err != nil {
		log.Printf("db error: %v+", err)
		return err
	}

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	files := form.File["images"]

	switch {
	case len(files) > 0:
		a.deleteImages(dbExercise.Images) // has to be before saveImages
		fileNames, err = a.saveImages(exercise.Name, c, files)
		if err != nil {
			log.Printf("upload error: %v+", err)
			return err
		}
	case len(dbExercise.Images) > 0:
		fileNames = dbExercise.Images
	}
	exercise.Images = fileNames

	_, err = gorm.G[Exercise](a.db).Where("id = ?", id).Updates(*a.ctx, *exercise)
	if err != nil {
		log.Printf("db error: %v+", err)
		return err
	}

	return nil
}

func (a *App) insertExercise(c *gin.Context, exercise *Exercise) error {
	count, err := gorm.G[Exercise](a.db).Where("name = ?", exercise.Name).Count(*a.ctx, "name")
	if err != nil {
		log.Printf("db error: %v+", err)
		return err
	}
	if count > 0 {
		err = errors.New("exercise with name '" + exercise.Name + "' already exists")
		log.Printf("duplication error: %v+", err)
		return err
	}

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	files := form.File["images"]
	fileNames, err := a.saveImages(exercise.Name, c, files)
	if err != nil {
		log.Printf("upload error: %v+", err)
		return err
	}
	exercise.Images = fileNames

	err = gorm.G[Exercise](a.db).Create(*a.ctx, exercise)
	if err != nil {
		log.Printf("db error: %v+", err)
		return err
	}

	return nil
}

func (a *App) saveImages(name string, c *gin.Context, files []*multipart.FileHeader) ([]string, error) {
	var saver func(*multipart.FileHeader, string, ...fs.FileMode) error
	fileNames := []string{}
	if a.mockFS != nil {
		saver = a.mockFS.SaveUploadedFile
	} else {
		saver = c.SaveUploadedFile
	}

	for idx, file := range files {
		fileName := name + "_" + strconv.Itoa(idx)
		log.Printf("saving file %s as ./static/images/%s", file.Filename, fileName)
		err := saver(file, "./static/images/"+fileName)
		if err != nil {
			return nil, err
		}
		fileNames = append(fileNames, fileName)
	}

	return fileNames, nil
}

func (a *App) deleteImages(files []string) {
	var remover func(string) error
	if a.mockFS != nil {
		remover = a.mockRM.Remove
	} else {
		remover = os.Remove
	}

	for _, file := range files {
		log.Printf("removing file ./static/images/%s", file)
		err := remover("./static/images/" + file)
		if err != nil { // ignore error
			log.Printf("remove file error: %v+", err)
		}
	}
}

func join(values any, sep string) string {
	v := reflect.ValueOf(values)
	if v.Kind() != reflect.Slice {
		return fmt.Sprint(values)
	}

	var b strings.Builder
	for i := range v.Len() {
		elem := v.Index(i).Interface()
		s, _ := elem.(fmt.Stringer)
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(s.String())
	}
	return b.String()
}

func exerciseAction(action string, id uint) any {
	switch action {
	case "Del":
		return template.HTML(`<button hx-delete="/exercise/` + strconv.FormatUint(uint64(id), 10) + `" hx-confirm="Delete exercise?">Del</button>`)
	case "Edit":
		return template.HTML(`<button hx-get="/exercise/` + strconv.FormatUint(uint64(id), 10) + `" hx-push-url="/exercise/` + strconv.FormatUint(uint64(id), 10) + `">Edit</button>`)
	case "Add":
		return template.HTML(`<button hx-get="/exercise/` + strconv.FormatUint(uint64(id), 10) + `" hx-push-url="/exercise/` + strconv.FormatUint(uint64(id), 10) + `">Add</button>`)
	default:
		return ""
	}
}
