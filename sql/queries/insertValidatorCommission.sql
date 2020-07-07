-- This query will attempt to insert the current commission for a validator. It
-- will attempt to find an existing entry for this validator with the current
-- commission and if found, will insert no row at all. This keeps the changes
-- in the table unique.
--
-- $1 = New Commission
-- $2 = Validator Address
-- $3 = Height

--------------------------------------------------------------------------------

-- name: insertValidatorCommission
INSERT INTO validator_state (commission, date, height, validator)
SELECT $1, NOW(), $3, $2
WHERE  NOT EXISTS (
    SELECT   commission
    FROM     validator_state
    WHERE    commission = $1 AND
             validator  = $2
    ORDER BY id DESC
    LIMIT 1
);
