-- Insert Transactions into the database, there should be one transaction in
-- this table for EACH event type stored in others, such as burn, transfer, or
-- escrow events.

--------------------------------------------------------------------------------

-- name: insertTransaction
INSERT INTO transactions ("method", "payload", "height", "date", "sender", "fee", "gas", "gas_price", "hash")
VALUES                   ($1      , $2       , $3      , $4    , $5      , $6   , $7   , $8         , $9);
