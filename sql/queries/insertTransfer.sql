-- Write a Transfer Event to the database, this isn't a full transaction,
-- though a full transaction should be written to the transactions table that
-- matches any event here.
--
-- TODO: Constraint such that a Transaction/Event must exist in the database at
-- the same time.

--------------------------------------------------------------------------------

-- name: insertTransfer
INSERT INTO transfers ("from", "to", "tokens", "hash", "height", "date")
VALUES                ($1    , $2  , $3      , $4    , $5      , $6);
