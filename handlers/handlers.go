package handlers

import (
	"bitcoin-transaction/functions"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
)

func HomePageHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Pre-flight request handling
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Serve files in the current directory
	http.FileServer(http.Dir("static")).ServeHTTP(w, r)
}

func BalanceHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "db/transactions.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT amount FROM transactions WHERE spent = 0")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var btcBalance float64
	for rows.Next() {
		var amount float64
		err := rows.Scan(&amount)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		btcBalance += amount
	}

	eurBalance := functions.BtcToEur(btcBalance)
	// Round EUR_balance to two decimal places
	eurBalance = math.Round(eurBalance*100) / 100

	balance := map[string]string{"BTC_balance": strconv.FormatFloat(btcBalance, 'f', 5, 64), "EUR_balance": strconv.FormatFloat(eurBalance, 'f', 2, 64)} // display only until 2 decimal places

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

func ListTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	db, err := sql.Open("sqlite3", "db/transactions.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT transaction_id, amount, spent, created_at FROM transactions")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var transactions []functions.Transaction
	for rows.Next() {
		var tx functions.Transaction
		err := rows.Scan(&tx.TransactionID, &tx.Amount, &tx.Spent, &tx.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		transactions = append(transactions, tx)
	}

	// Encode transactions as JSON
	jsonBytes, err := json.Marshal(transactions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write JSON response
	_, err = w.Write(jsonBytes)
	if err != nil {
		log.Println("Error writing JSON response:", err)
	}
}

func TransferMoneyHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "db/transactions.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var data map[string]float64
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	amountBTC := data["amount"]

	// Check if the transfer amount is valid
	if amountBTC < 0.00001 {
		http.Error(w, "Transfer amount too small", http.StatusBadRequest)
		return
	}

	// Retrieve unspent transactions from the database
	unspentTransactions, err := functions.GetUnspentTransactions(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate total balance from unspent transactions
	totalBalance := functions.CalculateTotalBalance(unspentTransactions)

	// Check if there is enough balance to cover the transfer
	if totalBalance < amountBTC {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}

	// Mark used unspent transactions as spent
	leftoverAmount := functions.MarkTransactionsAsSpent(db, unspentTransactions, amountBTC)

	// If there is a leftover amount, create a new unspent transaction
	if leftoverAmount > 0 {
		functions.CreateNewUnspentTransaction(db, leftoverAmount)
	}

	fmt.Fprintf(w, "Transfer of %.2f BTC completed", amountBTC)
}
