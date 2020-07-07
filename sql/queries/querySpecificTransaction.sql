-- This query specifically filters for a transaction where the hash is
-- specified explicitly. This is to facilitate Anthem's need to search by hash.
--------------------------------------------------------------------------------

-- name: querySpecificTransaction
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
WHERE    hash = $1
LIMIT    1;
