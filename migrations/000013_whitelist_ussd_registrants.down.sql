-- 000013: Revert — remove Maurice and Sanah vendor records, users, and whitelist entries.

DELETE FROM credit_scores
WHERE vendor_id IN (
  SELECT id FROM vendors WHERE phone IN ('+23276889012', '+23276889013')
);

DELETE FROM ussd_subscribers
WHERE subscriber_id IN (
  'be3a4ef5-f55e-5f78-b09c-a25dc69c5c5c',
  'e6b9c01c-827d-5305-bcdb-fe935c7e7796',
  'c08d98dc-4851-5ccb-8d13-75725ca49a17'
);

DELETE FROM ussd_allowed_subscribers
WHERE subscriber_id_hash IN (
  'be3a4ef5-f55e-5f78-b09c-a25dc69c5c5c',
  'e6b9c01c-827d-5305-bcdb-fe935c7e7796',
  'c08d98dc-4851-5ccb-8d13-75725ca49a17'
);

DELETE FROM vendors WHERE phone IN ('+23276889012', '+23276889013');

DELETE FROM users WHERE id IN (
  'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
  'b2c3d4e5-f6a7-8901-bcde-f12345678901'
);
