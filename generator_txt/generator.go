package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
	"unsafe"
)

func generateNumbers(statsChan chan<- map[[6]uint8]int, stopProcCounter chan<- uint8) {
	var balls [69]uint8
	var numbers [6]uint8
	stats := make(map[[6]uint8]int)
	for a := 0; true; a++ {

		for i := 0; i < 69; i++ {
			balls[i] = uint8(i + 1)
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

		for sorted := false; sorted != true; {
			sorted = true
			for i := 0; i < 4; i++ {
				if numbers[i] > numbers[i+1] {
					numbers[i], numbers[i+1] = numbers[i+1], numbers[i]
					sorted = false
				}
			}
		}
		numbers[5] = uint8(random.Int31n(26)) + 1
		stats[numbers] += 1
		if a%10000 == 0 {
			statsChan <- stats
			_ = stats
			stats = make(map[[6]uint8]int)
			_, err := os.Stat("stop")
			if !os.IsNotExist(err) {
				stopProcCounter <- 1
				fmt.Println("stopping goroutine")
				return
			}
		}
	}
}

func main() {
	collections := make(chan map[[6]uint8]int, 8)
	stats := make(map[[6]uint8]int)
	stopProcCounter := make(chan uint8, 9)

	go generateNumbers(collections, stopProcCounter)
	go generateNumbers(collections, stopProcCounter)
	go generateNumbers(collections, stopProcCounter)
	go generateNumbers(collections, stopProcCounter)
	go generateNumbers(collections, stopProcCounter)
	go generateNumbers(collections, stopProcCounter)
	go generateNumbers(collections, stopProcCounter)
	go generateNumbers(collections, stopProcCounter)
	for value := range collections {
		for key, count := range value {
			stats[key] += count
		}
		printSizeOfMap(&stats)
		fmt.Println(len(stats))
		if len(stopProcCounter) == 8 {
			stopProcCounter <- 1
			close(collections)
		}
	}
	max := 1
	var numbers [6]uint8
	f, _ := os.Create("dump.txt")
	defer f.Close()
	for key, value := range stats {
		fmt.Fprintln(f, key, value)
		if value > max {
			max = value
			numbers = key
		}
	}
	fmt.Println(max, numbers)

}

func printSizeOfMap(theMap *map[[6]uint8]int) {
	var key [6]uint8
	var value int
	fmt.Println(uint64(len(*theMap)*8) + uint64(len(*theMap)*8*int(unsafe.Sizeof(key))) + uint64(len(*theMap)*8*int(unsafe.Sizeof(value))))
}
