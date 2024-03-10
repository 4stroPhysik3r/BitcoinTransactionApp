package functions

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Transaction represents a Bitcoin transaction
type Transaction struct {
	TransactionID string    `json:"transaction_id"`
	Amount        float64   `json:"amount"`
	Spent         bool      `json:"spent"`
	CreatedAt     time.Time `json:"created_at"`
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

	// Insert transactions into the database
	for _, tx := range transactions {
		// Check if transaction ID already exists
		var exists bool
		err := stmtCheck.QueryRow(tx.TransactionID).Scan(&exists)
		if err != nil {
			return err
		}
		// Transaction already exists, skip insertion
		if exists {
			continue
		}

		// Insert transaction into the database
		_, err = db.Exec("INSERT INTO transactions (transaction_id, amount, spent, created_at) VALUES (?, ?, ?, ?)",
			tx.TransactionID, tx.Amount, tx.Spent, tx.CreatedAt)
		if err != nil {
			return err
		}
	}
	log.Println("Inserted transaction data into database")

	return nil
}

func BtcToEur(amountBTC float64) float64 {
	// Fetch exchange rate from API
	rate, err := fetchEurToBtcRate()
	if err != nil {
		log.Println("Error fetching exchange rate:", err)
		return 0
	}

	// Implement BTC to EUR conversion logic here
	return amountBTC * rate
}

func EurToBTC(amountEUR float64) float64 {
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
		return 0, fmt.Errorf("BC fetch: response does not contain data key")
	}

	for _, entry := range data {
		symbol, ok := entry["symbol"].(string)
		if ok && symbol == "BTC/EUR" {
			valueStr, ok := entry["value"].(string)
			if !ok {
				return 0, fmt.Errorf("BC fetch: value for BTC/EUR is not a string")
			}
			rate, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return 0, err
			}
			return rate, nil
		}
	}

	return 0, fmt.Errorf("BC fetch: exchange rate for BTC/EUR not found in response")
}

func GetUnspentTransactions(db *sql.DB) ([]Transaction, error) {
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

func CalculateTotalBalance(transactions []Transaction) float64 {
	var totalBalance float64
	for _, tx := range transactions {
		totalBalance += tx.Amount
	}
	return totalBalance
}

func MarkTransactionsAsSpent(db *sql.DB, transactions []Transaction, amountBTC float64) float64 {
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

func CreateNewUnspentTransaction(db *sql.DB, amountBTC float64) {
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
