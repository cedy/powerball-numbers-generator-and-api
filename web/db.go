package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"sort"
	"strconv"
	"time"
)

type combinationsData struct {
	Digit1    int       `db:"digit1" json:"digit1"`
	Digit2    int       `db:"digit2" json:"digit2"`
	Digit3    int       `db:"digit3" json:"digit3"`
	Digit4    int       `db:"digit4" json:"digit4"`
	Digit5    int       `db:"digit5" json:"digit5"`
	Pb        int       `db:"pb" json:"powerball"`
	Count     int       `db:"count" json:"count,omitempty"`
	DayCount  int       `db:"dayCount" json:"dayCount"`
	WeekCount int       `db:"weekCount" json:"weekCount"`
	YearCount int       `db:"yearCount" json:"yearCount"`
	Time      time.Time `db:"time" json:"time"`
}

func (comb *combinationsData) numbersFormatted() *map[string]string {
	numbers := fmt.Sprintf("%d %d %d %d %d %d", comb.Digit1, comb.Digit2, comb.Digit3, comb.Digit4, comb.Digit5, comb.Pb)
	time := comb.Time.Format("Mon Jan 2 2006")
	return &map[string]string{
		"count":     strconv.Itoa(comb.Count),
		"dayCount":  strconv.Itoa(comb.DayCount),
		"weekCount": strconv.Itoa(comb.WeekCount),
		"yearCount": strconv.Itoa(comb.YearCount),
		"date":      time,
		"numbers":   numbers,
	}
}

func (comb *combinationsData) getHash() (string, error) {
	numbers := []int{comb.Digit1, comb.Digit2, comb.Digit3, comb.Digit4, comb.Digit5, comb.Pb}
	for i, number := range numbers {
		if number > 70 || number < 1 {
			return "", errors.New("Error: Number must be  within 1-69 range")
		}
		// Powerball range 1-26
		if i == 5 && number > 26 {
			return "", errors.New("Error: Powerball number must be within 1-26 range")
		}
	}
	sort.Ints(numbers[:5])
	hash := fmt.Sprintf("%d%d%d%d%d%d", numbers[0], numbers[1], numbers[2], numbers[3], numbers[4], numbers[5])
	return hash, nil
}

func isDigit(str string) bool {
	for char := range str {
		if char < 48 && char > 57 {
			return false
		}
	}
	return true
}

type date struct {
	Year  string `uri:"year"`
	Month string `uri:"month"`
	Day   string `uri:"day"`
}

func (d *date) DateString() (string, error) {
	var date string
	year, err := strconv.Atoi(d.Year)
	// year should be 4 digits and within a range 1970-current year
	if err != nil || len(d.Year) != 4 || year > time.Now().UTC().Year() || year < 1970 {
		return "", errors.New("Year is incorrect, please use for digit and be within current and 1970.")
	}
	date = d.Year
	// month should be 2 digits and within a range 1-12
	if d.Month != "" || len(d.Month) == 2 {
		month, err := strconv.Atoi(d.Month)
		if err != nil {
			return "", errors.New("Month is incorrect, please provide month as 2 digit.")
		} else if month < 1 || month > 12 {
			return "", errors.New("Month should be within 01-12 range.")
		} else {
			date = date + "-" + d.Month
		}
	} else {
		return date, nil
	}
	// day should be 2 digits and within a range 1-31
	if d.Day != "" || len(d.Day) == 2 {
		day, err := strconv.Atoi(d.Day)
		if err != nil {
			return "", errors.New("Day is incorrent, please provide day as 2 digits.")
		} else if day < 1 || day > 31 {
			return "", errors.New("Day should be within 01-31 range.")
		} else {
			date = date + "-" + d.Day
			return date, nil
		}
	} else {
		return date, nil
	}
}

func setupConnection() *sql.DB {
	db, err := sql.Open("mysql", "testuser:password@tcp(127.0.0.1:33061)/mpl?parseTime=true")
	db.SetMaxOpenConns(100)
	if err != nil {
		panic(err.Error())
	}
	setAC, err := db.Query("SET autocommit = 1;")
	setAC.Close()

	return db
}

func getNumbers(conditional string, db *sql.DB) (*[]*map[string]string, error) {
	query := fmt.Sprintf("SELECT digit1, digit2, digit3, digit4, digit5, pb, count, dayCount, weekCount, yearCount, time from tale  %v", conditional)
	results, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	combinations := make([]*map[string]string, 0, len(query))
	var result combinationsData
	for results.Next() {
		results.Scan(&result.Digit1, &result.Digit2, &result.Digit3, &result.Digit4, &result.Digit5, &result.Pb, &result.Count, &result.DayCount,
			&result.WeekCount, &result.YearCount, &result.Time)
		combinations = append(combinations, result.numbersFormatted())
	}
	return &combinations, nil
}
