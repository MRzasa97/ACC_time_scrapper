package app

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func Authenticate() (LoginResponse, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter login: ")
	login, _ := reader.ReadString('\n')
	login = strings.TrimSpace(login)

	fmt.Print("Enter password: ")
	bytePassword, _ := term.ReadPassword(int(syscall.Stdin))
	password := string(bytePassword)
	password = strings.TrimSpace(password)

	fmt.Println()

	if login == "" || password == "" {
		log.Fatal("login and password cannot be empty")
	}
	fmt.Printf("Login: %s\n", login)
	fmt.Printf("Password: %s\n", password)
	credentials := map[string]string{
		"username": login,
		"password": password,
	}
	jsonData, err := json.Marshal(credentials)
	fmt.Printf("jsonData: %v\n", jsonData)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("error marshaling JSON: %s", err)
	}

	response, err := http.Post("http://localhost:8000/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return LoginResponse{}, fmt.Errorf("error sending request: %s", err)
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("error reading response body: %s", err)
	}

	if response.StatusCode != http.StatusOK {
		return LoginResponse{}, fmt.Errorf("login failed: %s", string(responseBody))
	}
	var result LoginResponse
	for _, cookie := range response.Cookies() {
		if cookie.Name == "token" {
			result.Token = cookie.Value
		}
	}

	return result, nil
}
