BEGIN;

-- Drop Commission table. The up migration for this takes 3 steps but to
-- reverse it we only need to remove the colum that was added.
ALTER TABLE  public.escrow_changes
DROP  COLUMN commission;

-- Drop Sequence tracking ID column
DROP  TABLE    IF EXISTS public.validator_state;
DROP  SEQUENCE IF EXISTS public.validator_state_id_seq;

COMMIT;
