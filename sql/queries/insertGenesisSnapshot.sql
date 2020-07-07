-- Insert a Genesis Snapshot, these are full JSON dumps of the state of the
-- chain. These should be stored as diffs when possible to reduce the massive
-- size this would otherwise be.

--------------------------------------------------------------------------------

-- name: insertGenesisSnapshot
INSERT INTO genesis_snapshots ("snapshot_data", "height", "date")
VALUES                        ($1             , $2      , $3);
