ALTER TABLE `user_subscribe_device`
    ADD COLUMN `is_addon` TINYINT(1) NOT NULL DEFAULT 0
        COMMENT 'V4.3 1=user-purchased addon (deletable) 0=plan-included base';
