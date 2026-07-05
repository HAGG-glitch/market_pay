-- 000014: Revert — set field_agent_id back to NULL for USSD-registered vendors.

UPDATE vendors
SET field_agent_id = NULL,
    updated_at = NOW()
WHERE field_agent_id = 'b8f1077b-c186-40d1-8889-e3c10cad7fa8'
  AND deleted_at IS NULL
  AND id IN (
    SELECT DISTINCT us.vendor_id
    FROM ussd_subscribers us
    WHERE us.vendor_id IS NOT NULL
  );
