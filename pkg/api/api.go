package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

var (
	transactions = make(map[string]Data)
	mu           sync.Mutex
)

func main() {
	http.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlePostTransaction(w, r)
		case http.MethodGet:
			handleGetTransaction(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	port := ":8080"
	fmt.Printf("Server listening in port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}

func handlePostTransaction(w http.ResponseWriter, r *http.Request) {
	var data Data

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		log.Printf("handlePostTransaction - Error reading body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(b, &data)
	if err != nil {
		log.Printf("handlePostTransaction - Error unmarshaling body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	transactions[data.ID] = data
	mu.Unlock()

	result := Response{Status: "Success"}

	response, err := json.Marshal(result)
	if err != nil {
		log.Printf("handlePostTransaction - Error marshaling result")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func handleGetTransaction(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	id := query.Get("id")
	if id == "" {
		err := errors.New("param id is missing")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, exists := transactions[id]

	if !exists {
		err := errors.New("id not found")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := json.Marshal(data)
	if err != nil {
		log.Printf("handlePostTransaction - Error marshaling result")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
