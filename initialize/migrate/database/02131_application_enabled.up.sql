-- 客户端启用开关:管理员可关闭某个客户端,关闭后用户端不再显示该客户端,
-- 但管理端和 UA 自动匹配仍可命中(避免误关导致老用户拉不到订阅)。
ALTER TABLE `subscribe_application`
    ADD COLUMN `enabled` TINYINT(1) NOT NULL DEFAULT 1
        COMMENT 'Visible on user-facing client list (1=show, 0=hide)';
