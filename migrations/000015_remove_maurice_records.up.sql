-- 000015: Remove Maurice Bangura's records so he can re-register fresh.

DO $$
DECLARE
    maurice_phone TEXT := '+232099618027';
    v_user_id UUID;
    v_vendor_id UUID;
BEGIN
    -- Find the user and vendor IDs
    SELECT id INTO v_user_id FROM users WHERE phone = maurice_phone;
    SELECT id INTO v_vendor_id FROM vendors WHERE phone = maurice_phone;

    -- Clean up loan-related records first
    DELETE FROM repayment_schedules WHERE loan_id IN (SELECT id FROM loans WHERE vendor_id = v_vendor_id);
    DELETE FROM loan_repayments WHERE vendor_id = v_vendor_id;
    DELETE FROM loans WHERE vendor_id = v_vendor_id;

    -- Clean up vendor-related records
    DELETE FROM payments WHERE vendor_id = v_vendor_id;
    DELETE FROM group_members WHERE vendor_id = v_vendor_id;
    DELETE FROM credit_scores WHERE vendor_id = v_vendor_id;
    DELETE FROM ussd_subscribers WHERE vendor_id = v_vendor_id;

    -- Remove from whitelist
    DELETE FROM ussd_allowed_subscribers
    WHERE subscriber_id_hash IN (
        'be3a4ef5-f55e-5f78-b09c-a25dc69c5c5c',
        'e6b9c01c-827d-5305-bcdb-fe935c7e7796'
    );

    -- Remove vendor and user
    DELETE FROM vendors WHERE id = v_vendor_id;
    DELETE FROM users WHERE id = v_user_id;
END $$;
