package testutil

// TestPGDump is a simple example pgdump file for a table containing a few cars
const TestPGDump = `
--
-- PostgreSQL database dump
--

-- Dumped from database version 11.1 (Debian 11.1-1.pgdg90+1)
-- Dumped by pg_dump version 11.1 (Debian 11.1-1.pgdg90+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: cars; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cars (
    make character varying(50),
    model character varying(50),
    year smallint,
    note character varying(100) DEFAULT 'literally a rocket'::character varying
);


ALTER TABLE public.cars OWNER TO postgres;

--
-- Data for Name: cars; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.cars (make, model, year, note) FROM stdin;
Tesla	Roadster	2008	literally a rocket
Bugatti	Chiron	2016	literally a rocket
Dodge	Viper	2015	literally a rocket
Honda	Civic	1998	only a rocket if it has a spoiler
Toyota	Corolla	2005	greatest car ever made
\.


--
-- PostgreSQL database dump complete
--
`

// TestPGDumpNewRows is the same as TestPGDump, except that it has two new rows
const TestPGDumpNewRows = `
--
-- PostgreSQL database dump
--

-- Dumped from database version 11.1 (Debian 11.1-1.pgdg90+1)
-- Dumped by pg_dump version 11.1 (Debian 11.1-1.pgdg90+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: cars; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cars (
    make character varying(50),
    model character varying(50),
    year smallint,
    note character varying(100) DEFAULT 'literally a rocket'::character varying
);


ALTER TABLE public.cars OWNER TO postgres;

--
-- Data for Name: cars; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.cars (make, model, year, note) FROM stdin;
Fiat	Panda	1981	If you extend the middle, it makes a great limo
Little Tikes	Cozy Coupe	2018	Roomier than the panda, but the headlights are actually eyes
Sun Fresh	Russet	2019	literally a potato
\.


--
-- PostgreSQL database dump complete
--
`

// TestPGDumpNewHeader has no rows, but has a new header
const TestPGDumpNewHeader = `
--
-- PostgreSQL database dump
--

-- Dumped from database version 11.1 (Debian 11.1-1.pgdg90+1)
-- Dumped by pg_dump version 11.1 (Debian 11.1-1.pgdg90+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: cars; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cars (
    make character varying(50),
    model character varying(50),
    year smallint,
    note character varying(100) DEFAULT 'literally a potato'::character varying
);


ALTER TABLE public.cars OWNER TO postgres;

--
-- Data for Name: cars; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.cars (make, model, year, note) FROM stdin;
\.


--
-- PostgreSQL database dump complete
--
`
