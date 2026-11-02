CREATE TABLE orders (
    id VARCHAR(255) NOT NULL,
    price DOUBLE NOT NULL,
    tax DOUBLE NOT NULL,
    final_price DOUBLE NOT NULL,
    PRIMARY KEY (id)
);
