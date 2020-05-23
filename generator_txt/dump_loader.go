package main

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"strings"
)

func parseLine(line string) []string {
	cleanedLine := strings.Replace(strings.Replace(line, "[", "", 1), "]", "", 1)
	return strings.Split(cleanedLine, " ")
}

func constructQuery(data []string) *string {
	query := "INSERT INTO tale (hash, digit1, digit2, digit3, digit4, digit5, pb, count) " +
		"VALUES (%s, %s, %s, %s, %s, %s, %s, %s) ON DUPLICATE KEY UPDATE count = VALUES(count) + count"
	hash := strings.Join(data[:6], "")
	query = fmt.Sprintf(query, hash, data[0], data[1], data[2], data[3], data[4], data[5], data[6])
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
	file, err := os.Open("dump.txt")
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
