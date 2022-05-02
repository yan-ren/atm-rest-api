CREATE USER admin WITH PASSWORD 'admin123' CREATEDB;

CREATE DATABASE dev WITH OWNER admin;

\connect dev;

CREATE EXTENSION pgcrypto;

CREATE SCHEMA test AUTHORIZATION admin;

CREATE TABLE test.customer(
	id serial PRIMARY KEY,
	email VARCHAR(255),
	password VARCHAR(255),
	first_name VARCHAR(255),
	last_name VARCHAR(255)
);

CREATE TABLE test.account(id serial PRIMARY KEY, balance INTEGER);

CREATE TABLE test.customer_account(
	customer_id integer REFERENCES test.customer(id),
	account_id integer REFERENCES test.account(id),
	PRIMARY KEY(customer_id, account_id)
);

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA test TO admin;

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA test TO admin;

INSERT INTO
	test.customer(email, password, first_name, last_name)
VALUES('test@axiomzen.co', '1234', 'axiom', 'zen');

INSERT INTO test.account(balance) VALUES (100);

INSERT INTO test.account(balance) VALUES (200);

INSERT INTO
	test.customer_account(customer_id, account_id)
VALUES
	(1, 1);

INSERT INTO
	test.customer_account(customer_id, account_id)
VALUES
	(1, 2);