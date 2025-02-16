CREATE TABLE Items (
    "item_id" serial PRIMARY KEY,
    "item_type" varchar(50) NOT NULL UNIQUE,
    "item_price" smallint NOT NULL
);

INSERT INTO Items (item_type, item_price) VALUES
    ('t-shirt', 80),
    ('cup', 20),
    ('book', 50),
    ('pen', 10),
    ('powerbank', 200),
    ('hoody', 300),
    ('umbrella', 200),
    ('socks', 10),
    ('wallet', 50),
    ('pink-hoody', 500);

CREATE INDEX idx_item_type ON Items(item_type);

CREATE TABLE Users (
    "user_id" serial PRIMARY KEY,
    "username" varchar NOT NULL UNIQUE,
    "password" varchar NOT NULL,
    "coins" int NOT NULL DEFAULT 1000
);
CREATE INDEX idx_username ON Users(username);
INSERT INTO Users (username, password) VALUES
    ('testuser', 'testpassword');

CREATE TABLE Inventory (
    "inventory_id" serial PRIMARY KEY,
    "user_id" int REFERENCES Users(user_id) NOT NULL,
    "item_type" varchar(50) REFERENCES Items(item_type) NOT NULL,
    "quantity" int NOT NULL DEFAULT 1,
    UNIQUE (user_id, item_type)
);
CREATE INDEX idx_inventory_user_item ON Inventory (user_id, item_type);

CREATE TABLE Transfers (
    "transfer_id" serial PRIMARY KEY,
    "from_username" varchar REFERENCES Users(username) NOT NULL,
    "to_username" varchar REFERENCES Users(username) NOT NULL,
    "amount" int NOT NULL
);
CREATE INDEX idx_transfers_from_username ON Transfers(from_username);
CREATE INDEX idx_transfers_to_username ON Transfers(to_username);
