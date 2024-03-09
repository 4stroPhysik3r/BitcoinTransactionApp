package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
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

// var test_data = "transaction-data.json"
// var test_data = "example2.json"
var test_data = "example1.json"

func main() {
	mux := http.NewServeMux()

	initDB()
	generateTransactionID()

	// Handlers
	mux.HandleFunc("/", homePageHandler)
	mux.HandleFunc("/transactions", listTransactionsHandler)
	mux.HandleFunc("/balance", balanceHandler)
	mux.HandleFunc("/transfer", transferMoneyHandler)
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

	err = insertTransactionsFromJSON(db, "data/"+test_data)
	if err != nil {
		log.Fatal("Data: Error inserting data: ", err)
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

func balanceHandler(w http.ResponseWriter, r *http.Request) {
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
	// Round EUR_balance to two decimal places
	eurBalance = math.Round(eurBalance*100) / 100

	balance := map[string]string{"BTC_balance": strconv.FormatFloat(btcBalance, 'f', 5, 64), "EUR_balance": strconv.FormatFloat(eurBalance, 'f', 2, 64)} // display only until 2 decimal places

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
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

func listTransactionsHandler(w http.ResponseWriter, r *http.Request) {
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

func transferMoneyHandler(w http.ResponseWriter, r *http.Request) {
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
	// Convert EUR amount to BTC
	amountBTC := eurToBTC(amountEUR)

	// Check if the transfer amount is valid
	if amountBTC < 0.00001 {
		http.Error(w, "Transfer amount too small", http.StatusBadRequest)
		return
	}

	// Retrieve unspent transactions from the database
	unspentTransactions, err := getUnspentTransactions(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate total balance from unspent transactions
	totalBalance := calculateTotalBalance(unspentTransactions)

	// Check if there is enough balance to cover the transfer
	if totalBalance < amountBTC {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}

	// Mark used unspent transactions as spent
	leftoverAmount := markTransactionsAsSpent(db, unspentTransactions, amountBTC)

	// If there is a leftover amount, create a new unspent transaction
	if leftoverAmount > 0 {
		createNewUnspentTransaction(db, leftoverAmount)
	}

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

func eurToBTC(amountEUR float64) float64 {
	// Fetch exchange rate from API
	rate, err := fetchEurToBtcRate()
	if err != nil {
		log.Println("Error fetching exchange rate:", err)
		return 0
	}

	// Implement EUR to BTC conversion logic here
	return amountEUR / rate
}

func fetchEurToBtcRate() (float64, error) {
	resp, err := http.Get("http://api-cryptopia.adca.sh/v1/prices/ticker")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var responseData map[string][]map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		return 0, err
	}

	data, ok := responseData["data"]
	if !ok {
		return 0, fmt.Errorf("response does not contain data key")
	}

	for _, entry := range data {
		symbol, ok := entry["symbol"].(string)
		if ok && symbol == "BTC/EUR" {
			valueStr, ok := entry["value"].(string)
			if !ok {
				return 0, fmt.Errorf("value for BTC/EUR is not a string")
			}
			rate, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return 0, err
			}
			return rate, nil
		}
	}

	return 0, fmt.Errorf("exchange rate for BTC/EUR not found in response")
}

func getUnspentTransactions(db *sql.DB) ([]Transaction, error) {
	// Query the database to retrieve unspent transactions
	rows, err := db.Query("SELECT transaction_id, amount FROM transactions WHERE spent = 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var unspentTransactions []Transaction
	for rows.Next() {
		var tx Transaction
		err := rows.Scan(&tx.TransactionID, &tx.Amount)
		if err != nil {
			return nil, err
		}
		unspentTransactions = append(unspentTransactions, tx)
	}

	return unspentTransactions, nil
}

func calculateTotalBalance(transactions []Transaction) float64 {
	var totalBalance float64
	for _, tx := range transactions {
		totalBalance += tx.Amount
	}
	return totalBalance
}

func markTransactionsAsSpent(db *sql.DB, transactions []Transaction, amountBTC float64) float64 {
	var totalUsedAmount float64
	for _, tx := range transactions {
		// Mark the transaction as spent
		_, err := db.Exec("UPDATE transactions SET spent = 1 WHERE transaction_id = ?", tx.TransactionID)
		if err != nil {
			log.Println("Error marking transaction as spent:", err)
			continue
		}
		totalUsedAmount += tx.Amount
		// Check if the total used amount is greater than or equal to the transfer amount
		if totalUsedAmount >= amountBTC {
			break
		}
	}
	// Calculate leftover amount
	leftoverAmount := totalUsedAmount - amountBTC
	return leftoverAmount
}

func createNewUnspentTransaction(db *sql.DB, amountBTC float64) {
	// Generate transaction ID
	transactionID, err := generateTransactionID()
	if err != nil {
		log.Println("Error generating transaction ID:", err)
		return
	}

	// Insert a new unspent transaction with the leftover amount
	_, err = db.Exec("INSERT INTO transactions (transaction_id, amount, spent, created_at) VALUES (?, ?, ?, ?)",
		transactionID, amountBTC, 0, time.Now())
	if err != nil {
		log.Println("Error creating new unspent transaction:", err)
	}
}

func generateTransactionID() (string, error) {
	// Generate a random byte slice
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode the random bytes to hexadecimal string
	transactionID := hex.EncodeToString(randomBytes)
	return transactionID, nil
}
