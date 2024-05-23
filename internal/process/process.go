package process

import (
	"fmt"
	"os"
	"time"

	"github.com/MRzasa97/ACC_time_scrapper/internal/app"
)

func RunProcess(stop chan os.Signal) {
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
}
