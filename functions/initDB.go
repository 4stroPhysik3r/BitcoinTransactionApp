package functions

import (
	"database/sql"
	"log"
)

var testData = []string{
	"badExample.json",
	"emptyData.json",
	"example1.json",
	"example2.json",
	"transaction-data.json",
}

var db *sql.DB

func InitDB() {
	log.Println("Database Initiated")

	var err error
	db, err = sql.Open("sqlite3", "db/transactions.db")
	if err != nil {
		log.Fatal("DB: Error opening database", err)
	}

	/// testing purposes only
	// Drop previous tables
	_, err = db.Exec(`DROP TABLE IF EXISTS transactions;`)
	if err != nil {
		log.Fatal("DB: Error dropping tables", err)
	}

	// Create transactions table if it does not exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS transactions (
		transaction_id TEXT PRIMARY KEY,
		amount REAL,
		spent BOOLEAN NOT NULL,
		created_at TIMESTAMP
	);`)
	if err != nil {
		log.Fatal("DB: Error executing tables", err)
	}

	// "testData" represents array of with test data from the "data" folder
	err = insertTransactionsFromJSON(db, "data/"+testData[4]) // change data here
	if err != nil {
		log.Fatal("Data: Error inserting data: ", err)
	}
	////////////////////////////////
}
