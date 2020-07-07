-- Fetch all Transactions from the database, but filters out any query by the
-- registry module (these seem pointless to show for 99% of users).

--------------------------------------------------------------------------------

-- name: queryAllTransactions
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
WHERE    method NOT LIKE '%registry%'
ORDER BY date
LIMIT    $2
OFFSET   $1;
