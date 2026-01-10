-- Add short_code column to user_device table
ALTER TABLE `user_device` ADD COLUMN `short_code` VARCHAR(255) DEFAULT '' COMMENT 'Short Code' AFTER `identifier`;
