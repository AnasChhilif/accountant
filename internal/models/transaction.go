package models

import (
	"time"
)


type Transaction struct {
	ID          int64     `json:"id"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Notes       string    `json:"notes,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}


type TransactionInput struct {
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Notes       string    `json:"notes,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
}


var ValidCategories = map[string]bool{
	"income":   true,
	"expense":  true,
	"transfer": true,
}