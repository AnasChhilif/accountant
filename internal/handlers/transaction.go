package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"accountant/internal/models"
	"accountant/internal/database"
)

type TransactionHandler struct {
    db database.Service
}

func NewTransactionHandler(db database.Service) *TransactionHandler {
    return &TransactionHandler{db: db}
}
// createTransactionHandler handles POST /api/transactions
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")


	var input models.TransactionInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	
	if input.Description == "" || input.Category == "" {
		http.Error(w, `{"error":"Description and category are required"}`, http.StatusBadRequest)
		return
	}

	
	if input.Date.IsZero() {
		input.Date = time.Now()
	}

	// Create transaction
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	transaction, err := h.db.CreateTransaction(ctx, input)
	if err != nil {
		if err.Error() == "invalid category" {
			http.Error(w, `{"error":"Invalid category"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error":"Error creating transaction"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}

// getTransactionHandler handles GET /api/transactions/{id}
func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid transaction ID"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	transaction, err := h.db.GetTransaction(ctx, id)
	if err != nil {
		if errors.Is(err, errors.New("transaction not found")) {
			http.Error(w, `{"error":"Transaction not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"Error retrieving transaction"}`, http.StatusInternalServerError)
		return
	}


	json.NewEncoder(w).Encode(transaction)
}

// updateTransactionHandler handles PUT /api/transactions/{id}
func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid transaction ID"}`, http.StatusBadRequest)
		return
	}

	// Parse request body
	var input models.TransactionInput
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if input.Description == "" || input.Category == "" {
		http.Error(w, `{"error":"Description and category are required"}`, http.StatusBadRequest)
		return
	}

	
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	transaction, err := h.db.UpdateTransaction(ctx, id, input)
	if err != nil {
		if err.Error() == "invalid category" {
			http.Error(w, `{"error":"Invalid category"}`, http.StatusBadRequest)
			return
		}
		if errors.Is(err, errors.New("transaction not found")) {
			http.Error(w, `{"error":"Transaction not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"Error updating transaction"}`, http.StatusInternalServerError)
		return
	}

	
	json.NewEncoder(w).Encode(transaction)
}

// deleteTransactionHandler handles DELETE /api/transactions/{id}
func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Invalid transaction ID"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = h.db.DeleteTransaction(ctx, id)
	if err != nil {
		if errors.Is(err, errors.New("transaction not found")) {
			http.Error(w, `{"error":"Transaction not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"Error deleting transaction"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// listTransactionsHandler handles GET /api/transactions
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	filters := make(map[string]string)
	
	
	if category := r.URL.Query().Get("category"); category != "" {
		filters["category"] = category
	}
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		filters["start_date"] = startDate
	}
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		filters["end_date"] = endDate
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		filters["limit"] = limit
	}

	// List transactions
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	transactions, err := h.db.ListTransactions(ctx, filters)
	if err != nil {
		http.Error(w, `{"error":"Error listing transactions"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(transactions)
}