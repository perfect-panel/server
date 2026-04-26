-- V4.3 决策 25:客户端教程 CMS。
-- subscribe_application 表加 tutorial_key,关联 site_content.content_key
-- (例如 'client_tutorial_shadowrocket'),让每个客户端可以挂一份多语言教程。
ALTER TABLE `subscribe_application`
    ADD COLUMN `tutorial_key` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci
        NOT NULL DEFAULT '' COMMENT 'site_content.content_key (multi-lang tutorial)';
