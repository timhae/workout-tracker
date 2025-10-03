package main

import (
	"time"
)

type Exercise struct {
	ID              uint
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Name            string
	Force           Force
	Level           Level
	Mechanic        Mechanic
	Category        Category
	Equipment       []string
	PrimaryMuscles  []string
	SecondayMuscles []string
	Instructions    []string
	images          []string
}

type Force uint

const (
	Pull Force = iota
	Push
)

var forceName = map[Force]string{
	Pull: "Pull",
	Push: "Push",
}

func (f Force) String() string {
	return forceName[f]
}

type Level uint

const (
	Easy Level = iota
	Middle
	Hard
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
	Strength Category = iota
	Endurance
)

var categoryName = map[Category]string{
	Strength:  "Strength",
	Endurance: "Endurance",
}

func (c Category) String() string {
	return categoryName[c]
}

//   {
//     "name": "3/4 Sit-Up",
//     "force": "pull",
//     "level": "beginner",
//     "mechanic": "compound",
//     "equipment": "body only",
//     "primaryMuscles": [
//       "abdominals"
//     ],
//     "secondaryMuscles": [],
//     "instructions": [
//       "Lie down on the floor and secure your feet. Your legs should be bent at the knees.",
//       "Place your hands behind or to the side of your head. You will begin with your back on the ground. This will be your starting position.",
//       "Flex your hips and spine to raise your torso toward your knees.",
//       "At the top of the contraction your torso should be perpendicular to the ground. Reverse the motion, going only ¾ of the way down.",
//       "Repeat for the recommended amount of repetitions."
//     ],
//     "category": "strength",
//     "images": [
//       "3_4_Sit-Up/0.jpg",
//       "3_4_Sit-Up/1.jpg"
//     ],
//     "id": "3_4_Sit-Up"
//   },
// Muskelgruppe
//     Arme(92)
//     — Oberarme(78)
//     —— Bizeps(46)
//     —— Trizeps(32)
//     — Unterarme(14)
//     Bauch(98)
//     — gerade Bauchmuskeln(72)
//     —— obere Bauchmuskeln(52)
//     —— untere Bauchmuskeln(51)
//     — seitliche Bauchmuskeln(44)
//     Beine(207)
//     — Hüfte(149)
//     —— Abduktoren(15)
//     —— Po(136)
//     — Oberschenkel(152)
//     —— Adduktoren(22)
//     —— Beinbizeps(97)
//     —— Quadrizeps(96)
//     — Unterschenkel(35)
//     Brust(31)
//     — mittlere Brust(18)
//     — obere Brust(14)
//     — untere Brust(11)
//     Rücken(77)
//     — oberer Rücken(45)
//     —— Breiter Rückenmuskel(24)
//     —— Trapezmuskel(33)
//     — unterer Rücken(32)
//     Schulter(46)
//     — hintere Schultern(16)
//     — seitliche Schultern(22)
//     — vordere Schultern(27)

// Ausrüstung
//     Fitnessband
//     Hantelbank
//     Kabelturm
//     Kettlebell
//     Kurzhantel
//     Langhantel
//     Multipresse
//     ohne Ausrüstung
//     Sonstiges
//     SZ-Stange
//     Trainingsmaschine

// Übungstyp
//     Ausdauer
//     Beweglichkeit
//     Kraft

// Schwierigkeit
//     Leicht
//     Mittel
//     Schwer

// Trainingsziel
//     Abnehmen
//     Beweglichkeit
//     Haltung
//     Koordination
//     Kraftsteigerung
//     Muskelaufbau
//     Rehabilitation/Prävention
