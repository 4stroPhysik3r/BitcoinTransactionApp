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
