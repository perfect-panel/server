-- V4.3 设备数计费 + 共享流量池模型
-- Phase 1: schema 变更(决策 1-40 中 schema 相关项)

-- 1) subscribe 表扩展(决策 1, 4, 7, 10, 21)
ALTER TABLE `subscribe`
    ADD COLUMN `unit_price_per_device`    bigint NOT NULL DEFAULT 0  COMMENT 'Per-device unit price (cents)' AFTER `unit_price`,
    ADD COLUMN `max_device_count`         bigint NOT NULL DEFAULT 20 COMMENT 'Max device slots a user can buy' AFTER `device_limit`,
    ADD COLUMN `traffic_addon_unit_price` bigint NOT NULL DEFAULT 0  COMMENT 'Traffic addon unit price (cents per unit)' AFTER `traffic`,
    ADD COLUMN `traffic_addon_unit_size`  bigint NOT NULL DEFAULT 1073741824 COMMENT 'Traffic addon step size in bytes (default 1GiB)' AFTER `traffic_addon_unit_price`,
    ADD COLUMN `commission_rate`          bigint NOT NULL DEFAULT 10 COMMENT 'Invite commission percent (0-100)' AFTER `deduction_ratio`;

-- 2) user_subscribe 表扩展(决策 2, 6, 20, 40)
-- 旧 token / uuid 字段保留但不再使用(决策 15:新机场,无需迁移),
-- 必须先 drop 这两个 UNIQUE 索引,否则多行同时为空会冲突。
ALTER TABLE `user_subscribe`
    DROP INDEX `uni_user_subscribe_token`,
    DROP INDEX `uni_user_subscribe_uuid`,
    ADD COLUMN `device_count`  bigint      NOT NULL DEFAULT 1 COMMENT 'Allowed device slots' AFTER `subscribe_id`,
    ADD COLUMN `traffic_addon` bigint      NOT NULL DEFAULT 0 COMMENT 'Addon traffic bytes (added to base quota)' AFTER `traffic`,
    ADD COLUMN `throttled_at`  datetime(3) NULL DEFAULT NULL  COMMENT 'Throttle start time (1Mbps after over-quota)' AFTER `status`,
    ADD COLUMN `cut_off_at`    datetime(3) NULL DEFAULT NULL  COMMENT 'Cut-off time = throttled_at + 24h' AFTER `throttled_at`,
    ADD COLUMN `notified_90`   tinyint(1)  NOT NULL DEFAULT 0 COMMENT 'Sent 90% warning' AFTER `cut_off_at`,
    ADD COLUMN `notified_100`  tinyint(1)  NOT NULL DEFAULT 0 COMMENT 'Sent throttle-start notice' AFTER `notified_90`,
    ADD COLUMN `notified_12h`  tinyint(1)  NOT NULL DEFAULT 0 COMMENT 'Sent 12h-throttle reminder' AFTER `notified_100`,
    ADD COLUMN `notified_24h`  tinyint(1)  NOT NULL DEFAULT 0 COMMENT 'Sent cut-off notice' AFTER `notified_12h`;

