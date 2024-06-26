package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/MRzasa97/ACC_time_scrapper/internal/app"
	"github.com/MRzasa97/ACC_time_scrapper/internal/process"
)

func isProcessRunningWindows(processName string) (bool, error) {
	cmd := exec.Command("tasklist")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false, err
	}
	output := out.String()
	return strings.Contains(output, processName), nil
}

func main() {
	token, err := app.Authenticate()
	if err != nil {
		log.Fatalf("Error Authentication! %s", err)
	}
	fmt.Printf("token: %s\n", token.Token)

	isAccRunning, err := isProcessRunningWindows("AC2-Win64-Shipping.exe")
	if err != nil {
		log.Fatalf("error checking process: %v\n", err)
	}
	if isAccRunning {
		stop := make(chan os.Signal, 1)

		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

		go process.FuelProcess(stop)
		// go process.RunProcess(stop, token.Token)
		<-stop
		fmt.Println("Application stopped")
	} else {
		log.Fatal("can't find assetto corsa process. Check if the game is running.")
	}
}
