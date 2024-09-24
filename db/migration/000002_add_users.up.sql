CREATE TABLE users (
  id bigserial NOT NULL,
  username varchar PRIMARY KEY,
  hashed_password varchar NOT NULL,
  full_name varchar NOT NULL,
  email varchar UNIQUE NOT NULL,
  created_at timestamptz NOT NULL DEFAULT (NOW()),
  modified_at timestamptz NOT NULL DEFAULT (NOW())
);

ALTER TABLE accounts ADD FOREIGN KEY ("owner") REFERENCES users ("username");

ALTER TABLE accounts ADD CONSTRAINT "owner_currency" UNIQUE ("owner", "currency");
