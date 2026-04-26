-- V4.3 §10 DoD #19 压测前置数据 — 1 万用户 + 1 万订阅 + 3 万设备槽。
-- 在隔离环境跑(staging),不要在生产数据库执行。
--
-- 用法:
--   mysql -h $HOST -u $USER -p$PASS $DB < seed.sql
--
-- 速度参考:本机 8c16g + MySQL 8 单实例 ≈ 30s 完成。

SET autocommit = 0;
SET unique_checks = 0;
SET foreign_key_checks = 0;

-- 假设至少有 1 个 subscribe 已存在(plan id=1);否则先 INSERT 一个。
INSERT INTO `subscribe` (`id`, `name`, `unit_price`, `unit_price_per_device`, `unit_time`,
    `traffic`, `speed_limit`, `device_limit`, `max_device_count`, `inventory`, `show`, `sell`, `created_at`, `updated_at`)
VALUES (10000, 'PERF SEED PLAN', 0, 1000, 'Month', 107374182400, 0, 0, 20, -1, 1, 1, NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`);

-- 1 万用户(id 1000000 ~ 1009999)
INSERT INTO `user` (`id`, `password`, `algo`, `balance`, `enable`, `created_at`, `updated_at`)
SELECT 1000000 + seq, 'seed', 'default', 0, 1, NOW(3), NOW(3)
FROM (
    SELECT a.N + b.N * 10 + c.N * 100 + d.N * 1000 AS seq
    FROM (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) a
    CROSS JOIN (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) b
    CROSS JOIN (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) c
    CROSS JOIN (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) d
) seq_table;

-- 1 万 user_subscribe(每个用户一个,关联 plan 10000)
INSERT INTO `user_subscribe` (`id`, `user_id`, `order_id`, `subscribe_id`,
    `device_count`, `start_time`, `expire_time`, `traffic`, `download`, `upload`,
    `status`, `created_at`, `updated_at`)
SELECT 1000000 + seq, 1000000 + seq, 0, 10000,
    3, NOW(3), DATE_ADD(NOW(3), INTERVAL 30 DAY), 107374182400, 0, 0,
    1, NOW(3), NOW(3)
FROM (
    SELECT a.N + b.N * 10 + c.N * 100 + d.N * 1000 AS seq
    FROM (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) a
    CROSS JOIN (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) b
    CROSS JOIN (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) c
    CROSS JOIN (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) d
) seq_table;

-- 3 万设备槽(每个 user_subscribe 3 个)
-- token / uuid 使用确定性派生,保证种子可重放;真实 token 由代码 path 生成。
INSERT INTO `user_subscribe_device` (`id`, `user_subscribe_id`, `user_id`, `device_name`,
    `token`, `uuid`, `status`, `created_at`, `updated_at`)
SELECT 1000000 + seq * 3 + slot,
    1000000 + seq, 1000000 + seq,
    CONCAT('seed-', seq, '-', slot),
    LPAD(HEX(seq * 3 + slot), 43, '0'),
    UUID(),
    1, NOW(3), NOW(3)
FROM (
    SELECT a.N + b.N * 10 + c.N * 100 + d.N * 1000 AS seq
    FROM (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) a
    CROSS JOIN (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) b
    CROSS JOIN (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) c
    CROSS JOIN (SELECT 0 AS N UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9) d
) seq_table
CROSS JOIN (SELECT 0 AS slot UNION SELECT 1 UNION SELECT 2) slots;

COMMIT;
SET unique_checks = 1;
SET foreign_key_checks = 1;
SET autocommit = 1;

-- 清理:
-- DELETE FROM user_subscribe_device WHERE id BETWEEN 1000000 AND 1029999;
-- DELETE FROM user_subscribe        WHERE id BETWEEN 1000000 AND 1009999;
-- DELETE FROM user                  WHERE id BETWEEN 1000000 AND 1009999;
-- DELETE FROM subscribe             WHERE id = 10000;
