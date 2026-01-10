-- Add device limit configuration to system table
INSERT IGNORE INTO `system` (`id`, `category`, `key`, `value`, `type`, `desc`, `created_at`, `updated_at`)
VALUES (43, 'register', 'DeviceLimit', '5', 'int', 'Device binding limit', NOW(), NOW());
