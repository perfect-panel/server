ALTER TABLE `User`
    ADD COLUMN `algo` VARCHAR(20) NOT NULL DEFAULT 'default' COMMENT 'Encryption Algorithm' AFTER `password`,
    ADD COLUMN `salt` VARCHAR(20) NOT NULL DEFAULT 'default' COMMENT 'Password Salt' AFTER `algo`;
