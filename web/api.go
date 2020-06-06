package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var db *sql.DB

func home(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func historyByPage(c *gin.Context) {
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page < 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": "Bad argument"})
		return
	}

	condition := fmt.Sprintf("ORDER BY time DESC LIMIT %d, 100", page*100)
	combinations, err := getNumbersHistory(condition, db)
	if err != nil {
		fmt.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": "DB err. "})
		return
	}
	c.JSON(http.StatusOK, combinations)
}

func byCount(c *gin.Context) {
	count, err := strconv.Atoi(c.Param("count"))
	if err != nil {
		fmt.Println(count)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": "Bad argument"})
		return
	}
	condition := fmt.Sprintf("WHERE count = %d", count)
	combinations, err := getNumbers(condition, db)
	if err != nil {
		fmt.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": "DB err. "})
		return
	}
	c.JSON(http.StatusOK, combinations)
}

func byTopCount(c *gin.Context) {
	// page has 100 records
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page < 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": "Bad argument"})
		return
	}
	// page 1 = Limit 0, 100; page 2 = Limit 100, 100 ...
	condition := fmt.Sprintf("ORDER by count DESC LIMIT %d, 100;", page*100)
	combinations, err := getNumbers(condition, db)
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
	combinations, err := getNumbers(condition, db)
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
	combinations, err := getNumbers(condition, db)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, combinations)
}

func main() {
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOriginFunc =
		func(r string) bool {
			// Allow only local connections
			ip := strings.Split(r[7:], ":")[0]
			if ip == "127.0.0.1" || ip == "localhost" {
				return true
			}
			return false
		}
	r.Use(cors.New(config))
	db = setupConnection()
	numbersChan := make(chan []int, 100)
	stopChan := make(chan int)
	broadcastChan := make(chan []int, 100)
	broadcastCommChan := make(chan chan string, 100)
	go resetCounts(db)
	go generateCombinations(numbersChan, broadcastChan, stopChan)
	go writeCombinationsToDB(db, numbersChan)
	go broadcastCombinations(broadcastChan, broadcastCommChan)
	r.LoadHTMLGlob("static/*.html")
	r.GET("/", home)
	r.GET("/ws", serveWs(broadcastCommChan))
	r.GET("/history/page/:page", historyByPage)
	r.GET("/count/:count", byCount)
	r.GET("/top/count/:page", byTopCount)
	r.GET("/date/:year", byDate)
	r.GET("/date/:year/:month", byDate)
	r.GET("/date/:year/:month/:day", byDate)
	r.GET("/numbers/:Digit1/:Digit2/:Digit3/:Digit4/:Digit5/:Pb", byHash)
	r.RunTLS(":8080", "localhost.crt", "localhost.key")
}
