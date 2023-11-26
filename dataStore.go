package main

import (
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const database = "sqlite.db"

type OpcData struct {
	Timestamp time.Time
	Double    float64
	Float     float32
	Int32     int32
	Running   bool
	Speed     uint32
}

type OpcDataModel struct {
	gorm.Model
	OpcData
}

type DataStore struct {
	Db *gorm.DB
}

func InitializeDataStore() DataStore {
	db, err := gorm.Open(sqlite.Open(database), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&OpcDataModel{})
	return DataStore{Db: db}
}

func (ds DataStore) Add(data *OpcData) {
	d := OpcDataModel{
		OpcData: *data,
	}
	ds.Db.Create(&d)
}

func (ds DataStore) Read(from time.Time, to time.Time) []OpcData {
	var entries []OpcDataModel
	//I'm sure there's a way to do this without having Timestamp as string in the query...
	ds.Db.Where("Timestamp BETWEEN ? AND ?", from, to).Find(&entries)
	result := []OpcData{}
	for _, entry := range entries {
		result = append(result, entry.OpcData)
	}
	return result
}
