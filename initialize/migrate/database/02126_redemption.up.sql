-- Create redemption_code table
CREATE TABLE IF NOT EXISTS `redemption_code` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
    `code` VARCHAR(255) NOT NULL COMMENT 'Redemption Code',
    `total_count` BIGINT NOT NULL DEFAULT 0 COMMENT 'Total Redemption Count',
    `used_count` BIGINT NOT NULL DEFAULT 0 COMMENT 'Used Redemption Count',
    `subscribe_plan` BIGINT NOT NULL DEFAULT 0 COMMENT 'Subscribe Plan',
    `unit_time` VARCHAR(50) NOT NULL DEFAULT 'month' COMMENT 'Unit Time: day, month, quarter, half_year, year',
    `quantity` BIGINT NOT NULL DEFAULT 1 COMMENT 'Quantity',
    `created_at` DATETIME NOT NULL COMMENT 'Creation Time',
    `updated_at` DATETIME NOT NULL COMMENT 'Update Time',
    `deleted_at` DATETIME DEFAULT NULL COMMENT 'Deletion Time',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_code` (`code`),
    KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Redemption Code Table';

-- Create redemption_record table
CREATE TABLE IF NOT EXISTS `redemption_record` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
    `redemption_code_id` BIGINT NOT NULL DEFAULT 0 COMMENT 'Redemption Code Id',
    `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT 'User Id',
    `subscribe_id` BIGINT NOT NULL DEFAULT 0 COMMENT 'Subscribe Id',
    `unit_time` VARCHAR(50) NOT NULL DEFAULT 'month' COMMENT 'Unit Time',
    `quantity` BIGINT NOT NULL DEFAULT 1 COMMENT 'Quantity',
    `redeemed_at` DATETIME NOT NULL COMMENT 'Redeemed Time',
    `created_at` DATETIME NOT NULL COMMENT 'Creation Time',
    PRIMARY KEY (`id`),
    KEY `idx_redemption_code_id` (`redemption_code_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Redemption Record Table';
