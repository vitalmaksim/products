BEGIN;

-- Users
INSERT INTO users (email, name, balance)
VALUES
    ('max@example.com', 'Max', 100000),
    ('ivan@example.com', 'Ivan', 0),
    ('alex@example.com', 'Alex', 300000)
ON CONFLICT (email) DO NOTHING;

-- Products
INSERT INTO products (external_id, name, description, price)
VALUES
    (1001, 'Premium Subscription', 'Premium subscription for 1 month', 15000),
    (1002, 'Game License', 'Digital game license', 30000),
    (1003, 'Cloud Storage', '100 GB cloud storage', 5000),
    (1004, 'Music Subscription', 'Music subscription for 1 year', 120000),
    (1005, 'Gift Card', 'Digital gift card', 50000)
ON CONFLICT (external_id) DO NOTHING;

COMMIT;

