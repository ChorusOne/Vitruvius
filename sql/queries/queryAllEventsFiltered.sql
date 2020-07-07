-- This query is similar to queryAllEvents, but allows filtering every payload
-- by some string. This is useful for finding addresses.

--------------------------------------------------------------------------------

-- name: queryAllEventsFiltered
WITH all_events AS (
    SELECT   0                                AS height,
             NOW()::text                      AS when,
             'transfer'                       AS kind,
             ( SELECT row_to_json(_)
             FROM  (SELECT t.*) AS _)::text AS payload
    FROM     transfers t
    UNION

    -- Condense Escrow Events
    SELECT   0                                 AS height,
             NOW()::text                       AS when,
             'escrow'                          AS kind,
             json_build_object(kind,
             ( SELECT row_to_json(_)
             FROM  (SELECT e.*) AS _))::text AS payload
    FROM     escrow_changes e
    UNION

    -- Condense Burn Events
    SELECT   0                              AS height,
             NOW()::text                    AS when,
             'burn'                         AS kind,
             ( SELECT row_to_json(_)
             FROM  (SELECT b.*) AS _)::text AS payload
    FROM     burns b
)

SELECT   *
FROM     all_events
WHERE    payload LIKE $1
ORDER BY "when"
LIMIT    $3
OFFSET   $2;
