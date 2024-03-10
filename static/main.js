const balanceDiv = document.getElementById("balance");
const transactionsDiv = document.getElementById("transactions");
const transactionHistoryDiv = document.getElementById("transactionsHistory");

// Fetch balance data from Go server
fetch("/balance")
   .then((response) => response.json())
   .then((balance) => {
      balanceDiv.innerHTML = `
            <h2>Your Balance:</h2>
            <p><strong>BTC Balance:</strong> ${balance.BTC_balance} ₿ / ${numberWithCommas(balance.EUR_balance)} €</p>
            `;
   })
   .catch((error) =>
      console.error("Error fetching balance data:", error)
   );

// Format balance into human readable format
function numberWithCommas(x) {
   return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
}

// Fetch transaction data from Go server
fetch("/transactions")
   .then((response) => response.json())
   .then((data) => {
      data.forEach((data) => {
         if (data && data.amount > 0) {
            transactionsDiv.innerHTML = ""
         }
      })

      data.forEach((transaction) => {
         if (!transaction.spent && transaction.amount != 0) { // Check if transaction is not spent or zero

            const transactionDiv = document.createElement("div");
            transactionDiv.classList.add("transaction");

            transactionDiv.innerHTML = `
               <p class="transaction"><strong>Transaction ID:</strong> ${transaction.transaction_id}</p>
               <p class="amount"><strong>Amount:</strong> ${transaction.amount}</p>
               <p class="spent"><strong>Spent:</strong> ${transaction.spent ? "Yes" : "No"}</p>
               <p class="date"><strong>Created At:</strong> ${new Date(transaction.created_at).toLocaleString()}</p>
               `;

            transactionsDiv.appendChild(transactionDiv);
         }
      });

   })
   .catch((error) => {
      transactionsDiv.innerHTML = ""

      const noTransactionsMessage = document.createElement("div");
      noTransactionsMessage.textContent = "Error fetching unspent transactions";
      transactionsDiv.appendChild(noTransactionsMessage);
      console.error("Error fetching transaction data:", error)
   });

// Fetch transaction data from Go server
fetch("/transactions")
   .then((response) => response.json())
   .then((data) => {
      data.forEach((data) => {
         if (data.spent) {
            transactionHistoryDiv.innerHTML = ""
         }
      })

      data.forEach((transaction) => {
         if (transaction.spent) { // Check if transaction is not spent

            const transactionHistory = document.createElement("div");
            transactionHistory.classList.add("transactionHistory");

            transactionHistory.innerHTML = `
                  <p class="transaction"><strong>Transaction ID:</strong> ${transaction.transaction_id}</p>
                  <p class="amount"><strong>Amount:</strong> ${transaction.amount}</p>
                  <p class="spent"><strong>Spent:</strong> ${transaction.spent ? "Yes" : "No"}</p>
                  <p class="date"><strong>Created At:</strong> ${new Date(transaction.created_at).toLocaleString()}</p>
                  `;

            transactionHistoryDiv.appendChild(transactionHistory);
         }
      });
   })
   .catch((error) => {
      transactionHistoryDiv.innerHTML = ""

      const noHistoryMessage = document.createElement("div");
      noHistoryMessage.textContent = "Error fetching transaction history";
      transactionHistoryDiv.appendChild(noHistoryMessage);
      console.error("Error fetching transaction data:", error)
   });
