package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Transaction represents a Bitcoin transaction
type Transaction struct {
	TransactionID string    `json:"transaction_id"`
	Amount        float64   `json:"amount"`
	Spent         bool      `json:"spent"`
	CreatedAt     time.Time `json:"created_at"`
}

var transactions []Transaction
var db *sql.DB

func main() {
	db, err := sql.Open("sqlite3", "db/transactions.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create transactions table if it does not exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS transactions (
		transaction_id TEXT PRIMARY KEY,
		amount REAL,
		spent INTEGER,
		created_at TIMESTAMP
	);`)
	if err != nil {
		log.Fatal(err)
	}

	// handlers
	http.HandleFunc("/transactions", listTransactions)
	http.HandleFunc("/balance", getBalance)
	http.HandleFunc("/transfer", transferMoney)
	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func listTransactions(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT transaction_id, amount, spent, created_at FROM transactions")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var tx Transaction
		err := rows.Scan(&tx.TransactionID, &tx.Amount, &tx.Spent, &tx.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		transactions = append(transactions, tx)
	}

	json.NewEncoder(w).Encode(transactions)
}

func getBalance(w http.ResponseWriter, r *http.Request) {
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

	eurBalance := btcToEur(btcBalance)
	balance := map[string]float64{"BTC_balance": btcBalance, "EUR_balance": eurBalance}
	json.NewEncoder(w).Encode(balance)
}

func transferMoney(w http.ResponseWriter, r *http.Request) {
	var data map[string]float64
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	amountEUR := data["amount"]
	// Implement transfer logic here
	fmt.Fprintf(w, "Transfer of %.2f EUR completed", amountEUR)
}

func btcToEur(amountBTC float64) float64 {
	// Fetch exchange rate from API
	// Implement BTC to EUR conversion logic here
	return amountBTC * 50000 // Dummy conversion rate for demonstration
}
