// Fetch balance data from Go server
fetch("/balance")
   .then((response) => response.json())
   .then((balance) => {
      const balanceDiv = document.getElementById("balance");

      balanceDiv.innerHTML = `
            <h2>Your Balance:</h2>
            <p><strong>BTC Balance:</strong> ${balance.BTC_balance} ₿ / ${balance.EUR_balance} €</p>
            `;

      // document.body.appendChild(balanceDiv)
   })
   .catch((error) =>
      console.error("Error fetching balance data:", error)
   );

// Fetch transaction data from Go server
fetch("/transactions")
   .then((response) => response.json())
   .then((data) => {
      const transactionsDiv = document.getElementById("transactions");

      data.forEach((transaction) => {
         const transactionDiv = document.createElement("div");
         transactionDiv.classList.add("transaction");

         transactionDiv.innerHTML = `
               <p class="transaction"><strong>Transaction ID:</strong> ${transaction.transaction_id}</p>
               <p class="amount"><strong>Amount:</strong> ${transaction.amount}</p>
               <p class="spent"><strong>Spent:</strong> ${transaction.spent ? "Yes" : "No"}</p>
               <p class="date"><strong>Created At:</strong> ${new Date(transaction.created_at).toLocaleString()}</p>
               `;

         transactionsDiv.appendChild(transactionDiv);
      });
   })
   .catch((error) =>
      console.error("Error fetching transaction data:", error)
   );