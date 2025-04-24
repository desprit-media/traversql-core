INSERT INTO users (id, name) VALUES
(1, 'John Doe'),
(2, 'Jane Smith');

INSERT INTO orders (id, amount) VALUES
(1, 100.50),
(2, 200.75),
(3, 150.00),
(4, 75.25);

INSERT INTO payments (payment_id, amount) VALUES
(1, 50.25),
(2, 50.25),
(3, 200.75),
(4, 75.25);

INSERT INTO user_orders (user_id, order_id) VALUES
(1, 1),
(1, 2),
(2, 3),
(2, 4);

INSERT INTO order_payments (order_id, payment_id) VALUES
(1, 1),
(1, 2),
(2, 3),
(3, 4);