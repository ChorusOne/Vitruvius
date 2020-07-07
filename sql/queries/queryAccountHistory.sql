-- Fetch Account Snapshots for some address.

--------------------------------------------------------------------------------

-- name: queryAccountHistory
SELECT   * FROM account_snapshots
WHERE    address = $1
ORDER BY date
LIMIT    $3
OFFSET   $2;
