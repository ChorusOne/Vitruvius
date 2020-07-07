-- DO NOT EDIT.
-- This is the schema as it was before migrations were implemented, and so acts
-- as the base to start applying migrations from. To create a DB usable by the
-- extractor, apply this file with:
--
-- ```
-- $ vitruvius resetdb -r
-- ```
--
-- This file contains the full Schema definitions used by Oasis. They can be
-- used to construct a database from scratch capable of being used by vitruvius
-- for extraction.


-- Create Postgres Types
--------------------------------------------------------------------------------

CREATE TYPE public.escrow_change_kind AS ENUM (
    'add',
    'take',
    'reclaim'
);


-- Create Tables
--------------------------------------------------------------------------------

CREATE TABLE public.account_snapshots (
    id integer NOT NULL,
    address text,
    balance text,
    staked_balance text,
    debonding_balance text,
    rewards_balance text,
    delegations jsonb,
    is_validator boolean,
    is_delegator boolean,
    height integer,
    date timestamp without time zone
);

CREATE TABLE public.burns (
    id integer NOT NULL,
    date timestamp without time zone,
    hash text,
    height integer,
    owner text,
    tokens text
);

CREATE TABLE public.escrow_changes (
    id integer NOT NULL,
    date timestamp without time zone,
    escrow text,
    hash text,
    height integer,
    kind public.escrow_change_kind,
    owner text,
    tokens text
);

CREATE TABLE public.genesis_snapshots (
    id integer NOT NULL,
    snapshot_data jsonb,
    height integer,
    date timestamp without time zone
);

CREATE TABLE public.transactions (
    id integer NOT NULL,
    date timestamp without time zone,
    fee text,
    gas integer,
    gas_price text,
    hash text,
    height integer,
    method text NOT NULL,
    payload jsonb,
    sender text NOT NULL
);

CREATE TABLE public.transfers (
    id integer NOT NULL,
    date timestamp without time zone,
    "from" text,
    hash text,
    height integer,
    tokens text,
    "to" text
);


-- Create Sequences
--------------------------------------------------------------------------------

CREATE SEQUENCE public.account_snapshots_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE SEQUENCE public.burns_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE SEQUENCE public.escrow_changes_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE SEQUENCE public.genesis_snapshots_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE SEQUENCE public.transactions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE SEQUENCE public.transfers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


-- Set column ownership of sequences.
-- -----------------------------------------------------------------------------
-- If the table containing this column is dropped, the sequence will be dropped
-- as well, this makes sure we keep a hygienic database.

ALTER SEQUENCE public.transfers_id_seq         OWNED BY public.transfers.id;
ALTER SEQUENCE public.transactions_id_seq      OWNED BY public.transactions.id;
ALTER SEQUENCE public.genesis_snapshots_id_seq OWNED BY public.genesis_snapshots.id;
ALTER SEQUENCE public.escrow_changes_id_seq    OWNED BY public.escrow_changes.id;
ALTER SEQUENCE public.burns_id_seq             OWNED BY public.burns.id;
ALTER SEQUENCE public.account_snapshots_id_seq OWNED BY public.account_snapshots.id;


-- Set id column sequence defaults.
-- -----------------------------------------------------------------------------

ALTER TABLE ONLY public.account_snapshots ALTER COLUMN id SET DEFAULT nextval('public.account_snapshots_id_seq'::regclass);
ALTER TABLE ONLY public.burns             ALTER COLUMN id SET DEFAULT nextval('public.burns_id_seq'::regclass);
ALTER TABLE ONLY public.escrow_changes    ALTER COLUMN id SET DEFAULT nextval('public.escrow_changes_id_seq'::regclass);
ALTER TABLE ONLY public.genesis_snapshots ALTER COLUMN id SET DEFAULT nextval('public.genesis_snapshots_id_seq'::regclass);
ALTER TABLE ONLY public.transactions      ALTER COLUMN id SET DEFAULT nextval('public.transactions_id_seq'::regclass);
ALTER TABLE ONLY public.transfers         ALTER COLUMN id SET DEFAULT nextval('public.transfers_id_seq'::regclass);


-- Set Table Constraints
--------------------------------------------------------------------------------

ALTER TABLE ONLY public.account_snapshots ADD CONSTRAINT account_snapshots_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.burns             ADD CONSTRAINT burns_pkey             PRIMARY KEY (id);
ALTER TABLE ONLY public.escrow_changes    ADD CONSTRAINT escrow_changes_pkey    PRIMARY KEY (id);
ALTER TABLE ONLY public.genesis_snapshots ADD CONSTRAINT genesis_snapshots_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.transactions      ADD CONSTRAINT transactions_pkey      PRIMARY KEY (id);
ALTER TABLE ONLY public.transfers         ADD CONSTRAINT transfers_pkey         PRIMARY KEY (id);


