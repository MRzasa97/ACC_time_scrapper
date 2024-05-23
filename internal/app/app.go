package app

import (
	"bytes"
	"encoding/binary"
	"fmt"
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

type PageFile interface {
	SPageFileGraphic | SPageFileStatic
}

type BestTime struct {
	Minutes      int32
	Seconds      int32
	Milliseconds int32
}

func readSharedMemory[T PageFile]() (*T, error) {
	var pageFile T
	var sharedMemoryName string
	switch (any)(*new(T)).(type) {
	case SPageFileGraphic:
		sharedMemoryName = "Local\\acpmf_graphics"
	case SPageFileStatic:
		sharedMemoryName = "Local\\acpmf_static"
	default:
		return nil, fmt.Errorf("unsupported type")
	}

	buf := &bytes.Buffer{}
	staticSize := (int32)(unsafe.Sizeof(pageFile))

	mem, err := shm.Open(sharedMemoryName, staticSize)
	if err != nil {
		return nil, fmt.Errorf("failed to open shared memory %w", err)
	}
	defer mem.Close()

	data := make([]byte, staticSize)
	if _, err := mem.Read(data); err != nil {
		return nil, fmt.Errorf("failed to read shared memory: %w", err)
	}
	buf.Write(data)

	if err := binary.Read(buf, binary.LittleEndian, &pageFile); err != nil {
		return nil, fmt.Errorf("failed to decode shared memory: %w", err)
	}

	if isEmpty(pageFile) {
		return nil, fmt.Errorf("failed to read shared memory.")
	}

	return &pageFile, nil
}

func isEmpty[T PageFile](pageFile T) bool {
	var t T
	return pageFile == t
}

func GetBestTime() (*BestTime, error) {
	bestTime := &BestTime{}
	pageFileGraphics, err := readSharedMemory[SPageFileGraphic]()
	if err != nil {
		return nil, err
	}

	bestTime.Minutes = (pageFileGraphics.IBestTime / (1000 * 60)) % 60
	bestTime.Seconds = (pageFileGraphics.IBestTime / 1000) % 60
	bestTime.Milliseconds = pageFileGraphics.IBestTime % 1000

	return bestTime, err
}

func GetCarName() (*string, error) {
	var endString string
	pageFileStatic, err := readSharedMemory[SPageFileStatic]()
	if err != nil {
		return nil, err
	}
	for _, value := range pageFileStatic.CarModel {
		endString += string(rune(value))
	}
	return &endString, err
}
