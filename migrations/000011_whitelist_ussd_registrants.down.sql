-- 000011: Revert — remove whitelisted USSD registrants.

DELETE FROM ussd_allowed_subscribers
WHERE subscriber_id_hash IN (
  'be3a4ef5-f55e-5f78-b09c-a25dc69c5c5c',
  'e6b9c01c-827d-5305-bcdb-fe935c7e7796',
  'c08d98dc-4851-5ccb-8d13-75725ca49a17'
);
