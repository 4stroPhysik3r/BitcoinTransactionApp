package main

import (
	"bitcoin-transaction/functions"
	"bitcoin-transaction/handlers"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	mux := http.NewServeMux()

	functions.InitDB()

	// Handlers
	mux.HandleFunc("/", handlers.HomePageHandler)
	mux.HandleFunc("/transactions", handlers.ListTransactionsHandler)
	mux.HandleFunc("/balance", handlers.BalanceHandler)
	mux.HandleFunc("/transfer", handlers.TransferMoneyHandler)
	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
