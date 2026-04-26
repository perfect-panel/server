-- V4.4 #45 用户协议版本管理。
-- 管理员在 site_content 修改 terms_of_use 时手工 bump version;
-- 用户已接受版本与当前版本不一致时强制重新接受。

ALTER TABLE `site_content`
    ADD COLUMN `version` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci
        NOT NULL DEFAULT '1' COMMENT 'Content version, bump on meaningful change';

ALTER TABLE `user`
    ADD COLUMN `terms_version` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci
        NOT NULL DEFAULT '' COMMENT 'Last accepted terms_of_use version';
