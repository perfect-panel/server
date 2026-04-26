# 📋 Changelog / 更新日志

This document records all notable changes to PPanel Server.
本文档记录了 PPanel Server 的所有重要变更。

## Version Guide / 版本说明
- 🔥 **Breaking Changes / 重大变更**: Contains breaking updates / 包含破坏性更新
- ✨ **Features / 新功能**: New features added / 添加的新特性
- 🐛 **Bug Fixes / 问题修复**: Fixed bugs / 修复的bug
- 📚 **Documentation / 文档**: Documentation updates / 文档相关更新
- 🎨 **Style / 样式**: Code formatting and style changes / 代码格式化、样式调整
- ♻️ **Refactoring / 重构**: Code refactoring / 代码重构
- ⚡️ **Performance / 性能**: Performance improvements / 性能优化
- 🧪 **Tests / 测试**: Test related changes / 测试相关
- 🔨 **Build / 构建**: Build system changes / 构建系统相关
- 👷 **CI/CD**: Continuous integration changes / CI/CD相关
- 🔧 **Chores / 杂项**: Other changes / 其他变更

---

## [4.3.0](https://github.com/perfect-panel/server/releases/tag/v4.3.0) (2026-04-26)

### 🔥 Breaking Changes / 重大变更

* **billing model / 计费模型:** Replaced IP-based concurrent limit with **device-slot model**. Each `user_subscribe` now manages N device rows in `user_subscribe_device`; each device has its own `token` + `uuid` and produces an independent subscribe URL. Old `device_limit` semantics ("max concurrent IPs") becomes "device slot count" / 用「设备槽位」模型替代旧的「并发 IP」限制。每个订阅管理 N 个设备行,每台设备独立 `token` + `uuid` 生成独立 URL
* **password derivation / 密码派生:** All protocols (Shadowsocks / Trojan / Hysteria 2) now derive password as `sha256(uuid)[:16]`, replacing the old shared subscribe-token approach. This unifies the UUID-based credential model across protocols / SS/Trojan/Hy2 协议密码统一为 `sha256(uuid)[:16]`,UUID 同时作为身份凭证

### ✨ Features / 新功能

* **device management / 设备管理:** New endpoints — `add device` (with quantity, prorated pricing), `delete addon device`, `reset device` (rate-limited 3/hr 10/day), `disable / enable / rename device`, `reset all devices in plan` / 设备管理:加购、删除加购、换设备、停用/启用/重命名、整套餐重置
* **device limit enforcement / 设备限制:** Cross-node IP aggregation via Redis. Each device allows 1 IP + 1 grace IP; sharing triggers kick-off / 跨节点 IP 聚合,每设备 1 IP + 1 宽限 IP,共享触发互踢
* **multi-line subscribe URL / 多线路订阅 URL:** `BuildSubscribeURLs` returns one URL per line in `SubscribeDomain` (textarea, multi-line) — main + backup + CDN routes / 单设备多 URL 输出,支持主线/备用/CDN
* **auto-update interval / 自动更新间隔:** New admin config `UpdateIntervalHours`. Subscribe handler smart-injects per UA: `Profile-Update-Interval` header for Clash family / Hiddify; `#!MANAGED-CONFIG <url> interval=<seconds>` directive prepended to body for Surge / Stash / 智能识别 UA 注入自动更新指令
* **subscribe domain / 订阅域名:** `BuildSubscribeURL` no longer a stub — uses configured `SubscribeDomain` (multi-line, take first) + `SubscribePath` (default `/v1/subscribe/config`). Falls back gracefully when domain empty / 修复 `BuildSubscribeURL` 占位实现,使用 admin 配置
* **friendly errors / 友好错误:** `applyResetFrequencyCheck` returns Chinese human-readable errors with next available time + remaining duration ("本小时换设备已达上限,请于 17:00 后重试") / 重置频率限制返回人话错误带等待时长
* **CMS / 内容管理:** `site_content` table for terms/tutorials, multi-language with fallback. Admin can edit via Markdown editor; user-side renders with template variables (`{{.SubscribeUrl}}`, `{{.AppScheme}}`, etc.) / CMS 站点内容:服务条款 + 11 客户端教程
* **audit log / 审计日志:** Chinese localization for action labels (登录 / 重置设备 / 加购 / 退订 etc.); new `audit_log` viewer in admin / 审计日志中文本地化
* **trial balance / 试用赠送:** New register config `TrialBalance` (default ¥2 = 200 cents); auto-credited on registration / 注册送 ¥2 余额配置
* **renewal pricing / 续订定价:** Renewal order now includes prorated cost of existing addon devices / 续订包含加购设备费用

### 🐛 Bug Fixes / 问题修复

* **add device handler:** Was not parsing JSON body — `quantity` always read as 0. Fix: `c.ShouldBindJSON(&req)` before path param parse / 加购接口未读 body 导致 quantity 失效
* **DeepCopy field loss:** Added missing V4.3 fields (`DeviceCount`, `TrafficAddon`, `IsAddon`, `UnitPricePerDevice`, etc.) to response types — `tool.DeepCopy` silently drops fields not present in target struct / DeepCopy 反射拷贝静默丢失 V4.3 字段
* **migration table name:** Initial `02132_subscribe_device_is_addon` migrated wrong table — corrected to `user_subscribe_device` / 迁移表名错误
* **purchase default quantity:** When plan `device_limit=2`, default purchase only created 1 device. Fixed `activateOrderLogic` fallback + added one-shot `SubscribeDeviceBackfill` to repair existing data / 购买默认设备数错误
* **error message routing:** Backend `msg` field was being ignored in favor of generic framework messages. Now custom server messages flow through to user toasts / 自定义错误消息被通用模板覆盖

### ⚡️ Performance / 性能

* **monaco / mathjs / etc:** Frontend cold start optimization (see frontend CHANGELOG)
* **redis aggregation:** Device IP set uses Redis `SADD` + TTL, scales linearly with concurrent devices / Redis 设备 IP 聚合,线性扩展

### ♻️ Refactoring / 重构

* **Hysteria 2 naming / Hysteria 2 命名:** Consistent `Hysteria 2` across logs, types, configs (was mixed with `Hy2` / `hysteria2`) / 跨项目统一 Hysteria 2 命名
* **device addon flag / 设备类型标志:** New `is_addon` column distinguishes user-purchased addon devices (deletable) from plan-included base devices (locked) / 加购/基础设备区分

### 🔨 Build / 构建

* **version default / 版本默认值:** Source-level default in `pkg/constant/version.go` set to `v4.3.0` for local builds. Release CI continues to inject `git describe --tags` via Makefile ldflags

### 🗄️ Migrations / 数据库迁移

New migrations introduced in this release:

| Migration | Purpose |
|---|---|
| `02126_device_billing` | Initial `user_subscribe_device` table with token + uuid + status |
| `02127_user_message` | User notification inbox |
| `02128_terms_version` | Terms version tracking for re-acceptance |
| `02129_user_tags` | User tag system |
| `02130_application_tutorial_key` | Per-client tutorial CMS link |
| `02131_application_enabled` | Application enabled toggle (replaces "default") |
| `02132_subscribe_device_is_addon` | Distinguish addon vs base devices |
| `02133_subscribe_update_interval` | Admin-configurable auto-update interval |

---


## Earlier versions / 历史版本

See git history for changes prior to V4.3.0.
