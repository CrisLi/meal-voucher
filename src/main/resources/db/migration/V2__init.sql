DROP TABLE IF EXISTS vouchers;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;

CREATE TABLE teams (
  id   INT PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE users (
  id           INT PRIMARY KEY,
  login        VARCHAR(64)  NOT NULL UNIQUE,
  password     VARCHAR(255) NOT NULL,
  display_name VARCHAR(255),
  role         VARCHAR(64)  NOT NULL,
  team_id      INT          NOT NULL,
  FOREIGN KEY (team_id) REFERENCES teams(id)
);

CREATE TABLE vouchers (
  id      INT PRIMARY KEY,
  team_id INT NOT NULL,
  year    INT NOT NULL,
  month   INT NOT NULL,
  quota   INT NOT NULL,
  FOREIGN KEY (team_id) REFERENCES teams(id)
);