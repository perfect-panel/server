-- V4.3:订阅自动更新间隔(小时),默认 24 小时。
-- 落地方式按客户端 UA 智能切:
--   * Clash / Hiddify: HTTP 响应头 Profile-Update-Interval(小时)
--   * Surge / Stash:   配置体首行 #!MANAGED-CONFIG <url> interval=<秒>
--   * 其它客户端:     不支持,客户端自行忽略
INSERT IGNORE INTO `system` (`category`, `key`, `value`, `type`, `desc`, `created_at`, `updated_at`)
VALUES
    ('subscribe', 'UpdateIntervalHours', '24', 'int', 'Subscription auto-update interval (hours)', NOW(), NOW());