-- 3) user_subscribe_device 设备槽表(决策 3, 9, 34, 36)
CREATE TABLE IF NOT EXISTS `user_subscribe_device`
(
    `id`                bigint                                                        NOT NULL AUTO_INCREMENT,
    `user_subscribe_id` bigint                                                        NOT NULL COMMENT 'User subscribe id',
    `user_id`           bigint                                                        NOT NULL COMMENT 'Denormalized user id (for index)',
    `device_name`       varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci  NOT NULL DEFAULT '' COMMENT 'User-friendly device name',
    `token`             varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci  NOT NULL DEFAULT '' COMMENT '32-byte base64url token',
    `uuid`              char(36)                                                      NOT NULL DEFAULT '' COMMENT 'Device uuid (proto-agnostic identity)',
    `last_seen_ip`      varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci  NOT NULL DEFAULT '' COMMENT 'Most recent client IP',
    `last_seen_at`      datetime(3)                                                            DEFAULT NULL COMMENT 'Most recent online time',
    `today_traffic`     bigint                                                        NOT NULL DEFAULT 0 COMMENT 'Today traffic bytes',
    `reset_count_hour`  smallint                                                      NOT NULL DEFAULT 0 COMMENT 'Reset count in current hour',
    `reset_count_day`   smallint                                                      NOT NULL DEFAULT 0 COMMENT 'Reset count in current day',
    `last_reset_at`     datetime(3)                                                            DEFAULT NULL COMMENT 'Last reset time',
    `status`            tinyint(1)                                                    NOT NULL DEFAULT 1 COMMENT '0:disabled 1:enabled',
    `created_at`        datetime(3)                                                            DEFAULT NULL COMMENT 'Creation time',
    `updated_at`        datetime(3)                                                            DEFAULT NULL COMMENT 'Update time',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_user_subscribe_device_token` (`token`),
    UNIQUE KEY `uni_user_subscribe_device_uuid` (`uuid`),
    KEY `idx_user_subscribe_id` (`user_subscribe_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

-- 4) traffic_addon_order 流量加购订单(决策 7)
CREATE TABLE IF NOT EXISTS `traffic_addon_order`
(
    `id`                bigint      NOT NULL AUTO_INCREMENT,
    `user_id`           bigint      NOT NULL COMMENT 'User id',
    `user_subscribe_id` bigint      NOT NULL COMMENT 'Target user subscribe',
    `order_id`          bigint      NOT NULL DEFAULT 0 COMMENT 'Order id (linked to order table if any)',
    `addon_bytes`       bigint      NOT NULL COMMENT 'Bytes added to pool',
    `amount`            bigint      NOT NULL COMMENT 'Paid amount in cents',
    `unit_price`        bigint      NOT NULL DEFAULT 0 COMMENT 'Unit price snapshot at purchase',
    `unit_size`         bigint      NOT NULL DEFAULT 0 COMMENT 'Unit size snapshot at purchase',
    `created_at`        datetime(3) DEFAULT NULL COMMENT 'Creation time',
    `updated_at`        datetime(3) DEFAULT NULL COMMENT 'Update time',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_user_subscribe_id` (`user_subscribe_id`),
    KEY `idx_order_id` (`order_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

-- 5) audit_log 审计日志(决策 13, 35)
CREATE TABLE IF NOT EXISTS `audit_log`
(
    `id`         bigint                                                        NOT NULL AUTO_INCREMENT,
    `user_id`    bigint                                                        NOT NULL DEFAULT 0 COMMENT 'Subject user id (0 if N/A)',
    `actor`      varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci  NOT NULL DEFAULT '' COMMENT 'user / admin / system',
    `actor_id`   bigint                                                        NOT NULL DEFAULT 0 COMMENT 'Actor user/admin id',
    `action`     varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci  NOT NULL DEFAULT '' COMMENT 'Action key (purchase, reset_device, ...)',
    `target`     varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'Target ref (e.g. device:1001)',
    `detail`     text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT 'Detail JSON',
    `client_ip`  varchar(45) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci  NOT NULL DEFAULT '' COMMENT 'Client IP',
    `created_at` datetime(3)                                                            DEFAULT NULL COMMENT 'Creation time',
    PRIMARY KEY (`id`),
    KEY `idx_user_action` (`user_id`, `action`),
    KEY `idx_actor` (`actor`, `actor_id`),
    KEY `idx_created_at` (`created_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

-- 6) site_content 站内 CMS:用户协议 + 客户端教程(决策 19, 25)
CREATE TABLE IF NOT EXISTS `site_content`
(
    `id`           bigint                                                        NOT NULL AUTO_INCREMENT,
    `content_key`  varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci  NOT NULL DEFAULT '' COMMENT 'Content key',
    `content_lang` varchar(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci   NOT NULL DEFAULT 'zh-CN' COMMENT 'Language',
    `title`        varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'Title',
    `body`         mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT 'Body (markdown or html)',
    `created_at`   datetime(3) DEFAULT NULL COMMENT 'Creation time',
    `updated_at`   datetime(3) DEFAULT NULL COMMENT 'Update time',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uni_site_content_key_lang` (`content_key`, `content_lang`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

INSERT IGNORE INTO `site_content` (`content_key`, `content_lang`, `title`, `body`, `created_at`, `updated_at`) VALUES
    ('terms_of_use',                  'zh-CN', '用户协议',                       '<管理员后续编辑>', NOW(3), NOW(3)),
    ('client_tutorial_v2rayn',        'zh-CN', 'v2rayN 导入教程',                '<管理员后续编辑>', NOW(3), NOW(3)),
    ('client_tutorial_clash',         'zh-CN', 'Clash 导入教程',                 '<管理员后续编辑>', NOW(3), NOW(3)),
    ('client_tutorial_stash',         'zh-CN', 'Stash 导入教程',                 '<管理员后续编辑>', NOW(3), NOW(3)),
    ('client_tutorial_shadowrocket',  'zh-CN', 'Shadowrocket 导入教程',          '<管理员后续编辑>', NOW(3), NOW(3)),
    ('client_tutorial_hiddify',       'zh-CN', 'Hiddify 导入教程',               '<管理员后续编辑>', NOW(3), NOW(3)),
    ('client_tutorial_clashmeta',     'zh-CN', 'Clash Meta for Android 导入教程', '<管理员后续编辑>', NOW(3), NOW(3)),
    ('client_tutorial_quantumult',    'zh-CN', 'Quantumult X 导入教程',          '<管理员后续编辑>', NOW(3), NOW(3)),
    ('client_tutorial_loon',          'zh-CN', 'Loon 导入教程',                  '<管理员后续编辑>', NOW(3), NOW(3)),
    ('client_tutorial_flclash',       'zh-CN', 'FlClash 导入教程',               '<管理员后续编辑>', NOW(3), NOW(3)),
    ('client_tutorial_surge',         'zh-CN', 'Surge 导入教程',                 '<管理员后续编辑>', NOW(3), NOW(3)),
    ('client_tutorial_surge_mac',     'zh-CN', 'Surge for Mac 导入教程',         '<管理员后续编辑>', NOW(3), NOW(3));

-- V4.3 通知模板(决策 20 + 7.1 通知矩阵)。模板支持 Go text/template,可用变量见 dispatcher 注释。
INSERT IGNORE INTO `site_content` (`content_key`, `content_lang`, `title`, `body`, `created_at`, `updated_at`) VALUES
    ('notice_traffic_90',       'zh-CN', '【流量预警】您的套餐已使用 90%',           '尊敬的用户:\n\n您的套餐 {{.SubscribeName}} 已使用 90%。剩余流量 {{.RemainHuman}}。\n如需补充,请到「我的订阅」加购流量包。\n\n— {{.SiteName}}', NOW(3), NOW(3)),
    ('notice_throttle_started', 'zh-CN', '【已限速】套餐流量已用完',                 '您的套餐 {{.SubscribeName}} 流量已耗尽,已限速到 1 Mbps。\n24 小时后将完全断网。\n请尽快「加购流量包」或「续费」恢复正常速率。\n\n— {{.SiteName}}', NOW(3), NOW(3)),
    ('notice_throttle_12h',     'zh-CN', '【限速倒计时】距离断网 12 小时',           '提醒:您的套餐 {{.SubscribeName}} 已限速 12 小时。\n再过 12 小时将完全断网。\n请尽快加购流量包或续费。\n\n— {{.SiteName}}', NOW(3), NOW(3)),
    ('notice_cutoff',           'zh-CN', '【已断网】套餐已停止服务',                 '您的套餐 {{.SubscribeName}} 因流量耗尽超过 24 小时已断网。\n续费或加购流量包后立即恢复。\n\n— {{.SiteName}}', NOW(3), NOW(3)),
    ('notice_traffic_restored', 'zh-CN', '【已恢复】流量加购完成',                   '您已成功加购 {{.AddonHuman}} 流量,套餐已恢复正常速率。\n\n— {{.SiteName}}', NOW(3), NOW(3)),
    ('notice_device_reset',     'zh-CN', '【设备已重置】订阅 URL 已更新',            '您的设备 {{.DeviceName}} 已重置,旧订阅 URL 不再可用。\n请在「我的订阅」复制新 URL 并重新导入客户端。\n\n— {{.SiteName}}', NOW(3), NOW(3)),
    ('notice_payment_success',  'zh-CN', '【扣款成功】',                            '您的订单 {{.OrderNo}} 已成功扣款 {{.AmountHuman}}。\n\n— {{.SiteName}}', NOW(3), NOW(3)),
    ('notice_admin_login_remote', 'zh-CN', '【异地登录】管理后台',                   '检测到管理后台异地登录:\n时间: {{.Timestamp}}\n位置: {{.Location}}\n如非本人操作,请立即修改密码。\n\n— {{.SiteName}}', NOW(3), NOW(3)),
    ('notice_expire_3d',        'zh-CN', '【套餐到期前 3 天】',                      '您的套餐 {{.SubscribeName}} 将于 3 天后到期({{.ExpireDate}}),请尽快续费。\n\n— {{.SiteName}}', NOW(3), NOW(3)),
    ('notice_expire_1d',        'zh-CN', '【套餐到期前 1 天】',                      '您的套餐 {{.SubscribeName}} 将于明天到期({{.ExpireDate}}),请尽快续费。\n\n— {{.SiteName}}', NOW(3), NOW(3));

-- 7) servers 表:直连白名单(决策 39)
-- 注:真实表名是 `servers`(02105 重建,02122 删掉了旧的 `server`)。
-- 加在表尾,不指定 AFTER(02105 schema 没有 `config` 列)。
ALTER TABLE `servers`
    ADD COLUMN `direct_list` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT 'JSON array of DIRECT host suffixes';

-- 8) order 表:device_count 记录购买/加购时的槽位数量
ALTER TABLE `order`
    ADD COLUMN `device_count` int NOT NULL DEFAULT 0 COMMENT 'V4.3 device-billing: device slots bought (0 = legacy time-billing)' AFTER `quantity`;
