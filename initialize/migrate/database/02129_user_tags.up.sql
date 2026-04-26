-- 用户标签:管理员自由打标(VIP / 试用 / 长期客户 等)。逗号分隔多标签。
ALTER TABLE `user`
    ADD COLUMN `tags` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci
        NOT NULL DEFAULT '' COMMENT 'User Tags (comma-separated)';
