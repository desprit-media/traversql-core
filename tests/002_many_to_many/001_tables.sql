CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    amount DECIMAL(10, 2) NOT NULL
);

-- PK as a non-"id" column
CREATE TABLE payments (
    payment_id SERIAL PRIMARY KEY,
    amount DECIMAL(10, 2) NOT NULL
);

CREATE TABLE user_orders (
    user_id INTEGER NOT NULL,
    order_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, order_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

CREATE TABLE order_payments (
    order_id INTEGER NOT NULL,
    payment_id INTEGER NOT NULL,
    PRIMARY KEY (order_id, payment_id),
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (payment_id) REFERENCES payments(payment_id) ON DELETE CASCADE
);
