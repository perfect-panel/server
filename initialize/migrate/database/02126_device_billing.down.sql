-- Rollback for V4.3 device-billing schema

ALTER TABLE `order`
    DROP COLUMN `device_count`;

ALTER TABLE `servers`
    DROP COLUMN `direct_list`;

DROP TABLE IF EXISTS `site_content`;
DROP TABLE IF EXISTS `audit_log`;
DROP TABLE IF EXISTS `traffic_addon_order`;
DROP TABLE IF EXISTS `user_subscribe_device`;

ALTER TABLE `user_subscribe`
    DROP COLUMN `notified_24h`,
    DROP COLUMN `notified_12h`,
    DROP COLUMN `notified_100`,
    DROP COLUMN `notified_90`,
    DROP COLUMN `cut_off_at`,
    DROP COLUMN `throttled_at`,
    DROP COLUMN `traffic_addon`,
    DROP COLUMN `device_count`,
    ADD UNIQUE KEY `uni_user_subscribe_token` (`token`),
    ADD UNIQUE KEY `uni_user_subscribe_uuid` (`uuid`);

ALTER TABLE `subscribe`
    DROP COLUMN `commission_rate`,
    DROP COLUMN `traffic_addon_unit_size`,
    DROP COLUMN `traffic_addon_unit_price`,
    DROP COLUMN `max_device_count`,
    DROP COLUMN `unit_price_per_device`;
