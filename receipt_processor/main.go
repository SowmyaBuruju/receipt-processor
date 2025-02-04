package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Total        string `json:"total"`
	Items        []Item `json:"items"`
}

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type ReceiptResponse struct {
	ID string `json:"id"`
}

type PointsResponse struct {
	Points int `json:"points"`
}

var (
	receiptStore = make(map[string]int)
	mutex        = &sync.Mutex{}
)

func processReceipt(w http.ResponseWriter, r *http.Request) {
	var receipt Receipt
	if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	points := calculatePoints(receipt)

	mutex.Lock()
	receiptStore[id] = points
	mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ReceiptResponse{ID: id})
}

func getPoints(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/receipts/")
	id = strings.TrimSuffix(id, "/points")

	mutex.Lock()
	points, exists := receiptStore[id]
	mutex.Unlock()

	if !exists {
		http.Error(w, "No receipt found for that ID", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PointsResponse{Points: points})
}

func calculatePoints(receipt Receipt) int {
	points := 0
	alphanumericRegex := regexp.MustCompile("[a-zA-Z0-9]")
	points += len(alphanumericRegex.FindAllString(receipt.Retailer, -1))

	total, _ := strconv.ParseFloat(receipt.Total, 64)
	if total == float64(int(total)) {
		points += 50
	}
	if math.Mod(total, 0.25) == 0 {
		points += 25
	}

	points += (len(receipt.Items) / 2) * 5

	for _, item := range receipt.Items {
		trimmedDesc := strings.TrimSpace(item.ShortDescription)
		if len(trimmedDesc)%3 == 0 {
			price, _ := strconv.ParseFloat(item.Price, 64)
			points += int(math.Ceil(price * 0.2))
		}
	}

	dateParts := strings.Split(receipt.PurchaseDate, "-")
	day, _ := strconv.Atoi(dateParts[2])
	if day%2 == 1 {
		points += 6
	}

	timeParts := strings.Split(receipt.PurchaseTime, ":")
	hour, _ := strconv.Atoi(timeParts[0])
	if hour >= 14 && hour < 16 {
		points += 10
	}

	return points
}
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Receipt Processor API is running. Use /receipts/process to submit a receipt."))
}

func main() {
	http.HandleFunc("/", homeHandler) // Ensure this is added
	http.HandleFunc("/receipts/process", processReceipt)
	http.HandleFunc("/receipts/", getPoints)

	fmt.Println("Starting server on :8080...")
	http.ListenAndServe(":8080", nil)
}
