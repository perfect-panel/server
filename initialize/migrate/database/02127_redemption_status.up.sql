-- Add status column to redemption_code table
ALTER TABLE `redemption_code` ADD COLUMN `status` TINYINT NOT NULL DEFAULT 1 COMMENT 'Status: 1=enabled, 0=disabled' AFTER `quantity`;
