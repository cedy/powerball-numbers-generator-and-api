package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"sort"
	"time"
)

func GenerateCombinations(numbersChan chan<- []int, stop <-chan int) {
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
		}
	}
}

func WriteCombinationsToDB(db *sql.DB, numbersChan <-chan []int) {
	for numbers := range numbersChan {
		if len(numbers) == 0 {
			return
		}
		query := fmt.Sprintf(`INSERT INTO tale (hash, digit1, digit2, digit3, digit4, digit5, pb, count) VALUES 
				 (%d%d%d%d%d%d, %d, %d, %d, %d, %d, %d, 1) ON DUPLICATE KEY UPDATE count = count + 1;`,
			numbers[0], numbers[1], numbers[2], numbers[3], numbers[4], numbers[5],
			numbers[0], numbers[1], numbers[2], numbers[3], numbers[4], numbers[5])
		insert, err := db.Query(query)
		if err != nil {
			fmt.Println(err.Error())
		}
		insert.Close()
	}
}
