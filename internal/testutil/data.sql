-- Create test database
-- CREATE DATABASE db_part_dump;

-- Connect to the test database
-- \c db_part_dump;

-- Set schema DROP SCHEMA alpha CASCADE; 
CREATE SCHEMA IF NOT EXISTS alpha;
SET search_path TO alpha;

-- Create tables with relationships
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'active'
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending'
);

CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL,
    product_name VARCHAR(100) NOT NULL,
    quantity INTEGER NOT NULL,
    price DECIMAL(10,2) NOT NULL
);

CREATE TABLE user_addresses (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    address_type VARCHAR(20) NOT NULL,
    street VARCHAR(100) NOT NULL,
    city VARCHAR(50) NOT NULL,
    country VARCHAR(50) NOT NULL
);

CREATE TABLE user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    theme VARCHAR(20) DEFAULT 'light',
    notifications_enabled BOOLEAN DEFAULT true,
    language VARCHAR(10) DEFAULT 'en'
);

-- Add this to your schema
CREATE TABLE user_payment_methods (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id INTEGER REFERENCES orders(id) ON DELETE SET NULL,
    payment_type VARCHAR(20) NOT NULL,
    card_number VARCHAR(20),
    expiry_date DATE,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraint to ensure either user_id or order_id is provided
    CONSTRAINT chk_user_or_order CHECK (
        (user_id IS NOT NULL) OR (order_id IS NOT NULL)
    )
);

-- Guid pk tables
CREATE TABLE coupons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    discount_percent DECIMAL(5,2) NOT NULL CHECK (discount_percent > 0 AND discount_percent <= 100),
    valid_from TIMESTAMP NOT NULL,
    valid_until TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE order_coupons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    coupon_id UUID NOT NULL REFERENCES coupons(id) ON DELETE CASCADE,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Cycle example
-- Step 1: Create empty tables without FKs first
CREATE TABLE table_one (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
    -- FK to table_two will be added later
);

CREATE TABLE table_two (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
    -- FK to table_three will be added later
);

CREATE TABLE table_three (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid()
    -- FK to table_one will be added later
);

-- Step 2: Add the foreign keys in a cycle
ALTER TABLE table_one
    ADD COLUMN two_id UUID,
    ADD CONSTRAINT fk_one_two FOREIGN KEY (two_id) REFERENCES table_two(id);

ALTER TABLE table_two
    ADD COLUMN three_id UUID,
    ADD CONSTRAINT fk_two_three FOREIGN KEY (three_id) REFERENCES table_three(id);

ALTER TABLE table_three
    ADD COLUMN one_id UUID,
    ADD CONSTRAINT fk_three_one FOREIGN KEY (one_id) REFERENCES table_one(id);


-- Insert users
INSERT INTO users (username, email, status, created_at) VALUES
('john_doe', 'john@example.com', 'active', '2025-01-01 10:00:00.928501'),
('jane_smith', 'jane@example.com', 'active', '2025-01-02 10:00:00.928502'),
('bob_wilson', 'bob@example.com', 'inactive', '2025-01-03 10:00:00'),
('sarah_jones', 'sarah@example.com', 'active', '2025-01-04 10:00:00'),
('mike_brown', 'mike@example.com', 'suspended', '2025-01-05 10:00:00');

-- Insert orders
INSERT INTO orders (user_id, total_amount, status, order_date) VALUES
(1, 99.99, 'completed', '2025-01-01 10:00:00'),
(1, 49.99, 'pending', '2025-01-01 10:00:00'),
(2, 199.99, 'completed', '2025-01-01 10:00:00'),
(3, 29.99, 'cancelled', '2025-01-01 10:00:00'),
(4, 149.99, 'completed', '2025-01-01 10:00:00'),
(4, 79.99, 'processing', '2025-01-01 10:00:00'),
(4, 199.99, 'completed', '2025-01-01 10:00:00');

-- Insert order items
INSERT INTO order_items (order_id, product_id, product_name, quantity, price) VALUES
(1, 101, 'Laptop', 1, 89.99),
(1, 102, 'Mouse', 1, 10.00),
(2, 103, 'Keyboard', 1, 49.99),
(3, 104, 'Monitor', 1, 199.99),
(4, 105, 'Headphones', 1, 29.99),
(5, 106, 'Tablet', 1, 149.99),
(6, 107, 'Phone Case', 2, 39.99),
(7, 108, 'Smartwatch', 1, 199.99);

