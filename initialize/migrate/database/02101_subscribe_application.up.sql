DROP TABLE IF EXISTS `subscribe_application`;
CREATE TABLE `subscribe_application` (
     `id` bigint NOT NULL AUTO_INCREMENT,
     `name` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'Application Name',
     `icon` mediumtext COLLATE utf8mb4_general_ci COMMENT 'Application Icon',
     `description` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT 'Application Description',
     `scheme` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'Application Scheme',
     `user_agent` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'User Agent',
     `is_default` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Is Default Application',
     `proxy_template` mediumtext COLLATE utf8mb4_general_ci COMMENT 'Proxy Template',
     `subscribe_template` mediumtext COLLATE utf8mb4_general_ci COMMENT 'Subscribe Template',
     `output_format` varchar(50) COLLATE utf8mb4_general_ci NOT NULL DEFAULT 'yaml' COMMENT 'Output Format',
     `download_link` text COLLATE utf8mb4_general_ci NOT NULL COMMENT 'Download Link',
     `created_at` datetime(3) DEFAULT NULL COMMENT 'Create Time',
     `updated_at` datetime(3) DEFAULT NULL COMMENT 'Update Time',
     PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
