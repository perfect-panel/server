-- Add index on refer_code column for faster lookup
ALTER TABLE `user` ADD INDEX `idx_refer_code` (`refer_code`);
