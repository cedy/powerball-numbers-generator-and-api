package main

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type combinationsData struct {
	Digit1     int       `db:"digit1" json:"digit1"`
	Digit2     int       `db:"digit2" json:"digit2"`
	Digit3     int       `db:"digit3" json:"digit3"`
	Digit4     int       `db:"digit4" json:"digit4"`
	Digit5     int       `db:"digit5" json:"digit5"`
	Pb         int       `db:"pb" json:"powerball"`
	Count      int       `db:"count" json:"count,omitempty"`
	DayCount   int       `db:"dayCount" json:"dayCount"`
	WeekCount  int       `db:"weekCount" json:"weekCount"`
	MonthCount int       `db:"monthCount" json:"monthCount"`
	YearCount  int       `db:"yearCount" json:"yearCount"`
	Time       time.Time `db:"time" json:"time"`
}

func (comb *combinationsData) numbersFormatted() *map[string]string {
	numbers := fmt.Sprintf("%d %d %d %d %d %d", comb.Digit1, comb.Digit2, comb.Digit3, comb.Digit4, comb.Digit5, comb.Pb)
	time := comb.Time.Format("Mon Jan 2 2006")
	return &map[string]string{
		"count":      strconv.Itoa(comb.Count),
		"dayCount":   strconv.Itoa(comb.DayCount),
		"weekCount":  strconv.Itoa(comb.WeekCount),
		"monthCount": strconv.Itoa(comb.MonthCount),
		"yearCount":  strconv.Itoa(comb.YearCount),
		"date":       time,
		"numbers":    numbers,
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
		return "", errors.New("year is incorrect, please use for digit and be within current and 1970")
	}
	date = d.Year
	// month should be 2 digits and within a range 1-12
	if d.Month != "" || len(d.Month) == 2 {
		month, err := strconv.Atoi(d.Month)
		if err != nil {
			return "", errors.New("month is incorrect, please provide month as 2 digit")
		} else if month < 1 || month > 12 {
			return "", errors.New("month should be within 01-12 range")
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
			return "", errors.New("day is incorrent, please provide day as 2 digits")
		} else if day < 1 || day > 31 {
			return "", errors.New("day should be within 01-31 range")
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

//in case of transaction error in resetCount, kick off new resetCounts in goroutine and panics with original error
func handleTransactionError(err error, db *sql.DB, tx *sql.Tx) {
	if err != nil {
		tx.Rollback()
		fmt.Println(err.Error())
		time.Sleep(10 * time.Second)
		go resetCounts(db)
		panic(err)
	}
}

func resetCounts(db *sql.DB) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("ResetCount failed, ", r)
		}
	}()

	recentlyReset := false
	for {
		currentTime := time.Now()
		if currentTime.Hour() == 3 && !recentlyReset {
			tx, err := db.Begin()
			handleTransactionError(err, db, tx)
			query := fmt.Sprintf("ALTER TABLE tale DROP COLUMN dayCount;")
			_, err = tx.Exec(query)
			handleTransactionError(err, db, tx)
			query = fmt.Sprintf("ALTER TABLE tale ADD COLUMN dayCount INT UNSIGNED DEFAULT 0 AFTER count;")
			_, err = tx.Exec(query)
			handleTransactionError(err, db, tx)
			if currentTime.Weekday() == time.Monday {
				tx, err := db.Begin()
				handleTransactionError(err, db, tx)
				query := fmt.Sprintf("ALTER TABLE tale DROP COLUMN weekCount;")
				_, err = tx.Exec(query)
				handleTransactionError(err, db, tx)
				query = fmt.Sprintf("ALTER TABLE tale ADD COLUMN weekCount INT UNSIGNED DEFAULT 0 AFTER dayCount;")
				_, err = tx.Exec(query)
				handleTransactionError(err, db, tx)
			}
			if currentTime.Day() == 1 {
				tx, err := db.Begin()
				handleTransactionError(err, db, tx)
				query := fmt.Sprintf("ALTER TABLE tale DROP COLUMN monthCount;")
				_, err = tx.Exec(query)
				handleTransactionError(err, db, tx)
				query = fmt.Sprintf("ALTER TABLE tale ADD COLUMN monthCount INT UNSIGNED DEFAULT 0 AFTER weekCount;")
				_, err = tx.Exec(query)

				if currentTime.Month() == time.January {
					tx, err := db.Begin()
					handleTransactionError(err, db, tx)
					query := fmt.Sprintf("ALTER TABLE tale DROP COLUMN yearCount;")
					_, err = tx.Exec(query)
					handleTransactionError(err, db, tx)
					query = fmt.Sprintf("ALTER TABLE tale ADD COLUMN yearCount INT UNSIGNED DEFAULT 0 AFTER monthCount;")
					_, err = tx.Exec(query)
				}
			}
			err = tx.Commit()
			if err != nil {
				go resetCounts(db)
			}
			recentlyReset = true
		}
		time.Sleep(60 * time.Second)
		if currentTime.Hour() == 4 && recentlyReset {
			recentlyReset = false
		}
	}
}

func getNumbers(condition string, db *sql.DB) (*[]*map[string]string, error) {
	query := fmt.Sprintf("SELECT digit1, digit2, digit3, digit4, digit5, pb, count, dayCount, weekCount, monthCount, yearCount, time from tale  %v", condition)
	results, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	combinations := make([]*map[string]string, 0)
	var result combinationsData
	for results.Next() {
		results.Scan(&result.Digit1, &result.Digit2, &result.Digit3, &result.Digit4, &result.Digit5, &result.Pb, &result.Count, &result.DayCount,
			&result.WeekCount, &result.MonthCount, &result.YearCount, &result.Time)
		combinations = append(combinations, result.numbersFormatted())
	}
	return &combinations, nil
}

func getNumbersHistory(condition string, db *sql.DB) (*[]*map[string]string, error) {
	query := fmt.Sprintf("SELECT digit1, digit2, digit3, digit4, digit5, pb, time from history %v", condition)
	results, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	combinations := make([]*map[string]string, 0)
	var result combinationsData
	for results.Next() {
		results.Scan(&result.Digit1, &result.Digit2, &result.Digit3, &result.Digit4, &result.Digit5, &result.Pb, &result.Time)
		combinations = append(combinations, result.numbersFormatted())
	}
	return &combinations, nil
}
