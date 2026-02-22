INSERT INTO assets (code) VALUES 
('GOLD'),
('DIAMOND')
ON CONFLICT (id) DO NOTHING;

INSERT INTO users (id, name) VALUES
('e1e1e1e1-e1e1-e1e1-e1e1-e1e1e1e1e1e1', 'Eren Yeager'),
('e2e2e2e2-e2e2-e2e2-e2e2-e2e2e2e2e2e2', 'Mikasa Akerman')
ON CONFLICT (id) DO NOTHING;

INSERT INTO wallets (id, label, user_id, asset_type_id, balance) VALUES 
('00000000-0000-0000-0000-000000000000', 'Treasury Gold', NULL, 1, 9999000),
('00000000-0000-0000-0000-000000000001', 'Treasury Diamond', NULL, 2, 9999900),
('00000000-0000-0000-0000-000000000002', 'Revenue Diamond', NULL, 2, 10),
('00000000-0000-0000-0000-000000000003', 'Revenue Gold', NULL, 1, 0),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Mikasa Gold Wallet', 'e2e2e2e2-e2e2-e2e2-e2e2-e2e2e2e2e2e2', 1, 1000),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Eren Diamond Wallet', 'e1e1e1e1-e1e1-e1e1-e1e1-e1e1e1e1e1e1', 2, 90)
ON CONFLICT (id) DO NOTHING;

INSERT INTO transactions (id, reference_id, type, status) VALUES 
('d1111111-1111-1111-1111-111111111111', 'mikasa_buy_gold', 'purchase', 'completed'),
('d2222222-2222-2222-2222-222222222222', 'eren_buy_diamond', 'purchase', 'completed'),
('d3333333-3333-3333-3333-333333333333', 'eren_redeem_diamond', 'redemption', 'completed')
ON CONFLICT (id) DO NOTHING;

INSERT INTO ledger_entries (id, transaction_id, wallet_id, direction, amount) VALUES 
(gen_random_uuid(), 'd1111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000000', 'debit', 1000),
(gen_random_uuid(), 'd1111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'credit', 1000),

(gen_random_uuid(), 'd2222222-2222-2222-2222-222222222222', '00000000-0000-0000-0000-000000000001', 'debit', 100),
(gen_random_uuid(), 'd2222222-2222-2222-2222-222222222222', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'credit', 100),

(gen_random_uuid(), 'd3333333-3333-3333-3333-333333333333', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'debit', 10),
(gen_random_uuid(), 'd3333333-3333-3333-3333-333333333333', '00000000-0000-0000-0000-000000000002', 'credit', 10);