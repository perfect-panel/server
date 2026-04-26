-- V4.3 站内信表(决策 7.1 通知矩阵 inMail 渠道)。
-- 由 noticeDispatchLogic 在派发邮件时同步写一行到此表;
-- 用户中心通过 GET /v1/portal/messages 拉取 + PUT /:id/read 标记已读。

CREATE TABLE IF NOT EXISTS `user_message`
(
    `id`              bigint                                                        NOT NULL AUTO_INCREMENT,
    `user_id`         bigint                                                        NOT NULL COMMENT 'Recipient user id',
    `category`        varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci  NOT NULL DEFAULT '' COMMENT 'notice template key, e.g. throttle_started / device_reset',
    `title`           varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'Pre-rendered title',
    `body`            mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT 'Pre-rendered body',
    `link`            varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'Optional deep link (e.g. /device-billing)',
    `read_at`         datetime(3)                                                            DEFAULT NULL COMMENT 'Read time, NULL = unread',
    `created_at`      datetime(3)                                                            DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_user_unread` (`user_id`, `read_at`),
    KEY `idx_created_at` (`created_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;
