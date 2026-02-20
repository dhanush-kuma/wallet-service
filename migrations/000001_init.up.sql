CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS assets (
    id SERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY,
    label TEXT,
    user_id UUID NULL REFERENCES users(id),
    asset_type_id INT NOT NULL REFERENCES assets(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    reference_id TEXT UNIQUE NOT NULL,
    type TEXT,
    status TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TYPE entry_direction AS ENUM ('credit', 'debit');

CREATE TABLE IF NOT EXISTS ledger_entries (
    id UUID PRIMARY KEY,
    transaction_id UUID Not NULL REFERENCES transactions(id),
    wallet_id UUID Not NULL REFERENCES wallets(id),
    direction entry_direction NOT NULL,
    amount BIGINT CHECK (amount > 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_wallet_direction ON ledger_entries(wallet_id, direction);
CREATE INDEX idx_ledger_transaction_id ON ledger_entries(transaction_id);
CREATE INDEX idx_wallets_user_id ON wallets(user_id);