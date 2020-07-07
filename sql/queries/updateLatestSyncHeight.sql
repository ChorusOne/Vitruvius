-- Update latest sync height. This should be run every time each block is
-- processed so that the extractor does not accidentally write cloned data.

--------------------------------------------------------------------------------

-- name: updateLatestSyncHeight
INSERT INTO chain_sync_state (kind          , value)
VALUES                       ('sync_height' , $1)
ON CONFLICT (kind)
DO UPDATE   SET value = $1;
