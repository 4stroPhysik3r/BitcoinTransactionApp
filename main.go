package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

var db *sql.DB

func main() {
	mux := http.NewServeMux()

	initDB()

	// Handlers
	mux.HandleFunc("/", homePageHandler)
	mux.HandleFunc("/transactions", listTransactions)
	mux.HandleFunc("/balance", getBalance)
	mux.HandleFunc("/transfer", transferMoney)
	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func initDB() {
	log.Println("Database Initiated")

	var err error
	db, err = sql.Open("sqlite3", "db/transactions.db")
	if err != nil {
		log.Fatal("DB: Error opening database", err)
	}
	// defer db.Close()

	// Create transactions table if it does not exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS transactions (
		transaction_id TEXT PRIMARY KEY,
		amount REAL,
		spent INTEGER,
		created_at TIMESTAMP
	);`)
	if err != nil {
		log.Fatal("DB: Error executing tables", err)
	}

	err = insertTransactionsFromJSON(db, "data/transaction-data.json")
	if err != nil {
		log.Fatal("Data: Error inserting dummy data: ", err)
	}
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
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
	http.FileServer(http.Dir(".")).ServeHTTP(w, r)
}

func insertTransactionsFromJSON(db *sql.DB, filePath string) error {
	// Open the JSON file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode the JSON data
	var transactions []Transaction
	err = json.NewDecoder(file).Decode(&transactions)
	if err != nil {
		return err
	}

	// Prepare SQL statement for checking if transaction exists
	stmtCheck, err := db.Prepare("SELECT EXISTS(SELECT 1 FROM transactions WHERE transaction_id = ?)")
	if err != nil {
		return err
	}
	defer stmtCheck.Close()

	log.Println("Inserting transaction data into database")
	// Insert transactions into the database
	for _, tx := range transactions {
		// Check if transaction ID already exists
		var exists bool
		err := stmtCheck.QueryRow(tx.TransactionID).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			// Transaction already exists, skip insertion
			continue
		}

		// Insert transaction into the database
		_, err = db.Exec("INSERT INTO transactions (transaction_id, amount, spent, created_at) VALUES (?, ?, ?, ?)",
			tx.TransactionID, tx.Amount, tx.Spent, tx.CreatedAt)
		if err != nil {
			return err
		}
	}

	return nil
}

func listTransactions(w http.ResponseWriter, r *http.Request) {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	// log.Println("List transactions handler fired!")

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

func getBalance(w http.ResponseWriter, r *http.Request) {
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

	eurBalance := btcToEur(btcBalance)
	balance := map[string]float64{"BTC_balance": btcBalance, "EUR_balance": eurBalance}
	json.NewEncoder(w).Encode(balance)
}

func transferMoney(w http.ResponseWriter, r *http.Request) {
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
	amountEUR := data["amount"]
	// Implement transfer logic here
	fmt.Fprintf(w, "Transfer of %.2f EUR completed", amountEUR)
}

func btcToEur(amountBTC float64) float64 {
	// Fetch exchange rate from API
	rate, err := fetchEurToBtcRate()
	if err != nil {
		log.Println("Error fetching exchange rate:", err)
		return 0
	}

	// Implement BTC to EUR conversion logic here
	return amountBTC * rate
}

func fetchEurToBtcRate() (float64, error) {
	resp, err := http.Get("http://api-cryptopia.adca.sh/v1/prices/ticker")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var data map[string]float64
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return 0, err
	}

	rate, ok := data["BTC_EUR"]
	if !ok {
		return 0, fmt.Errorf("exchange rate not found in response")
	}

	return rate, nil
}
