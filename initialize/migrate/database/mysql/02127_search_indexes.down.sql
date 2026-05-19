ALTER TABLE `nodes`
    DROP INDEX `idx_nodes_port`,
    DROP INDEX `idx_nodes_tags`,
    DROP INDEX `idx_nodes_address`,
    DROP INDEX `idx_nodes_name`;

ALTER TABLE `servers`
    DROP INDEX `idx_servers_address`,
    DROP INDEX `idx_servers_name`;

ALTER TABLE `payment`
    DROP INDEX `idx_payment_name`;

ALTER TABLE `coupon`
    DROP INDEX `idx_coupon_name`;

ALTER TABLE `user`
    DROP INDEX `idx_user_refer_code`;

ALTER TABLE `order`
    DROP INDEX `idx_order_coupon`,
    DROP INDEX `idx_order_trade_no`;
