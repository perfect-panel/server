ALTER TABLE `User`
    ADD COLUMN IF NOT EXISTS `algo` VARCHAR(20) NOT NULL DEFAULT 'default' COMMENT 'Encryption Algorithm' AFTER `password`,
    ADD COLUMN IF NOT EXISTS `salt` VARCHAR(20) NOT NULL DEFAULT 'default' COMMENT 'Password Salt' AFTER `algo`;