-- Insert user addresses
INSERT INTO user_addresses (user_id, address_type, street, city, country) VALUES
(1, 'home', '123 Main St', 'New York', 'USA'),
(1, 'work', '456 Office Rd', 'New York', 'USA'),
(2, 'home', '789 Park Ave', 'Boston', 'USA'),
(3, 'home', '321 Oak St', 'Chicago', 'USA'),
(4, 'home', '654 Pine Rd', 'Seattle', 'USA'),
(4, 'vacation', '987 Beach Blvd', 'Miami', 'USA');

-- Insert user preferences
INSERT INTO user_preferences (user_id, theme, notifications_enabled, language) VALUES
(1, 'dark', true, 'en'),
(2, 'light', false, 'en'),
(3, 'dark', true, 'es'),
(4, 'light', true, 'fr'),
(5, 'dark', false, 'de');

-- Insert payment methods with both FKs populated
INSERT INTO user_payment_methods (user_id, order_id, payment_type, card_number, expiry_date, is_default, created_at) VALUES
(1, 1, 'credit_card', '4111111111111111', '2025-12-01', true,  '2025-01-01 10:00:00'),
(1, NULL, 'paypal', NULL, NULL, false,                       '2025-01-02 11:00:00'),
(2, 3, 'credit_card', '4222222222222222', '2024-10-01', true, '2025-01-03 12:00:00'),
(3, 4, 'debit_card', '4333333333333333', '2023-08-01', true,  '2025-01-04 13:00:00'),
(4, 5, 'credit_card', '4444444444444444', '2026-05-01', true, '2025-01-05 14:00:00'),
(4, 6, 'apple_pay', NULL, NULL, false,                      '2025-01-06 15:00:00'),
(4, NULL, 'google_pay', NULL, NULL, false,                  '2025-01-07 16:00:00');
-- Some with only user_id (no order_id)
INSERT INTO user_payment_methods (user_id, payment_type, card_number, expiry_date, is_default, created_at) VALUES
(1, 'bank_transfer', NULL, NULL, false, '2025-01-08 17:00:00'),
(2, 'credit_card', '4555555555555555', '2024-12-01', false, '2025-01-09 18:00:00');


INSERT INTO coupons (id, code, discount_percent, valid_from, valid_until)
VALUES
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'WELCOME10', 10.00, '2025-01-01', '2025-12-31'),
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'SPRING20', 20.00, '2025-03-01', '2025-06-01'),
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'SUMMER15', 15.00, '2025-06-01', '2025-09-01'),
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14', 'BLACKFRIDAY50', 50.00, '2025-11-25', '2025-11-30'),
('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a15', 'NEWYEAR25', 25.00, '2025-12-30', '2026-01-05');

-- Link orders with coupons
INSERT INTO order_coupons (order_id, coupon_id)
SELECT 1, id FROM coupons WHERE code = 'WELCOME10';
INSERT INTO order_coupons (order_id, coupon_id)
SELECT 2, id FROM coupons WHERE code = 'SPRING20';
INSERT INTO order_coupons (order_id, coupon_id)
SELECT 3, id FROM coupons WHERE code = 'SUMMER15';
INSERT INTO order_coupons (order_id, coupon_id)
SELECT 5, id FROM coupons WHERE code = 'BLACKFRIDAY50';
INSERT INTO order_coupons (order_id, coupon_id)
SELECT 7, id FROM coupons WHERE code = 'NEWYEAR25';


-- Insert initial
INSERT INTO table_one (id) VALUES ('11111111-1111-1111-1111-111111111111');
INSERT INTO table_two (id) VALUES ('22222222-2222-2222-2222-222222222222');
INSERT INTO table_three (id) VALUES ('33333333-3333-3333-3333-333333333333');

-- Update to form the cycle
UPDATE table_one
SET two_id = '22222222-2222-2222-2222-222222222222'
WHERE id = '11111111-1111-1111-1111-111111111111';
UPDATE table_two
SET three_id = '33333333-3333-3333-3333-333333333333'
WHERE id = '22222222-2222-2222-2222-222222222222';
UPDATE table_three
SET one_id = '11111111-1111-1111-1111-111111111111'
WHERE id = '33333333-3333-3333-3333-333333333333';

