-- START WITH THESE QUERIES

-- 1. See all your data in a readable format
SELECT
    tenant_id,
    enabled,
    JSON_PRETTY(email_providers) as email_config,
    JSON_PRETTY(batch_config) as batch_settings
FROM partner_configs;

-- 2. Quick overview of notifications
SELECT
    tenant_id,
    type,
    status,
    COUNT(*) as count
FROM notifications
GROUP BY tenant_id, type, status
ORDER BY tenant_id, type, status;

-- 3. Recent notifications with details
SELECT
    tenant_id,
    type,
    status,
    headline,
    LEFT(address, 30) as address,
    create_time
FROM notifications
ORDER BY create_time DESC
    LIMIT 10;

-- 4. Extract email addresses from JSON config
SELECT
    tenant_id,
    JSON_EXTRACT(email_providers, '$[0].config.MSGBonusFrom') as bonus_email,
    JSON_EXTRACT(email_providers, '$[0].config.MSGPromoFrom') as promo_email,
    JSON_EXTRACT(email_providers, '$[0].config.MSGSystemFrom') as system_email
FROM partner_configs
WHERE tenant_id = 1001;