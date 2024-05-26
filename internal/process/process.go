package process

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/MRzasa97/ACC_time_scrapper/internal/app"
)

var (
	lastSentTime string
	mutex        sync.Mutex
)

type smData struct {
	CarModel  string `json:"car_model"`
	BestTime  string `json:"best_time"`
	TrackName string `json:"track_name"`
}

func RunProcess(stop chan os.Signal, token string) {
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

			trackName, err := app.GetTrackName()
			if err != nil {
				fmt.Printf("failed to read shared memory: %v", err)
				fmt.Print("Waiting for data...")
				time.Sleep(5 * time.Second)
				continue
			}

			mutex.Lock()
			if *bestTime != lastSentTime {
				lastSentTime = *bestTime
				mutex.Unlock()
				var data = smData{BestTime: *bestTime, CarModel: *carModel, TrackName: *trackName}

				err = sendDataToAPI(data, token)
				if err != nil {
					fmt.Printf("Error sending data to API: %s", err)
					fmt.Print("Waiting for data...")
					time.Sleep(5 * time.Second)
					continue
				}
			} else {
				mutex.Unlock()
			}
		}
	}
}

func sendDataToAPI(data smData, token string) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed converting data to json")
	}
	req, err := http.NewRequest("POST", "http://localhost:8000/acc/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprint("Bearer %s", token))

	cookie := &http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	}
	req.AddCookie(cookie)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %s", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("record was not created. status code: %d", response.StatusCode)
	}

	return nil
}
