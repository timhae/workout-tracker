package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func CreateDb() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=Europe/Berlin",
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&Exercise{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (a *App) GetAllExercises() (*[]Exercise, error) {
	var items []Exercise
	items, err := gorm.G[Exercise](a.db).Find(*a.ctx)
	if err != nil {
		return nil, err
	}

	return &items, nil
}

// // Create
// err = gorm.G[Product](db).Create(ctx, &Product{Code: "D42", Price: 100})

// // Read
// product, err := gorm.G[Product](db).Where("id = ?", 1).First(ctx) // find product with integer primary key
// products, err := gorm.G[Product](db).Where("code = ?", "D42").Find(ctx) // find product with code D42

// // Update - update product's price to 200
// err = gorm.G[Product](db).Where("id = ?", product.ID).Update(ctx, "Price", 200)
// // Update - update multiple fields
// err = gorm.G[Product](db).Where("id = ?", product.ID).Updates(ctx, map[string]interface{}{"Price": 200, "Code": "F42"})

// // Delete - delete product
// err = gorm.G[Product](db).Where("id = ?", product.ID).Delete(ctx)
