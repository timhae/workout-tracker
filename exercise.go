package main

import (
	"encoding/json"
	"time"
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
	Instructions     string      `form:"instructions" binding:"required"`
	Images           string
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
