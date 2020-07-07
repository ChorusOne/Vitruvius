-- Get the number of account snapshots that exist for a user, this is for
-- pagination purposes.

--------------------------------------------------------------------------------

-- name: queryAccountHistoryLength
SELECT count(*) FROM account_snapshots
WHERE  address = $1;
