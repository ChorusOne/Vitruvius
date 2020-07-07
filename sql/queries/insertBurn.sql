-- Write a Burn Event to the database, this isn't a full transaction, though a
-- full transaction should be written to the transactions table that matches
-- any event here.

--------------------------------------------------------------------------------

-- name: insertBurn
INSERT INTO burns ("owner", "tokens", "hash", "height", "date")
VALUES            ($1     , $2      , $3    , $4      , $5);
