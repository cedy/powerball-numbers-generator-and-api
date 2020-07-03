package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var ApiEndpoint string

func init() {
	ApiEndpoint = "https://www.powerball.com/api/v1/numbers/powerball/recent?_format=json"
}

type parseCombination struct {
	Numbers    string `json:"field_winning_numbers"`
	Date       string `json:"field_draw_date"`
	Multiplier string `json:"field_multiplier"`
}

func (n *parseCombination) Get() error {
	res, err := http.Get(ApiEndpoint)
	if err != nil {
		return err
	}
	// We expect to get 3 recent powerball drawings
	// Read first one, since it is the most recent
	dec := json.NewDecoder(res.Body)
	dec.Token()
	err = dec.Decode(n)
	if err != nil {
		return err
	}
	return nil
}

func (n *parseCombination) Save(DB *sql.DB) error {
	if n.Numbers == "" || n.Date == "" {
		return fmt.Errorf("can't save empty combination, numbers and date should be filled")
	}
	query := "INSERT INTO history (hash, digit1, digit2, digit3, digit4, digit5, pb, time) " +
		"VALUES (%s, %s, %s, %s, %s, %s, %s, '%s')"
	nums := strings.Split(n.Numbers, ",")
	var hash string
	var trimmedNum string
	for indx, n := range nums {
		trimmedNum = strings.TrimPrefix(n, "0")
		nums[indx] = trimmedNum
		hash += trimmedNum
	}
	query = fmt.Sprintf(query, hash, nums[0], nums[1], nums[2], nums[3], nums[4], nums[5], n.Date+" 00:00:00")
	q, err := DB.Query(query)
	if err != nil {
		return err
	}
	return q.Close()
}

func parseAndSavePB(DB *sql.DB, c clock) {
	pc := &parseCombination{}
	err := pc.Get()
	for i := 1; i < 10; i++ {
		if err == nil {
			break
		}
		c.Sleep(time.Duration(i*22) * time.Minute)
		err = pc.Get()
	}
	err = pc.Save(DB)
	for i := 1; i < 5; i++ {
		if err == nil {
			break
		}
		c.Sleep(time.Duration(i) * time.Minute)
		err = pc.Save(DB)
	}
	if err != nil {
		panic(err)
	}
}

// RunParseLatestPBCombination parses Powerball numbers and saves it into DB every drawing
func RunParseLatestPBCombination(DB *sql.DB, clock clock, logger *log.Logger) {
	// result should be available every Thursday and Sunday at 3 am
	// sleep until that time and then fetch and save winning numbers
	defer func() {
		if r := recover(); r != nil {
			logger.Printf("wasn't able to parse/save PB combinations: %s", r)
		}
	}()
	var sleepTime time.Duration
	for {
		now := clock.Now()
		// converts current time to hours, and looks for specific range
		// 0-3 - Sunday drawing coming, 3-99 - Thursday drawing is next, 99-168 - Sunday drawing coming
		nowInHours := int(now.Weekday())*24 + now.Hour()
		if nowInHours >= 3 && nowInHours < 99 {
			// next drawing on Thursday at 3am / at 75 hour
			sleepTime = time.Duration(99-nowInHours) * time.Hour
		} else {
			// next drawing on Sunday at 3am / at 3 hour
			if nowInHours < 3 {
				sleepTime = time.Duration(3-nowInHours) * time.Hour
			} else {
				sleepTime = time.Duration(168-nowInHours+3) * time.Hour
			}
		}
		logger.Printf("Current time %s, waiting %s before parsing", now, sleepTime)
		clock.Sleep(sleepTime)
		parseAndSavePB(DB, clock)
	}
}
