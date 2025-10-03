package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type RouteParams struct {
	db *gorm.DB
}

func CreateDb() (*RouteParams, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Europe/Berlin",
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	params := &RouteParams{db: db}
	return params, nil
}

func (this *RouteParams) Fetch() []Exercise {
	var items []Exercise
	this.db.Order("Id DESC").Find(&items)

	return items
}
