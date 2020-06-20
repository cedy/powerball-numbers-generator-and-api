package main

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetHash(t *testing.T) {
	var goodCases = []struct {
		input    combinationsData
		expected string
	}{
		{combinationsData{5, 4, 3, 2, 1, 1, 10, 2, 4, 6, 8, time.Now()}, "123451"},
		{combinationsData{65, 64, 63, 62, 61, 25, 10, 2, 4, 6, 8, time.Now()}, "616263646525"},
	}

	for _, test := range goodCases {
		result, _ := test.input.GetHash()
		if result != test.expected {
			t.Errorf("Hash of %v was incorrect, got: %s, want: %s.", test.input, result, test.expected)
		}
	}
	var badCases = []combinationsData{
		combinationsData{1, 1, 2, 2, 4, 1, 10, 2, 4, 6, 8, time.Now()},
		combinationsData{0, 69, 11, 62, 61, 25, 10, 2, 4, 6, 8, time.Now()},
		combinationsData{1, 79, 11, 62, 61, 25, 10, 2, 4, 6, 8, time.Now()},
		combinationsData{1, 2, 11, 62, 61, 29, 10, 2, 4, 6, 8, time.Now()},
		combinationsData{1, 2, 11, 62, 61, 0, 10, 2, 4, 6, 8, time.Now()},
		combinationsData{1, 2, 11, 62, 61, -2, 10, 2, 4, 6, 8, time.Now()},
	}

	for _, test := range badCases {
		result, err := test.GetHash()
		if err == nil {
			t.Errorf("Expected error, got %s", result)
		}
	}
}

func TestDateString(t *testing.T) {
	var goodCases = []struct {
		input    date
		expected string
	}{
		{date{"2020", "1", "01"}, "2020-01-01"},
		{date{"2020", "12", "31"}, "2020-12-31"},
		{date{"2020", "2", "29"}, "2020-02-29"},
		{date{"2000", "10", "1"}, "2000-10-01"},
		{date{"1999", "01", "12"}, "1999-01-12"},
	}

	for _, test := range goodCases {
		result, err := test.input.DateString()
		if result != test.expected {
			if err != nil {
				result = err.Error()
			}
			t.Errorf("Expected year:%s month:%s day:%s, got %s",
				test.input.Year, test.input.Month, test.input.Day, result)
		}
	}

	var badCases = []date{
		date{"19", "01", "12"},
		date{"0000", "12", "12"},
		date{"9999", "12", "12"},
		date{"20201", "12", "12"},
		date{"-2020", "12", "12"},
		date{"-202", "12", "12"},
		date{"2002", "31", "12"},
		date{"2020", "-12", "12"},
		date{"2020", "12", "-12"},
		date{"2020", "-2", "12"},
		date{"2020", "12", "-2"},
		date{"2020", "2", "31"},
		date{"2020", "0", "12"},
		date{"2020", "12", "0"},
		date{"0x00", "12", "0"},
		date{"AABB", "12", "0"},
		date{"2020", "BA", "31"},
		date{"2020", "2", "CA"},
		date{"2020", "\\", "''"},
	}

	for _, test := range badCases {
		result, err := test.DateString()
		if err == nil {
			t.Errorf("Expected error, got %s", result)
		}
	}
}

type mockClock struct {
	staticTime       time.Time
	sleepDurationSec time.Duration
}

func (c mockClock) Now() time.Time        { return c.staticTime }
func (c mockClock) Sleep(d time.Duration) { time.Sleep(c.sleepDurationSec * time.Second) }

func TestResetCounts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectExec("ALTER TABLE tale DROP COLUMN dayCount;")
	//clockMock := mockClock{time.Date(2020, time.January, 11, 3, 5, 0, 0, time.UTC), 1}
	exitChan := make(chan bool)
	time.AfterFunc(2*time.Second, func() { exitChan <- true })
	//resetCounts(db, clockMock, exitChan)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
