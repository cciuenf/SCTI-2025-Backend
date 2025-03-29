-- create.sql

-- Create the database
CREATE DATABASE "scti-db";

-- Create the user with a password
CREATE USER "scti-user" WITH ENCRYPTED PASSWORD 'scti#01';

-- Grant all privileges on the database to the user
GRANT ALL PRIVILEGES ON DATABASE "scti-db" TO "scti-user";

-- Allow the user to create schemas and tables
ALTER USER "scti-user" CREATEDB;

-- Connect to the database and grant privileges on all schemas and tables
\connect "scti-db";
GRANT ALL PRIVILEGES ON SCHEMA public TO "scti-user";
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO "scti-user";
