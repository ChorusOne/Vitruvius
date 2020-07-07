-- This query is similar to queryAllTransactions, but also allows filtering
-- specifically for transactions that are associated with a specific oasis
-- address.

--------------------------------------------------------------------------------

-- name: queryAllTransactionsFiltered
SELECT   id,
         date,
         fee,
         gas,
         gas_price,
         hash,
         height,
         method,
         payload,
         sender
FROM     transactions
WHERE    (
    sender           LIKE $1 OR
    payload->>'to'   LIKE $1 OR
    payload->>'from' LIKE $1
)
ORDER BY date
LIMIT    $3
OFFSET   $2;
