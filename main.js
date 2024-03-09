// Fetch balance data from Go server
fetch("/balance")
   .then((response) => response.json())
   .then((balance) => {
      const balanceDiv = document.getElementById("balance");

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
      const transactionsDiv = document.getElementById("transactions");

      if (data.length === 0) {
         const noHistoryMessage = document.createElement("div");
         noHistoryMessage.textContent = "No transaction history";
         transactionsDiv.appendChild(noHistoryMessage);
      } else {
         data.forEach((transaction) => {
            if (!transaction.spent) { // Check if transaction is not spent
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
      }
   })
   .catch((error) =>
      console.error("Error fetching transaction data:", error)
   );

// Fetch transaction data from Go server
fetch("/transactions")
   .then((response) => response.json())
   .then((data) => {
      const transactionHistoryDiv = document.getElementById("transactionsHistory");

      // Check if there are any spent transactions
      const noSpentTransactions = data.every(transaction => !transaction.spent);

      if (noSpentTransactions) {
         const noHistoryMessage = document.createElement("div");
         noHistoryMessage.textContent = "No transaction history";
         transactionHistoryDiv.appendChild(noHistoryMessage);
      } else {
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
      }
   })
   .catch((error) =>
      console.error("Error fetching transaction data:", error)
   );