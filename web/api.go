package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
)

var db *sql.DB
var configPath = flag.String("conf", "./conf.json", "Path to the configuration file")
var conf *Configuration
var apiLogger *log.Logger
var webLogger *log.Logger

func handlePrintError(e error) {
	if e != nil {
		log.Println(e.Error())
	}
}

func setupLoggers() {
	var apiLogFD, webLogFD, accesslogFD *os.File
	apiLogFD = os.Stdout
	webLogFD = os.Stdout
	accesslogFD = os.Stdout
	if conf.Production {
		var err error
		apiLogFD, err = os.OpenFile(conf.APIServerLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		handlePrintError(err)
		gin.DisableConsoleColor()
		accesslogFD, err = os.OpenFile(conf.WebServerAccessLogPath, os.O_CREATE|os.O_APPEND, 0644)
		handlePrintError(err)
		webLogFD, err = os.OpenFile(conf.WebServerLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		handlePrintError(err)
		gin.SetMode(gin.ReleaseMode)
	}

	apiLogger = log.New(apiLogFD, "", log.LstdFlags)
	apiLogger.Println("Logger initialized ", time.Now())

	webLogger = log.New(webLogFD, "", log.LstdFlags)
	webLogger.Println("Logger initialized ", time.Now())

	gin.DefaultWriter = io.MultiWriter(accesslogFD)
	gin.DefaultErrorWriter = io.MultiWriter(webLogger.Writer())
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
		apiLogger.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"err": "DB err. "})
		return
	}
	c.JSON(http.StatusOK, combinations)
}

func byCount(c *gin.Context) {
	count, err := strconv.Atoi(c.Param("count"))
	if err != nil {
		apiLogger.Println(count)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": "Bad argument"})
		return
	}
	condition := fmt.Sprintf("WHERE count = %d", count)
	combinations, err := getNumbers(condition, db)
	if err != nil {
		apiLogger.Println(err.Error())
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
		apiLogger.Println(err.Error())
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
		apiLogger.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, combinations)
}

func byHash(c *gin.Context) {
	var numbers combinationsData
	c.BindUri(&numbers)
	hash, err := numbers.GetHash()
	if err != nil {
		apiLogger.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	condition := fmt.Sprintf("WHERE hash = %s", hash)
	combinations, err := getNumbers(condition, db)
	if err != nil {
		apiLogger.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, combinations)
}

func main() {
	flag.Parse()
	conf = getConfiguration(*configPath)
	setupLoggers()
	runtime.GOMAXPROCS(conf.MaxCPUs)
	db = setupConnection(conf.DBuser, conf.DBpassword, conf.DBhost, conf.DBport, conf.DBname)
	numbersChan := make(chan []int, 100)
	broadcastChan := make(chan []int, 100)
	broadcastCommChan := make(chan chan string, 100)
	exitChan := make(chan bool)
	go resetCounts(db, appClock{}, exitChan)
	for i := 0; i < conf.RandomGeneratorsWorkers; i++ {
		go generateCombinations(numbersChan, broadcastChan)
		go writeCombinationsToDB(db, numbersChan)
		go writeCombinationsToDB(db, numbersChan)
	}
	go broadcastCombinations(broadcastChan, broadcastCommChan)
	router := mux.NewRouter()
	webLoggerWrapper := loggingMiddleware(webLogger)
	loggedRouter := webLoggerWrapper(router)
	api := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOriginFunc =
		func(r string) bool {
			return true
		}
	api.Use(cors.New(config))
	spa := spaHandler{staticPath: "static", indexPath: "index.html"}
	router.PathPrefix("/").Handler(spa)
	api.GET("/ws", serveWs(broadcastCommChan))
	api.GET("/history/page/:page", historyByPage)
	api.GET("/count/:count", byCount)
	api.GET("/top/count/:page", byTopCount)
	api.GET("/date/:year", byDate)
	api.GET("/date/:year/:month", byDate)
	api.GET("/date/:year/:month/:day", byDate)
	api.GET("/numbers/:Digit1/:Digit2/:Digit3/:Digit4/:Digit5/:Pb", byHash)
	apiSrv := &http.Server{
		Handler:      api,
		Addr:         ":2222",
		WriteTimeout: 14 * time.Second,
		ReadTimeout:  14 * time.Second,
	}
	frontEndSrv := &http.Server{
		Handler:      loggedRouter,
		Addr:         fmt.Sprintf(":%s", conf.HTTPSport),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go apiSrv.ListenAndServeTLS(conf.ServerCert, conf.ServerCertKey)
	go http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqhost := strings.Split(r.Host, ":")[0]
		apiLogger.Println("Redirecting")
		http.Redirect(w, r, "https://"+reqhost+"/history", http.StatusMovedPermanently)
	}))
	apiLogger.Fatal(frontEndSrv.ListenAndServeTLS(conf.ServerCert, conf.ServerCertKey))
}
