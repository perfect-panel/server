-- Remove device limit configuration from system table
DELETE FROM `system` WHERE `category` = 'register' AND `key` = 'DeviceLimit';
