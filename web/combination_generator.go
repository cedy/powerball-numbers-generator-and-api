package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"sort"
	"time"
)

func generateCombinations(numbersChan chan<- []int, broadcastChan chan<- []int, stop <-chan int) {
	var balls [69]int
	numbers := make([]int, 6)
	select {
	case _ = <-stop:
		return
	default:
		for {
			for i := 0; i < 69; i++ {
				balls[i] = int(i + 1)
			}
			random := rand.New(rand.NewSource(time.Now().UnixNano() + 22))
			for i := 0; i < 5; i++ {
				isBallAvailable := false
				for isBallAvailable != true {
					ballIndex := random.Intn(69)
					ball := balls[ballIndex]
					if ball != 0 {
						numbers[i] = ball
						balls[ballIndex] = 0
						isBallAvailable = true
					}
				}

			}
			sort.Ints(numbers[:5])

			numbers[5] = int(random.Int31n(26)) + 1
			numbersChan <- numbers
			broadcastChan <- numbers
			time.Sleep(time.Millisecond * 500)
		}
	}
}

func writeCombinationsToDB(db *sql.DB, numbersChan chan []int) {
	for {
		numbers := <-numbersChan
		if len(numbers) == 0 {
			continue
		}
		query := fmt.Sprintf(`INSERT INTO tale (hash, digit1, digit2, digit3, digit4, digit5, pb, count, dayCount, weekCount, monthCount, yearCount) VALUES 
							 (%d%d%d%d%d%d, %d, %d, %d, %d, %d, %d, 1, 1, 1, 1, 1) 
							 ON DUPLICATE KEY UPDATE 
							 count = count + 1,
							 dayCount = dayCount + 1,
							 weekCount = weekCount + 1,
							 monthCount = monthCount + 1,
							 yearCount = yearCount + 1;`,
			numbers[0], numbers[1], numbers[2], numbers[3], numbers[4], numbers[5],
			numbers[0], numbers[1], numbers[2], numbers[3], numbers[4], numbers[5])
		insert, err := db.Query(query)
		if err != nil {
			apiLogger.Println(err.Error())
			numbersChan <- numbers
			time.Sleep(5 * time.Second)
			continue
		}
		insert.Close()
	}
}

func broadcastCombinations(numbersChan <-chan []int, commChan <-chan chan string) {
	subscribersList := make([]chan string, 0, 100)
	for {
		select {
		case numbers := <-numbersChan:
			message := fmt.Sprintf("%d %d %d %d %d %d", numbers[0], numbers[1], numbers[2], numbers[3], numbers[4], numbers[5])
			// iterate over channels list, check which channel to close, send a message to open channels
			for _, ch := range subscribersList {
				ch <- message
			}
		case channel := <-commChan:
			// iterate over subscribers list, if chan is in the list, remove it, otherwise add it to the list
			isDeleted := false
			for index, ch := range subscribersList {
				apiLogger.Println(channel, ch)
				if channel == ch {
					subscribersList = append(subscribersList[0:index], subscribersList[index+1:len(subscribersList)]...)
					isDeleted = true
				}
			}
			if !isDeleted {
				isDeleted = false
				subscribersList = append(subscribersList, channel)
			}
		}
	}
}
