-- 000007: Revert — remove seeded vendors and loan officer
DELETE FROM vendors WHERE field_agent_id = 'b8f1077b-c186-40d1-8889-e3c10cad7fa8';
DELETE FROM users WHERE id IN (
    'aec6a745-4b99-485a-b71a-5577551198d7',
    '648774d1-3b5b-4326-a700-c14245736db1',
    '78de1c58-e7af-4e0c-a1d7-26af075c8d8b',
    '3d7966a4-cb35-4023-b89b-c7b69abcd8fc'
);
DELETE FROM users WHERE id = 'b8f1077b-c186-40d1-8889-e3c10cad7fa8';
