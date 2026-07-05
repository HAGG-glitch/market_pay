-- 000011: Whitelist USSD-registered vendors (Maurice, Sanah) so they can access the menu.

INSERT INTO ussd_allowed_subscribers (subscriber_id_hash, label, is_active)
VALUES
  ('be3a4ef5-f55e-5f78-b09c-a25dc69c5c5c', 'Maurice Bangura', TRUE),
  ('e6b9c01c-827d-5305-bcdb-fe935c7e7796', 'Maurice Bangura', TRUE),
  ('c08d98dc-4851-5ccb-8d13-75725ca49a17', 'Sanah Marah', TRUE)
ON CONFLICT (subscriber_id_hash)
DO UPDATE SET
    label     = EXCLUDED.label,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();
