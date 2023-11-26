package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var ds *DataStore

type Data struct {
	Timestamp time.Time
	Double    float64
	Float     float32
	Int32     int32
	Running   bool
	Speed     uint32
}

func getDataInTimeFrame(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	fromInt, err := strconv.ParseInt(c.Params.ByName("from"), 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	from := time.Unix(fromInt, 0).UTC()

	toInt, err := strconv.ParseInt(c.Params.ByName("to"), 10, 64)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	to := time.Unix(toInt, 0).UTC()
	log.Printf("getDataInTimeFrame %v - %v", from, to)

	entries := ds.Read(from, to)

	c.JSON(http.StatusOK, entries)
}

// Not using ctx. Does it leak somehow?
func RunRestServer(ctx context.Context, dataStore *DataStore) {
	ds = dataStore

	router := gin.Default()
	router.GET("/data/:from/:to", getDataInTimeFrame)

	log.Print("Starting REST API")
	router.Run("localhost:8080")
}
