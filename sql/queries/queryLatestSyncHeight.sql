-- Query the last processed height for the extractor. This is important for
-- processing data that we've missed if the extractor was killed, as well as
-- just making sure we don't sync from 0 every time the extractor restarts.

--------------------------------------------------------------------------------

-- name: queryLatestSyncHeight
SELECT value::int8
FROM   chain_sync_state
WHERE  kind = 'sync_height';
