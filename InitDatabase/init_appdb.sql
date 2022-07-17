CREATE TABLE public.apptable (
	uuid int8 NOT NULL GENERATED ALWAYS AS IDENTITY,
	ts timestamp NOT NULL DEFAULT now(),
	name varchar NULL,
	hash varchar NULL
);
