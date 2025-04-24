INSERT INTO users (id, name) VALUES
(1, 'John Doe'),
(2, 'Jane Smith'),
(3, 'Bob Johnson'),
(4, 'Alice Williams'),
(5, 'Charlie Brown');

INSERT INTO orders (id, user_id, amount) VALUES
(1, 1, 99.99),
(2, 1, 149.50),
(3, 2, 75.25),
(4, 3, 299.99),
(5, 4, 45.75),
(6, 2, 199.99),
(7, 5, 88.50),
(8, 3, 120.00),
(9, 4, 65.99),
(10, 5, 250.00);

INSERT INTO payments (id, order_id, amount) VALUES
(1, 1, 99.99),
(2, 2, 149.50),
(3, 3, 75.25),
(4, 4, 299.99),
(5, 5, 45.75),
(6, 6, 199.99),
(7, 7, 88.50),
(8, 8, 120.00),
(9, 9, 65.99),
(10, 10, 250.00);