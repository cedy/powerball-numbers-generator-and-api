package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type pbCombination struct {
	numbers []string
	date    string
}

// row: 02/03/2010,17 22 36 37 52 24,2
func parseLine(line string) pbCombination {
	splittedLine := strings.Split(line, ",")
	return pbCombination{numbers: strings.Split(splittedLine[1], " "), date: splittedLine[0]}
}

func constructQuery(data pbCombination) *string {
	query := "INSERT IGNORE INTO history (hash, digit1, digit2, digit3, digit4, digit5, pb, time) " +
		"VALUES (%s, %s, %s, %s, %s, %s, %s, '%s')"
	hash := strings.Join(data.numbers, "")
	//01/31/2019 month/date/year
	date := strings.Split(data.date, "/")
	year := date[2]
	month := date[0]
	day := date[1]
	mysql_datetime := year + "-" + month + "-" + day + " 00:00:00"
	query = fmt.Sprintf(query, hash, data.numbers[0], data.numbers[1], data.numbers[2], data.numbers[3], data.numbers[4], data.numbers[5], mysql_datetime)
	println(query)
	return &query
}

func executeQuery(queryChan <-chan *string) {
	db, err := sql.Open("mysql", "testuser:password@tcp(127.0.0.1:33061)/mpl")
	db.SetMaxOpenConns(100)
	if err != nil {
		panic(err.Error())
	}
	setAC, err := db.Query("SET autocommit = 1;")
	setAC.Close()
	var query *string
	for query = range queryChan {
		insStm, err := db.Query(*query)

		if err != nil {
			panic(err.Error())
		}
		insStm.Close()
	}
	if query == nil {
		return
	}
}

func main() {
	file, err := os.Open("historical_data.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	queryChan := make(chan *string)
	scanner := bufio.NewScanner(file)

	go executeQuery(queryChan)
	for scanner.Scan() {
		line := scanner.Text()
		numbers := parseLine(line)
		queryChan <- constructQuery(numbers)
	}

	fmt.Println("Closing chan")
	close(queryChan)
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
