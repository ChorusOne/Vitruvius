-- Events are stored in SQL in different row shapes, this query will create
-- JSON objects out of each disparate event type and return a homogenous table
-- of events labeled by kind.

--------------------------------------------------------------------------------

-- name: queryAllEvents
WITH all_events AS (
    SELECT   0                              AS height,
             NOW()::text                    AS when,
             'transfer'                     AS kind,
             ( SELECT row_to_json(_)
             FROM  (SELECT t.*) AS _)::text AS payload
    FROM     transfers t
    UNION

    -- Condense Escrow Events
    SELECT   0                               AS height,
             NOW()::text                     AS when,
             'escrow'                        AS kind,
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
ORDER BY height
LIMIT    $2
OFFSET   $1;
