package main

import (
	"html/template"
	"time"

	"github.com/donseba/go-htmx"
	"github.com/gin-gonic/gin"
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
	data := map[string]any{
		"PossibleValues": possibleValues,
		"ValidationLink": template.HTMLAttr(`hx-post="/plan/validate"`),
		"Input":          Exercise{},
		"Button":         "Create",
	}
	page := htmx.NewComponent("templates/components/plan_form.html").SetData(data).Wrap(mainContent(), "Content")
	a.render(c, &page)
}
