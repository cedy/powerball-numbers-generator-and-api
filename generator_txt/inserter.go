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

func isDigit(char rune) bool {
	if char > 47 && char < 58 {
		return true
	}
	return false
}

func parseNumbers(line string) *[]string {
	digits := make([]string, 0, 7)
	number := make([]rune, 0, 2)
	for indx, chr := range line {
		if isDigit(chr) {
			number = append(number, chr)
		}
		// flush the number if char is not digit or we at the end of a line
		if (len(number) >= 1 && !isDigit(chr)) || indx+1 == len(line) {
			digits = append(digits, string(number))
			number = make([]rune, 0, 2)
		}
	}
	return &digits
}

func constructQuery(data *[]string) *string {
	query := "INSERT INTO tale (hash, digit1, digit2, digit3, digit4, digit5, pb, count) " +
		"VALUES (%s, %s, %s, %s, %s, %s, %s, %s) ON DUPLICATE KEY UPDATE " +
		"count = count + VALUES(count);"
	hash := strings.Join((*data)[:6], "")
	query = fmt.Sprintf(query, hash, (*data)[0], (*data)[1], (*data)[2], (*data)[3], (*data)[4], (*data)[5], (*data)[6])
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
	file, err := os.Open("100k.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	queryChan := make(chan *string)
	scanner := bufio.NewScanner(file)

	go executeQuery(queryChan)
	go executeQuery(queryChan)
	go executeQuery(queryChan)
	go executeQuery(queryChan)
	go executeQuery(queryChan)
	go executeQuery(queryChan)
	go executeQuery(queryChan)
	go executeQuery(queryChan)
	for scanner.Scan() {
		line := scanner.Text()
		numbers := parseNumbers(line)
		queryChan <- constructQuery(numbers)
	}

	fmt.Println("Closing chan")
	close(queryChan)
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
