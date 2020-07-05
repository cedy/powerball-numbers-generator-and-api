package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

type parseClock struct {
	staticTime time.Time
	sleepCalls []time.Duration
}

func (c parseClock) Now() time.Time { return c.staticTime }
func (c *parseClock) Sleep(d time.Duration) {
	c.sleepCalls = append(c.sleepCalls, d)
}
func (c *parseClock) SetStaticTime(t time.Time) {
	c.staticTime = t
}

func TestRunParseLatestsPBCombination(t *testing.T) {
	var testCases = []struct {
		currentTime   time.Time
		expectedSleep time.Duration
	}{
		{time.Date(2020, time.July, 1, 0, 0, 0, 0, time.UTC), time.Duration(27) * time.Hour},
		{time.Date(2020, time.July, 2, 2, 0, 0, 0, time.UTC), time.Hour},
		{time.Date(2020, time.July, 2, 3, 0, 0, 0, time.UTC), time.Duration(72) * time.Hour},
		{time.Date(2020, time.July, 4, 20, 0, 0, 0, time.UTC), time.Duration(7) * time.Hour},
		{time.Date(2020, time.July, 5, 1, 0, 0, 0, time.UTC), time.Duration(2) * time.Hour},
	}
	clockMock := &parseClock{time.Now(), make([]time.Duration, 0, 5)}
	for _, test := range testCases {
		clockMock.SetStaticTime(test.currentTime)
		runTestHelper(t, clockMock)
		if clockMock.sleepCalls[0] != test.expectedSleep {
			t.Errorf("Expected sleep %s, got %s", test.expectedSleep, clockMock.sleepCalls[0])
		}
		clockMock.sleepCalls = make([]time.Duration, 0, 5)
	}
}

func runTestHelper(t *testing.T, c clock) {
	dbMock, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer dbMock.Close()
	mock.ExpectQuery("INSERT INTO history (hash, digit1, digit2, digit3, digit4, digit5, pb, time) VALUES (123456, 1, 2, 3, 4, 5, 6, '2020-06-27 00:00:00')").WithArgs().WillReturnError(fmt.Errorf("Expected error"))
	mock.ExpectQuery("INSERT INTO history (hash, digit1, digit2, digit3, digit4, digit5, pb, time) VALUES (123456, 1, 2, 3, 4, 5, 6, '2020-06-27 00:00:00')").WithArgs().WillReturnError(fmt.Errorf("Expected error"))
	mock.ExpectQuery("INSERT INTO history (hash, digit1, digit2, digit3, digit4, digit5, pb, time) VALUES (123456, 1, 2, 3, 4, 5, 6, '2020-06-27 00:00:00')").WithArgs().WillReturnError(fmt.Errorf("Expected error"))
	mock.ExpectQuery("INSERT INTO history (hash, digit1, digit2, digit3, digit4, digit5, pb, time) VALUES (123456, 1, 2, 3, 4, 5, 6, '2020-06-27 00:00:00')").WithArgs().WillReturnError(fmt.Errorf("Expected error"))
	mock.ExpectQuery("INSERT INTO history (hash, digit1, digit2, digit3, digit4, digit5, pb, time) VALUES (123456, 1, 2, 3, 4, 5, 6, '2020-06-27 00:00:00')").WithArgs().WillReturnError(fmt.Errorf("Expected error"))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `[{
							"field_winning_numbers": "01,02,03,04,05,06",
							"field_multiplier": "2",
							"field_draw_date": "2020-06-27"}]`)
	}))
	defer ts.Close()
	ApiEndpoint = ts.URL
	// This time should trigger all resets
	RunParseLatestPBCombination(dbMock, c, log.New(ioutil.Discard, t.Name()+" :", log.LstdFlags))
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
