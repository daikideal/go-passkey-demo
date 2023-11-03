CREATE DATABASE mydb;

-- mydb へ接続
\c mydb;

CREATE SCHEMA myschema;

CREATE ROLE myuser
WITH
  LOGIN PASSWORD 'mypassword';

GRANT ALL PRIVILEGES SCHEMA myschema TO myuser;