-- Insert an Account snapshot into the database, this row should be equivelent
-- to the shape of data returned by the /api/account/<address> endpoint.

--------------------------------------------------------------------------------

-- name: insertSnapshot
INSERT INTO account_snapshots (
    "address",
    "balance",
    "staked_balance",
    "debonding_balance",
    "rewards_balance",
    "delegations",
    "is_validator",
    "is_delegator",
    "height",
    "date"
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10);
