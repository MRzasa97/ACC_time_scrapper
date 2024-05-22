package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"unsafe"

	"github.com/hidez8891/shm"
)

type SPageFileGraphic struct {
	PacketId                 int32
	Status                   int32
	SessionType              int32
	CurrentTime              [15]uint16
	LastTime                 [15]uint16
	BestTime                 [15]uint16
	Split                    [15]uint16
	CompletedLaps            int32
	Position                 int32
	ICurrentTime             int32
	ILastTime                int32
	IBestTime                int32
	SessionTimeLeft          float32
	DistanceTraveled         float32
	IsInPit                  int32
	CurrentSectorIndex       int32
	LastSectorTime           int32
	NumberOfLaps             int32
	TyreCompound             [33]uint16
	ReplayTimeMultiplier     float32
	NormalizedCarPosition    float32
	ActiveCars               int32
	CarCoordinates           [60][3]float32
	CarId                    [60]int32
	PlayerCarId              int32
	PenaltyTime              float32
	Flag                     int32
	PenaltyShortCut          int32
	IdealLineOn              int32
	IsInPitLane              int32
	SurfaceGrip              float32
	MandatoryPitDone         int32
	WindSpeed                float32
	WindDirection            float32
	IsSetupMenuVisible       int32
	MainDisplayIndex         int32
	SecondaryDisplayIndex    int32
	TC                       int32
	TCCut                    int32
	EngineMap                int32
	ABS                      int32
	FuelXLap                 int32
	RainLights               int32
	FlashingLights           int32
	LightStage               int32
	ExhaustTemperature       float32
	WiperLevel               int32
	DriverStintTotalTimeLeft int32
	DriverStintTimeLeft      int32
	RainTyres                int32
}

type SPageFileStatic struct {
	SMVersion                [15]uint16
	ACVersion                [15]uint16
	NumberOfSessions         int32
	NumCars                  int32
	CarModel                 [33]uint16
	Track                    [33]uint16
	PlayerName               [33]uint16
	PlayerSurName            [33]uint16
	PlayerNickname           [33]uint16
	SectorCount              int32
	MaxTorque                float32
	MaxPower                 float32
	MaxRPM                   int32
	MaxFuel                  float32
	MaxSuspensionTravel      [4]float32
	TyreRadius               float32
	MaxTurboBoost            float32
	Deprecated1              float32
	Deprecated2              float32
	PenaltiesEnabled         int32
	AidFuelRate              int32
	AidTireRate              int32
	AidMechanicalDamage      float32
	AidAllowTyreBlankets     int32
	AidStability             float32
	AidAutoClutch            int32
	AidAutoBlip              int32
	HasDRS                   int32
	HasERS                   int32
	HasKERS                  int32
	KERSMaxJ                 float32
	EngineBrakeSettingsCount int32
	ERSPowerControllerCount  int32
	TrackSplineLength        float32
	TrackConfiguration       [33]uint16
	ERSMaxJ                  float32
	IsTimedRace              int32
	HasExtraLap              int32
	CarSkin                  [33]uint16
	ReversedGridPosition     int32
	PitWindowStart           int32
	PitWindowEnd             int32
	IsOnline                 int32
}

type BestTime struct {
	minutes      int32
	seconds      int32
	milliseconds int32
}

func readSharedMemoryGraphics() (*SPageFileGraphic, error) {
	sharedMemoryName := "Local\\acpmf_graphics"
	buf := &bytes.Buffer{}
	pageFile := &SPageFileGraphic{}
	graphicsSize := (int32)(unsafe.Sizeof(*pageFile))
	mem, err := shm.Open(sharedMemoryName, graphicsSize)
	if err != nil {
		return nil, fmt.Errorf("Failed to open shared memory: %w", err)
	}

	data := make([]byte, graphicsSize)
	if _, err := mem.Read(data); err != nil {
		return nil, fmt.Errorf("Failed to read shared memory: %w", err)
	}
	buf.Write(data)

	if err := binary.Read(buf, binary.LittleEndian, pageFile); err != nil {
		return nil, fmt.Errorf("Failed to decode shared memory: %w", err)
	}

	return pageFile, nil
}

func readSharedMemoryStatic() (*SPageFileStatic, error) {
	sharedMemoryName := "Local\\acpmf_static"
	buf := &bytes.Buffer{}
	pageFile := &SPageFileStatic{}
	staticSize := (int32)(unsafe.Sizeof(*pageFile))
	mem, err := shm.Open(sharedMemoryName, staticSize)
	if err != nil {
		return nil, fmt.Errorf("Failed to open shared memory %w", err)
	}

	data := make([]byte, staticSize)
	if _, err := mem.Read(data); err != nil {
		return nil, fmt.Errorf("Failed to read shared memory: %w", err)
	}
	buf.Write(data)

	if err := binary.Read(buf, binary.LittleEndian, pageFile); err != nil {
		return nil, fmt.Errorf("Failed to decode shared memory: %w", err)
	}

	return pageFile, nil
}

func convertUint16ArrayToString(arr [33]uint16) string {
	// Find the end of the string (null terminator)
	var endString string
	for _, value := range arr {
		endString += string(rune(value))
	}
	return endString
}

func convertMillisecondsToTimeStruct(milliseconds int32) *BestTime {
	bestTime := &BestTime{}
	bestTime.minutes = (milliseconds / (1000 * 60)) % 60
	bestTime.seconds = (milliseconds / 1000) % 60
	bestTime.milliseconds = milliseconds % 1000

	return bestTime
}

func main() {
	pageFileGraphics, err := readSharedMemoryGraphics()
	if err != nil {
		log.Fatalf("Faile to read shared memory: %v", err)
	}

	pageFileStatics, err := readSharedMemoryStatic()
	if err != nil {
		log.Fatalf("failed to read shared memory: %v", err)
	}

	bestTime := convertMillisecondsToTimeStruct(pageFileGraphics.IBestTime)
	car := convertUint16ArrayToString(pageFileStatics.CarModel)
	fmt.Printf("Best Time: %d:%d:%d ms\n", bestTime.minutes, bestTime.seconds, bestTime.milliseconds)
	fmt.Printf("%s", car)
}
