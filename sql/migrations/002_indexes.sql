-- Index for user lookups by email
CREATE INDEX idx_users_email ON users(email);

-- Index for category lookups by user
CREATE INDEX idx_categories_user_id ON categories(user_id);

-- Index for transaction lookups by user and date range
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_occurred_at ON transactions(occurred_at DESC);
CREATE INDEX idx_transactions_user_date ON transactions(user_id, occurred_at DESC);

-- Index for summary queries
CREATE INDEX idx_transactions_category ON transactions(category_id);
