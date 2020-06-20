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

func (comb *combinationsData) GetHash() (string, error) {
	numbers := []int{comb.Digit1, comb.Digit2, comb.Digit3, comb.Digit4, comb.Digit5, comb.Pb}
	sort.Ints(numbers[:5])
	for i, number := range numbers {
		if number > 70 || number < 1 {
			return "", errors.New("Error: Number must be  within 1-69 range")
		}
		// Powerball range 1-26
		if i == 5 && (number < 1 || number > 26) {
			return "", errors.New("Error: Powerball number must be within 1-26 range")
		}
		if i < 5 {
			// check for duplicates
			if number == numbers[i+1] {
				return "", errors.New("Error: Powerball combination can cotain only unique numbers")
			}
		}
	}
	hash := fmt.Sprintf("%d%d%d%d%d%d", numbers[0], numbers[1], numbers[2], numbers[3], numbers[4], numbers[5])
	return hash, nil
}

type date struct {
	Year  string `uri:"year"`
	Month string `uri:"month"`
	Day   string `uri:"day"`
}

func isLeapYear(year int) bool {
	leapFlag := false
	if year%4 == 0 {
		if year%100 == 0 {
			if year%400 == 0 {
				leapFlag = true
			} else {
				leapFlag = false
			}
		} else {
			leapFlag = true
		}
	} else {
		leapFlag = false
	}
	return leapFlag
}

func (d *date) DateString() (string, error) {
	var date string
	var month int
	var monthToDaysLength = []struct {
		Month string
		Days  int
	}{
		{"January", 31},
		{"February", 28},
		{"March", 31},
		{"April", 30},
		{"May", 31},
		{"June", 30},
		{"July", 31},
		{"August", 31},
		{"September", 30},
		{"October", 31},
		{"November", 30},
		{"December", 31},
	}
	var currentYear = time.Now().UTC().Year()
	if isLeapYear(currentYear) {
		monthToDaysLength[1].Days = 29
	}
	year, err := strconv.Atoi(d.Year)
	// year should be 4 digits and within a range 1970-current year
	if err != nil || len(d.Year) != 4 || year > currentYear || year < 1970 {
		return "", errors.New("year is incorrect, please use for digit and be within current and 1970")
	}
	date = d.Year
	// month should be 2 digits and within a range 1-12
	if d.Month != "" || len(d.Month) <= 2 {
		month, err = strconv.Atoi(d.Month)
		if err != nil {
			return "", errors.New("month is incorrect, please provide month as number in rage 1-12")
		} else if month < 1 || month > 12 {
			return "", errors.New("month should be within 1-12 range")
		} else {
			date = fmt.Sprintf("%s-%02d", date, month)
		}
	} else {
		return date, nil
	}
	// day should be 2 digits and within a range 1-31
	if d.Day != "" || len(d.Day) <= 2 {
		day, err := strconv.Atoi(d.Day)
		if err != nil {
			return "", errors.New("day is incorrent, please provide a day as number in range 1-31")
		} else if day < 1 || day > monthToDaysLength[month-1].Days {
			// check for leap year
			return "", fmt.Errorf("day should be within 1-%d range in %s", monthToDaysLength[month-1].Days, monthToDaysLength[month-1].Month)
		} else {
			date = fmt.Sprintf("%s-%02d", date, day)
			return date, nil
		}
	} else {
		return date, nil
	}
}

func setupConnection(user string, password string, host string, port string, dbName string) *sql.DB {
	connectionParms := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, dbName)
	db, err := sql.Open("mysql", connectionParms)
	printAndExitIfError(err)
	db.SetMaxOpenConns(25)
	setAC, err := db.Query("SET autocommit = 1;")
	printAndExitIfError(err)
	setAC.Close()

	return db
}

func printAndExitIfError(err error) {
	if err != nil {
		apiLogger.Fatal(err)
	}
}

//in case of transaction error in resetCount, kick off new resetCounts in goroutine and panics with original error
func handleTransactionError(err error, db appDB, tx appTx, clock clock, exit <-chan bool) {
	if err != nil {
		tx.Rollback()
		apiLogger.Println(err.Error())
		clock.Sleep(10 * time.Second)
		go resetCounts(db, clock, exit)
		panic(err)
	}
}

type clock interface {
	Now() time.Time
	Sleep(d time.Duration)
}
type appClock struct{}

func (c appClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

func (c appClock) Now() time.Time {
	return time.Now()
}

type appTx interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Commit() error
	Rollback() error
}

type appDB interface {
	Begin() (*sql.Tx, error)
}

//resets counts daily, weekly, monthly and yearly respectivly
func resetCounts(db appDB, clock clock, exit <-chan bool) {
	defer func() {
		if r := recover(); r != nil {
			apiLogger.Println("ResetCount failed, ", r)
		}
	}()

	recentlyReset := false
	for {
		select {
		case <-exit:
			break
		default:
			currentTime := clock.Now()
			if currentTime.Hour() == 3 && !recentlyReset {
				tx, err := db.Begin()
				handleTransactionError(err, db, tx, clock, exit)
				query := fmt.Sprintf("ALTER TABLE tale DROP COLUMN dayCount;")
				_, err = tx.Exec(query)
				handleTransactionError(err, db, tx, clock, exit)
				query = fmt.Sprintf("ALTER TABLE tale ADD COLUMN dayCount INT UNSIGNED DEFAULT 0 AFTER count;")
				_, err = tx.Exec(query)
				handleTransactionError(err, db, tx, clock, exit)
				if currentTime.Weekday() == time.Monday {
					tx, err := db.Begin()
					handleTransactionError(err, db, tx, clock, exit)
					query := fmt.Sprintf("ALTER TABLE tale DROP COLUMN weekCount;")
					_, err = tx.Exec(query)
					handleTransactionError(err, db, tx, clock, exit)
					query = fmt.Sprintf("ALTER TABLE tale ADD COLUMN weekCount INT UNSIGNED DEFAULT 0 AFTER dayCount;")
					_, err = tx.Exec(query)
					handleTransactionError(err, db, tx, clock, exit)
				}
				if currentTime.Day() == 1 {
					tx, err := db.Begin()
					handleTransactionError(err, db, tx, clock, exit)
					query := fmt.Sprintf("ALTER TABLE tale DROP COLUMN monthCount;")
					_, err = tx.Exec(query)
					handleTransactionError(err, db, tx, clock, exit)
					query = fmt.Sprintf("ALTER TABLE tale ADD COLUMN monthCount INT UNSIGNED DEFAULT 0 AFTER weekCount;")
					_, err = tx.Exec(query)

					if currentTime.Month() == time.January {
						tx, err := db.Begin()
						handleTransactionError(err, db, tx, clock, exit)
						query := fmt.Sprintf("ALTER TABLE tale DROP COLUMN yearCount;")
						_, err = tx.Exec(query)
						handleTransactionError(err, db, tx, clock, exit)
						query = fmt.Sprintf("ALTER TABLE tale ADD COLUMN yearCount INT UNSIGNED DEFAULT 0 AFTER monthCount;")
						_, err = tx.Exec(query)
					}
				}
				err = tx.Commit()
				if err != nil {
					handleTransactionError(err, db, tx, clock, exit)
				}
				recentlyReset = true
			}
			clock.Sleep(3600 * time.Second)
			if currentTime.Hour() == 4 && recentlyReset {
				recentlyReset = false
			}
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
