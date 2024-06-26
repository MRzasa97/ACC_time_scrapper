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
	lastSentTime   string
	lastTimeMs     int32
	lastTimesSlice []int32
	lastFuelLoad   float32
	pitStop        lastPitStop
	mutex          sync.Mutex
)

type lastPitStop struct {
	wasInPit      int
	completedLaps int
}

type smData struct {
	CarModel  string `json:"car_model"`
	BestTime  string `json:"best_time"`
	TrackName string `json:"track_name"`
	UserName  string `json:"user_name"`
}

func RunProcess(stop chan os.Signal, token string) {
	for {
		select {
		case <-stop:
			fmt.Println("Received stop signal. Exiting...")
			return
		default:
			time.Sleep(30 * time.Second)
			fmt.Println("Running API process...")
			bestTime, err := app.GetLastTime()
			if err != nil {
				fmt.Printf("failed to read shared memory: %v", err)
				fmt.Print("Waiting for data...")
				time.Sleep(5 * time.Second)
				continue
			}
			fmt.Printf("Packed %d", bestTime)

			// 	carModel, err := app.GetCarName()
			// 	if err != nil {
			// 		fmt.Printf("failed to read shared memory: %v", err)
			// 		fmt.Print("Waiting for data...")
			// 		time.Sleep(5 * time.Second)
			// 		continue
			// 	}

			// 	trackName, err := app.GetTrackName()
			// 	if err != nil {
			// 		fmt.Printf("failed to read shared memory: %v", err)
			// 		fmt.Print("Waiting for data...")
			// 		time.Sleep(5 * time.Second)
			// 		continue
			// 	}

			// 	mutex.Lock()
			// 	if *bestTime != lastSentTime {
			// 		lastSentTime = *bestTime
			// 		mutex.Unlock()
			// 		var data = smData{BestTime: *bestTime, CarModel: *carModel, TrackName: *trackName}

			// 		err = sendDataToAPI(data, token)
			// 		if err != nil {
			// 			fmt.Printf("Error sending data to API: %s", err)
			// 			fmt.Print("Waiting for data...")
			// 			time.Sleep(5 * time.Second)
			// 			continue
			// 		}
			// 	} else {
			// 		mutex.Unlock()
			// 	}
		}
	}
}

func FuelProcess(stop chan os.Signal) {
	for {
		select {
		case <-stop:
			fmt.Println("Received stop signal. Exiting...")
			return
		default:
			time.Sleep(5 * time.Second)
			currentTimeMs, err := app.GetLastTime()
			if err != nil {
				fmt.Printf("failed to read shared memory: %v", err)
				fmt.Print("Waiting for data...")
				time.Sleep(5 * time.Second)
				continue
			}
			fmt.Printf("Packet %d", currentTimeMs)
			isPit, err := app.GetIsInPitLane()
			if err != nil {
				fmt.Printf("failed to read pit status from shared memory: %v\n", err)
				fmt.Println("Waiting for data...")
				time.Sleep(5 * time.Second)
				continue
			}

			completedLaps, err := app.GetCompletedLaps()
			if err != nil {
				fmt.Printf("failed to read pit status from shared memory: %v\n", err)
				fmt.Println("Waiting for data...")
				time.Sleep(5 * time.Second)
				continue
			}

			fuelLoapXLap, err := app.GetFuelXLap()
			if err != nil {
				fmt.Printf("failed to read pit status from shared memory: %v\n", err)
				fmt.Println("Waiting for data...")
				time.Sleep(5 * time.Second)
				continue
			}

			sessionTimeLeft, err := app.GetSessionTimeLeft()
			if err != nil {
				fmt.Printf("failed to read pit status from shared memory: %v\n", err)
				fmt.Println("Waiting for data...")
				time.Sleep(5 * time.Second)
				continue
			}

			if isPit > 0 {
				fmt.Println("Car is Pitting!")
				mutex.Lock()
				pitStop.wasInPit = isPit
				pitStop.completedLaps = completedLaps
				mutex.Unlock()
			}

			mutex.Lock()
			if currentTimeMs != lastTimeMs {
				lastTimeMs = currentTimeMs
				if fuelLoapXLap > 0 && completedLaps > 0 {
					lastFuelLoad = fuelLoapXLap
				}
				if isPit == 0 && pitStop.completedLaps != completedLaps {
					lastTimesSlice = append(lastTimesSlice, lastTimeMs)
					fmt.Println("Added time to map!")
				}
				mutex.Unlock()
				fmt.Printf("Current time: %d \n", currentTimeMs)

			} else {
				mutex.Unlock()
			}

			mutex.Lock()
			if len(lastTimesSlice) > 1 {
				fuelLoad := countFuelLoad(lastTimesSlice, lastFuelLoad, sessionTimeLeft)
				mutex.Unlock()
				fmt.Printf("estimated fuel: %f \n last fuel: %f, session left: %f", fuelLoad, lastFuelLoad, sessionTimeLeft)
			} else {
				mutex.Unlock()
			}
		}
	}
}

func countFuelLoad(lastTimesSlice []int32, fuelLoadXLap float32, sessionTimeLeft float32) float32 {
	var timeSum int32
	for i := 0; i < len(lastTimesSlice); i++ {
		timeSum += lastTimesSlice[i]
	}
	timeSum += 2 * lastTimesSlice[len(lastTimesSlice)-1]
	avgTime := float32(timeSum / (int32(len((lastTimesSlice)) + 2)))
	lapsLeft := float32(sessionTimeLeft) / avgTime
	estimatedFuel := lapsLeft * fuelLoadXLap
	return estimatedFuel
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
