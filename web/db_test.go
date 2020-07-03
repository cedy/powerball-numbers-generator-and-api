package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
	staticTime time.Time
}

func (c mockClock) Now() time.Time        { return c.staticTime }
func (c mockClock) Sleep(d time.Duration) {}

func TestResetCounts(t *testing.T) {
	apiLogger = log.New(os.Stdout, t.Name()+": ", log.LstdFlags)
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer dbMock.Close()
	mock.ExpectBegin()
	mock.ExpectExec("ALTER TABLE tale DROP COLUMN dayCount;").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("ALTER TABLE tale ADD COLUMN dayCount INT UNSIGNED DEFAULT 0 AFTER count;").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("ALTER TABLE tale DROP COLUMN weekCount;").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("ALTER TABLE tale ADD COLUMN weekCount INT UNSIGNED DEFAULT 0 AFTER dayCount;").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("ALTER TABLE tale DROP COLUMN monthCount;").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("ALTER TABLE tale ADD COLUMN monthCount INT UNSIGNED DEFAULT 0 AFTER weekCount;").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("ALTER TABLE tale DROP COLUMN yearCount;").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("ALTER TABLE tale ADD COLUMN yearCount INT UNSIGNED DEFAULT 0 AFTER monthCount;").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	// This time should trigger all resets
	clockMock := &mockClock{time.Date(2018, time.January, 1, 3, 5, 0, 0, time.UTC)}
	exitChan := make(chan bool)
	time.AfterFunc(2*time.Millisecond, func() { exitChan <- true })
	resetCounts(dbMock, clockMock, exitChan)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestHandleTransactionError(t *testing.T) {
	apiLogger = log.New(ioutil.Discard, t.Name()+": ", log.LstdFlags)
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer dbMock.Close()
	mock.ExpectBegin()
	mock.ExpectExec("ALTER TABLE tale DROP COLUMN dayCount;").WillReturnError(fmt.Errorf("expected testing error"))
	mock.ExpectRollback()
	mock.ExpectBegin()
	mock.ExpectExec("ALTER TABLE tale DROP COLUMN dayCount;").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("ALTER TABLE tale ADD COLUMN dayCount INT UNSIGNED DEFAULT 0 AFTER count;").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	clockMock := &mockClock{time.Date(2020, time.January, 11, 3, 5, 0, 0, time.UTC)}
	exitChan := make(chan bool)
	time.AfterFunc(2*time.Millisecond, func() {
		exitChan <- true
		exitChan <- true
	})
	resetCounts(dbMock, clockMock, exitChan)
	<-exitChan
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations in resetCounts: %s", err)
	}
}
