-- This fetches the last event that added escrow to an account, by scanning for
-- the special address that identifies the common pool.

--------------------------------------------------------------------------------

-- name: queryGetLastRewardBalance
SELECT SUM(tokens::int8)
FROM   escrow_changes
WHERE  owner   = 'oasis1qrmufhkkyyf79s5za2r8yga9gnk4t446dcy3a5zm' AND
       escrow  = $1 AND
       height <= $2;
