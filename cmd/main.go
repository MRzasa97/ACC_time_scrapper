package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MRzasa97/ACC_time_scrapper/internal/app"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			select {
			case <-stop:
				fmt.Println("Received stop signal. Exiting...")
				return
			default:
				bestTime, err := app.GetBestTime()
				if err != nil {
					fmt.Printf("failed to read shared memory: %v", err)
					fmt.Print("Waiting for data...")
					time.Sleep(5 * time.Second)
					continue
				}

				carModel, err := app.GetCarName()
				if err != nil {
					fmt.Printf("failed to read shared memory: %v", err)
					fmt.Print("Waiting for data...")
					time.Sleep(5 * time.Second)
					continue
				}

				fmt.Printf("Best Time: %d:%d:%d ms\n", bestTime.Minutes, bestTime.Seconds, bestTime.Milliseconds)
				fmt.Printf("%s", *carModel)
				time.Sleep(5 * time.Second)
			}
		}
	}()
	<-stop
	fmt.Println("Application stopped")
	// bestTime, err := app.GetBestTime()
	// if err != nil {
	// 	log.Fatalf("failed to read best time: %v", err)
	// }

	// carModel, err := app.GetCarName()
	// if err != nil {
	// 	log.Fatalf("failed to read car model: %v", err)
	// }

	// fmt.Print("Waiting for data...")
	// fmt.Printf("Best Time: %d:%d:%d ms\n", bestTime.Minutes, bestTime.Seconds, bestTime.Milliseconds)
	// fmt.Printf("%s", *carModel)
	// time.Sleep(5 * time.Second)
}
