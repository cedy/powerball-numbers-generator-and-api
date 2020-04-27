package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

var db *sql.DB

func home(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func historyLast(c *gin.Context) {
	numberOfLastCombinations, err := strconv.Atoi(c.Param("last"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": "Bad argument"})
	} else {
		if numberOfLastCombinations < 0 {
			numberOfLastCombinations = -numberOfLastCombinations
		}
		condition := fmt.Sprintf("LIMIT %d", numberOfLastCombinations)
		combinations, err := getNumbersBy(condition, db)
		if err != nil {
			fmt.Println(err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": "DB err. "})
			return
		}
		c.JSON(http.StatusOK, combinations)
	}
}

func byCount(c *gin.Context) {
	count, err := strconv.Atoi(c.Param("count"))
	if err != nil {
		fmt.Println(count)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": "Bad argument"})
		return
	}
	condition := fmt.Sprintf("WHERE count = %d", count)
	combinations, err := getNumbersBy(condition, db)
	if err != nil {
		fmt.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": "DB err. "})
		return
	}
	c.JSON(http.StatusOK, combinations)
}

func byDate(c *gin.Context) {
	var date date
	c.BindUri(&date)
	dateString, err := date.DateString()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	condition := fmt.Sprintf("WHERE time LIKE '%s%%'", dateString)
	combinations, err := getNumbersBy(condition, db)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, combinations)
}

func byHash(c *gin.Context) {
	var numbers combinationsData
	c.BindUri(&numbers)
	hash, err := numbers.getHash()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	condition := fmt.Sprintf("WHERE hash = %s", hash)
	combinations, err := getNumbersBy(condition, db)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, combinations)
}

func main() {
	r := gin.Default()
	db = setupConnection()
	numbersChan := make(chan []int, 100)
	stopChan := make(chan int)
	broadcastChan := make(chan []int, 100)
	broadcastCommChan := make(chan chan string, 100)
	go GenerateCombinations(numbersChan, broadcastChan, stopChan)
	go WriteCombinationsToDB(db, numbersChan)
	go BroadcastCombinations(broadcastChan, broadcastCommChan)
	r.LoadHTMLGlob("static/*.html")
	r.GET("/", home)
	r.GET("/ws", serveWs(broadcastCommChan))
	r.GET("/history/last/:last", historyLast)
	r.GET("/count/:count", byCount)
	r.GET("/date/:year", byDate)
	r.GET("/date/:year/:month", byDate)
	r.GET("/date/:year/:month/:day", byDate)
	r.GET("/numbers/:Digit1/:Digit2/:Digit3/:Digit4/:Digit5/:Pb", byHash)
	r.Run()
}
