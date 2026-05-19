ALTER TABLE `order`
    ADD INDEX `idx_order_trade_no` (`trade_no`),
    ADD INDEX `idx_order_coupon` (`coupon`);

ALTER TABLE `user`
    ADD INDEX `idx_user_refer_code` (`refer_code`);

ALTER TABLE `coupon`
    ADD INDEX `idx_coupon_name` (`name`);

ALTER TABLE `payment`
    ADD INDEX `idx_payment_name` (`name`);

ALTER TABLE `servers`
    ADD INDEX `idx_servers_name` (`name`),
    ADD INDEX `idx_servers_address` (`address`);

ALTER TABLE `nodes`
    ADD INDEX `idx_nodes_name` (`name`),
    ADD INDEX `idx_nodes_address` (`address`),
    ADD INDEX `idx_nodes_tags` (`tags`),
    ADD INDEX `idx_nodes_port` (`port`);
