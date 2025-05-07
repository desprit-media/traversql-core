-- Insert categories
INSERT INTO categories (category_id, name, description) VALUES
(1, 'Electronics', 'Electronic devices and accessories'),
(2, 'Clothing', 'Apparel and fashion items'),
(3, 'Home & Kitchen', 'Products for home and kitchen use'),
(4, 'Books', 'Books, e-books, and publications'),
(5, 'Sports & Outdoors', 'Sporting goods and outdoor equipment');

-- Insert customers
INSERT INTO customers (customer_id, first_name, last_name, email, phone, date_of_birth, loyalty_points) VALUES
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'John', 'Doe', 'john.doe@example.com', '555-123-4567', '1985-07-15', 120),
('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'Jane', 'Smith', 'jane.smith@example.com', '555-987-6543', '1990-03-22', 85),
('c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'Michael', 'Johnson', 'michael.j@example.com', '555-456-7890', '1978-11-30', 250),
('d3eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'Emily', 'Williams', 'emily.w@example.com', '555-789-0123', '1992-05-08', 65),
('e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 'Robert', 'Brown', 'robert.b@example.com', '555-234-5678', '1982-09-17', 180);

-- Insert customer addresses
INSERT INTO customer_addresses (address_id, customer_id, address_type, street_address, city, state, postal_code, country, is_default) VALUES
(1, 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'BOTH', '123 Main St', 'New York', 'NY', '10001', 'USA', TRUE),
(2, 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'SHIPPING', '456 Oak Ave', 'Los Angeles', 'CA', '90001', 'USA', TRUE),
(3, 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'BILLING', '789 Pine Rd', 'Los Angeles', 'CA', '90002', 'USA', TRUE),
(4, 'c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'BOTH', '101 Maple Dr', 'Chicago', 'IL', '60007', 'USA', TRUE),
(5, 'd3eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'SHIPPING', '202 Cedar Ln', 'Houston', 'TX', '77001', 'USA', TRUE),
(6, 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 'BOTH', '303 Birch Blvd', 'Phoenix', 'AZ', '85001', 'USA', TRUE);

-- Insert products
INSERT INTO products (product_id, name, description, sku, price, stock_quantity, weight, dimensions) VALUES
(1, 'Smartphone X', '6.5-inch display, 128GB storage', 'PHON-X-128', 699.99, 50, 0.35, '{"width": 7.5, "height": 15.0, "depth": 0.8}'),
(2, 'Laptop Pro', '15-inch laptop with 16GB RAM, 512GB SSD', 'LAPT-PRO-15', 1299.99, 25, 2.5, '{"width": 35.0, "height": 24.0, "depth": 1.8}'),
(3, 'Cotton T-Shirt', 'Comfortable cotton t-shirt, available in multiple colors', 'APRL-TS-M', 19.99, 200, 0.2, '{"width": 50.0, "height": 70.0, "depth": 1.0}'),
(4, 'Coffee Maker', 'Programmable coffee maker with 12-cup capacity', 'HOME-CM-12', 49.99, 30, 3.0, '{"width": 25.0, "height": 35.0, "depth": 20.0}'),
(5, 'Wireless Earbuds', 'Bluetooth earbuds with noise cancellation', 'AUDIO-EB-NC', 129.99, 75, 0.1, '{"width": 5.0, "height": 5.0, "depth": 3.0}'),
(6, 'Yoga Mat', 'Non-slip yoga mat, 6mm thickness', 'SPORT-YM-6', 29.99, 100, 1.2, '{"width": 61.0, "height": 180.0, "depth": 0.6}'),
(7, 'Novel: The Mystery', 'Bestselling mystery novel', 'BOOK-MYS-01', 14.99, 150, 0.5, '{"width": 15.0, "height": 22.0, "depth": 2.5}'),
(8, 'Smart Watch', 'Fitness tracker and smartwatch', 'WEAR-SW-FIT', 199.99, 40, 0.05, '{"width": 4.0, "height": 4.0, "depth": 1.2}');

-- Connect products to categories (product_categories)
INSERT INTO product_categories (product_id, category_id) VALUES
(1, 1), -- Smartphone in Electronics
(2, 1), -- Laptop in Electronics
(3, 2), -- T-Shirt in Clothing
(4, 3), -- Coffee Maker in Home & Kitchen
(5, 1), -- Earbuds in Electronics
(6, 5), -- Yoga Mat in Sports & Outdoors
(7, 4), -- Novel in Books
(8, 1), -- Smart Watch in Electronics
(8, 5); -- Smart Watch also in Sports & Outdoors

-- Insert orders
INSERT INTO orders (order_id, customer_id, status, shipping_address_id, billing_address_id, shipping_method, payment_method, subtotal, tax, shipping_cost, total_amount, notes) VALUES
(1, 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'DELIVERED', 1, 1, 'Standard Shipping', 'Credit Card', 699.99, 56.00, 10.00, 765.99, NULL),
(2, 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'SHIPPED', 2, 3, 'Express Shipping', 'PayPal', 149.98, 12.00, 15.00, 176.98, 'Please leave at the door'),
(3, 'c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'PROCESSING', 4, 4, 'Standard Shipping', 'Credit Card', 1299.99, 104.00, 0.00, 1403.99, 'Gift wrap requested'),
(4, 'd3eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'PENDING', 5, 5, 'Standard Shipping', 'Bank Transfer', 44.98, 3.60, 5.00, 53.58, NULL),
(5, 'e4eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 'DELIVERED', 6, 6, 'Express Shipping', 'Credit Card', 329.98, 26.40, 15.00, 371.38, NULL);

-- Insert order items
INSERT INTO order_items (order_item_id, order_id, product_id, quantity, unit_price, total_price, discount_amount) VALUES
(1, 1, 1, 1, 699.99, 699.99, 0.00),
(2, 2, 3, 2, 19.99, 39.98, 0.00),
(3, 2, 7, 1, 14.99, 14.99, 0.00),
(4, 3, 2, 1, 1299.99, 1299.99, 0.00),
(5, 4, 7, 3, 14.99, 44.97, 0.00),
(6, 5, 5, 1, 129.99, 129.99, 0.00),
(7, 5, 8, 1, 199.99, 199.99, 0.00);

-- Update customer loyalty points based on orders
UPDATE customers
SET loyalty_points = loyalty_points + 50
WHERE customer_id IN (
    SELECT DISTINCT customer_id
    FROM orders
    WHERE status = 'DELIVERED'
);

-- Update product stock quantities based on orders
UPDATE products
SET stock_quantity = stock_quantity - (
    SELECT COALESCE(SUM(oi.quantity), 0)
    FROM order_items oi
    JOIN orders o ON oi.order_id = o.order_id
    WHERE oi.product_id = products.product_id
    AND o.status != 'CANCELLED'
);
