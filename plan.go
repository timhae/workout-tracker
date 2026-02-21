package main

import (
	"log"
	"time"

	"github.com/donseba/go-htmx"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PlanFilter struct {
	Name string `form:"name"`
}

type Plan struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time
	Sets      []Set
}

type Set struct {
	Units []Unit
}

type Unit struct {
	Exercise Exercise
	Pause    time.Time
}

func (a *App) ListPlans(c *gin.Context) {
	data := map[string]any{
		"Text": "Welcome to the plans page",
	}
	page := htmx.NewComponent("templates/pages/plans.html").SetData(data).Wrap(mainContent(), "Content")
	a.render(c, &page)
}

func (a *App) CreatePlan(c *gin.Context) {
	var exercises []Exercise
	exercises, err := gorm.G[Exercise](a.db).Order("id").Find(*a.ctx)
	if err != nil {
		log.Printf("db error: %v", err)
	}

	data := map[string]any{
		"Input":     Plan{},
		"Exercises": exercises,
		"Columns": []string{
			"Action", "Name", "Force", "Level", "Mechanic", "Category", "Primary", "Secondary", "Equipment", "Instructions", "Images",
		},
		"Actions":        []string{"Add"},
		"PossibleValues": possibleValues,
	}
	plan := htmx.NewComponent("templates/components/plan_form.html")
	table := htmx.NewComponent("templates/components/exercise_table.html").
		AddTemplateFunction("exerciseAction", exerciseAction).
		AddTemplateFunction("join", join)
	filter := htmx.NewComponent("templates/components/exercise_filter.html")
	page := htmx.NewComponent("templates/pages/create_plan.html").
		With(plan, "Plan").
		With(table, "Table").
		With(filter, "Filter").
		SetData(data).
		Wrap(mainContent(), "Content")
	a.render(c, &page)
}
