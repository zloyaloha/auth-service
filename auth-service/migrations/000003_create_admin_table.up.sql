CREATE TYPE admin_role AS ENUM ('admin', 'moderator');

CREATE TABLE IF NOT EXISTS admins
(
    id   INT PRIMARY KEY,
    role admin_role NOT NULL,
    FOREIGN KEY (id) REFERENCES users(id)
);