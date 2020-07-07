BEGIN;

-- This table tracks state information. The kind is a key for lookup of the
-- value and the value is text so serialization can store any type of data.
CREATE TABLE IF NOT EXISTS public.chain_sync_state(
    kind  text NOT NULL PRIMARY KEY,
    value text NOT NULL
);

COMMIT;
