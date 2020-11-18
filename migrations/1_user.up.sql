/*CREATE EXTENSION "uuid-ossp";*/
CREATE TABLE users (
  id text default uuid_generate_v4(),
  username text,
  email text,
  password text,
  last_login timestamp default now()
);

INSERT INTO users (id, username, email, password) VALUES (uuid_generate_v4(), 'piko', 'herpiko@gmail.com', 'XXXXX');
