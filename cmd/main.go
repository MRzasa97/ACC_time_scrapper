package main

import (
	"fmt"
	"log"

	"github.com/MRzasa97/ACC_time_scrapper/internal/app"
)

func main() {
	bestTime, err := app.GetBestTime()
	if err != nil {
		log.Fatalf("Faile to read shared memory: %v", err)
	}

	carModel, err := app.GetCarName()
	if err != nil {
		log.Fatalf("failed to read shared memory: %v", err)
	}

	fmt.Printf("Best Time: %d:%d:%d ms\n", bestTime.Minutes, bestTime.Seconds, bestTime.Milliseconds)
	fmt.Printf("%s", *carModel)
}
