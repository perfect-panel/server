# PPanel 服务端

<div align="center">

[![License](https://img.shields.io/github/license/perfect-panel/ppanel-server)](LICENSE)
![Go Version](https://img.shields.io/badge/Go-1.16%2B-blue)
[![Docker](https://img.shields.io/badge/Docker-Available-blue)](Dockerfile)

**PPanel 是一个纯净、专业、完美的开源代理面板工具，旨在成为您学习和实际使用的理想选择。**

[English](README.md) | [中文](readme_zh.md) | [报告问题](https://github.com/perfect-panel/ppanel-server/issues/new) | [功能请求](https://github.com/perfect-panel/ppanel-server/issues/new)

</div>

## 📋 概述

PPanel 服务端是 PPanel 项目的后端组件，为 PPanel 系统提供强大的 API 和核心功能。它使用 Go 语言构建，注重性能、安全性和可扩展性。

### 核心特性

- **多协议支持**：管理各种加密协议，包括 Shadowsocks、V2Ray、Trojan 等
- **隐私保护**：不收集用户日志，确保用户隐私和安全
- **极简设计**：易于使用的产品，同时保持业务逻辑的完整性
- **用户系统**：完整的用户管理，包含认证和授权
- **订阅管理**：处理用户订阅和服务提供
- **支付集成**：支持多种支付网关
- **订单管理**：处理和跟踪用户订单
- **工单系统**：客户支持和问题跟踪
- **节点管理**：服务器节点监控和控制
- **API 框架**：为前端应用提供全面的 API 接口

## 🚀 快速开始

### 前提条件

- Go 1.16+
- Docker（可选，用于容器化部署）

### 通过源代码运行

1. 克隆仓库

```bash
git clone https://github.com/perfect-panel/ppanel-server.git
cd ppanel-server
```

2. 安装依赖

```bash
go mod download
```

3. 生成代码

```bash
chmod +x script/generate.sh
./script/generate.sh
```

4. 构建项目

```bash
go build -o ppanel ppanel.go
```

5. 启动服务器

```bash
./ppanel run --config etc/ppanel.yaml
```

### 🐳 Docker 部署

1. 构建 Docker 镜像

```bash
docker build -t ppanel-server .
```

2. 运行容器

```bash
docker run -p 8080:8080 -v $(pwd)/etc/ppanel.yaml:/app/etc/ppanel.yaml ppanel-server
```

或使用 Docker Compose：

```bash
docker-compose up -d
```

## 📖 API 文档

PPanel 提供了全面的在线 API 文档：

- **官方 Swagger 文档**：[https://ppanel.dev/zh-CN/swagger/ppanel](https://ppanel.dev/zh-CN/swagger/ppanel)

该文档包含所有可用的 API 端点、请求/响应格式以及认证需求。

## 🔗 相关项目

| 项目               | 描述            | 链接                                                    |
|------------------|---------------|-------------------------------------------------------|
| PPanel Web       | PPanel 的前端应用  | [GitHub](https://github.com/perfect-panel/ppanel-web) |
| PPanel User Web  | PPanel 的用户界面  | [预览](https://user.ppanel.dev)                         |
| PPanel Admin Web | PPanel 的管理员界面 | [预览](https://admin.ppanel.dev)                        |

## 🌐 官方网站

访问我们的官方网站获取更多信息：[ppanel.dev](https://ppanel.dev/)

## 📁 目录结构

```
.
├── etc               # 配置文件目录
├── cmd               # 应用入口
├── queue             # 队列消费服务
├── generate          # 代码生成工具
├── initialize        # 系统初始化配置
├── go.mod            # Go 模块定义
├── internal          # 内部模块
│   ├── config        # 配置文件解析
│   ├── handler       # HTTP 接口处理
│   ├── middleware    # HTTP 中间件
│   ├── logic         # 业务逻辑处理
│   ├── svc           # 服务层封装
│   ├── types         # 类型定义
│   └── model         # 数据模型
├── scheduler         # 计划任务
├── pkg               # 公共工具代码
├── apis              # API 定义文件
├── script            # 构建脚本
└── doc               # 文档
```

## 💻 开发

### 格式化 API 文件

```bash
goctl api format --dir api/user.api
```

### 添加新 API

1. 在 `apis` 目录下创建新的 API 定义文件
2. 在 `ppanel.api` 文件中导入新的 API 定义
3. 运行生成脚本重新生成代码

```bash
./script/generate.sh
```

## 🤝 贡献

我们欢迎各种形式的贡献，无论是功能开发、错误修复还是文档改进。请查看[贡献指南](CONTRIBUTING_ZH.md)了解更多详情。

## 📄 许可证

本项目采用 [GPL-3.0 许可证](LICENSE) 授权。
