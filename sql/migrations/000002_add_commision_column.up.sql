BEGIN;

-- We're going to need a validator state table in order to track validator
-- commission over time. This is also used in the calculation of the commission
-- earned in the escrow events. This calculation is added further down in this
-- file to the escrow_changes table.

CREATE TABLE IF NOT EXISTS public.validator_state (
    id integer NOT NULL,
    commission text NOT NULL,
    date timestamp WITHOUT TIME ZONE,
    height integer,
    validator text NOT NULL,
);

CREATE SEQUENCE IF NOT EXISTS public.validator_state_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

-- Create a sequence and set the id column to use it as default.
ALTER SEQUENCE public.validator_state_id_seq
OWNED BY       public.validator_state.id;

ALTER TABLE ONLY public.validator_state
ALTER COLUMN     id
SET DEFAULT      nextval('public.validator_state_id_seq'::regclass);

--------------------------------------------------------------------------------

-- Add commission column to escrow_changes and set the default for all existing
-- rows to 0.
ALTER TABLE  public.escrow_changes
ADD COLUMN   commission            TEXT;

UPDATE public.escrow_changes
SET    commission = '0';

-- Set Default.
ALTER TABLE  public.escrow_changes
ALTER COLUMN commission            SET DEFAULT '0';

-- Calculate Commisions for all previous Escrow Change rows.
UPDATE public.escrow_changes ec
SET    commission = (tokens::int8 * (
    SELECT   vs.commission::int8
    FROM     validator_state vs
    WHERE    vs.height <= ec.height
    ORDER BY vs.height DESC
    LIMIT    1
))::text;

COMMIT;
