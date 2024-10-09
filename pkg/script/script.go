package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Vehicle struct {
	VRM     string `json:"vrm"`
	Country string `json:"country"`
	Make    string `json:"make"`
}

type Driver struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Address1  string `json:"address_1"`
	Address2  string `json:"address_2"`
	PostCode  string `json:"post_code"`
	City      string `json:"city"`
	Region    string `json:"region"`
	Country   string `json:"country"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

type Data struct {
	ID               string    `json:"id"`
	LocationDateTime time.Time `json:"location_datetime"`
	Location         string    `json:"location"`
	TotalAmount      float64   `json:"total_amount"`
	Currency         string    `json:"currency"`
	Vehicle          Vehicle   `json:"vehicle"`
	Driver           Driver    `json:"driver"`
}

type Response struct {
	Status string `json:"status"`
}

func main() {
	data, err := getMockedData()
	if err {
		log.Printf("main - Error getting mocked data")
		return
	}

	start := time.Now()
	callApiPost(data)
	duration := time.Since(start)
	fmt.Printf("Single POST duration: %s\n", duration)

	start = time.Now()
	callApiGet(data.ID)
	duration = time.Since(start)
	fmt.Printf("Single GET duration: %s\n", duration)

	const numCalls = 1000
	const numGoroutines = 5
	var wg sync.WaitGroup

	successCount := 0
	errorCount := 0
	var mu sync.Mutex

	start = time.Now()

	makeAPICalls := func() {
		defer wg.Done()
		for i := 0; i < numCalls/numGoroutines; i++ {
			if err := callApiPost(data); err != nil {
				mu.Lock()
				errorCount++
				mu.Unlock()
			} else {
				mu.Lock()
				successCount++
				mu.Unlock()
			}

			if err := callApiGet(data.ID); err != nil {
				mu.Lock()
				errorCount++
				mu.Unlock()
			} else {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}
	}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go makeAPICalls()
	}

	wg.Wait()
	duration = time.Since(start)

	fmt.Printf("Total duration for concurrent calls: %s\n", duration)
	fmt.Printf("Successful executions: %d\n", successCount)
	fmt.Printf("Error executions: %d\n", errorCount)
}

func getMockedData() (Data, bool) {
	layout := "2006-01-02T15:04:05Z"
	locationDateTime, err := time.Parse(layout, "2024-10-20T08:38:34Z")
	if err != nil {
		log.Printf("getMockedData - Error parsing dates: %s\n", err.Error())
		return Data{}, true
	}
	data := Data{
		ID:               "8834HR43F9FNF3F8J98",
		LocationDateTime: locationDateTime,
		Location:         "North Highway",
		TotalAmount:      10.5,
		Currency:         "EUR",
		Vehicle: Vehicle{
			VRM:     "1234BCD",
			Country: "ES",
			Make:    "SEAT",
		},
		Driver: Driver{
			FirstName: "Jose",
			LastName:  "Garcia",
			Address1:  "Apple Street",
			Address2:  "",
			PostCode:  "1234",
			City:      "Madrid",
			Region:    "",
			Country:   "ES",
			Phone:     "111-222-333",
			Email:     "josegarcia@abc.es",
		},
	}
	return data, false
}

func callApiPost(data Data) error {
	url := "http://localhost:8080/transactions"

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling data: %s\n", err.Error())
		return err
	}

	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error making POST request: %s\n", err.Error())
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		log.Println("Data posted successfully")
		return nil
	}
	log.Printf("Failed to post data: %s\n", response.Status)
	return fmt.Errorf("failed to post data: %s", response.Status)
}

func callApiGet(id string) error {
	url := "http://localhost:8080/transactions?id=" + id

	response, err := http.Get(url)
	if err != nil {
		log.Printf("Error making GET request: %s\n", err.Error())
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		var data Data
		if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
			log.Printf("Error decoding response: %s\n", err.Error())
			return err
		}

		log.Printf("Data retrieved successfully: %+v\n", data)
		return nil
	}
	log.Printf("Failed to get data: %s\n", response.Status)
	return fmt.Errorf("failed to get data: %s", response.Status)
}
