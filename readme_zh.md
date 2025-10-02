# PPanel 服务端

<div align="center">

[![License](https://img.shields.io/github/license/perfect-panel/server)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/perfect-panel/server)](https://goreportcard.com/report/github.com/perfect-panel/server)
[![Docker](https://img.shields.io/badge/Docker-Available-blue)](Dockerfile)
[![CI/CD](https://img.shields.io/github/actions/workflow/status/perfect-panel/server/release.yml)](.github/workflows/release.yml)

**PPanel 是一个纯净、专业、完美的开源代理面板工具，旨在成为您学习和实际使用的理想选择。**

[English](README.md) | [中文](readme_zh.md) | [报告问题](https://github.com/perfect-panel/server/issues/new) | [功能请求](https://github.com/perfect-panel/server/issues/new)

</div>

> **第一条**  
> 人人生而自由，在尊严与权利上一律平等。  
> 他们赋有理性与良知，应当以兄弟般的精神彼此相待。  
>
> **第十二条**  
> 任何人的隐私、家庭、住宅和通信不得任意干涉，其名誉与荣誉不得加以攻击。  
> 人人有权受到法律的保护，以免遭受这种干涉或攻击。  
>
> **第十九条**  
> 人人有思想与表达的自由；此项自由包括持有主张而不受干预，以及通过任何媒介、无论国界，自由寻求、接受和传播信息与思想。  
>
> *来源： [United Nations – Universal Declaration of Human Rights (UN.org)](https://www.un.org/sites/un2.un.org/files/2021/03/udhr.pdf)*

## 📋 概述

PPanel 服务端是 PPanel 项目的后端组件，为代理服务提供强大的 API 和核心功能。它基于 Go 语言开发，注重性能、安全性和可扩展性。

### 核心特性

- **多协议支持**：支持 Shadowsocks、V2Ray、Trojan 等多种加密协议。
- **隐私保护**：不收集用户日志，确保隐私和安全。
- **极简设计**：简单易用，保留完整的业务逻辑。
- **用户管理**：完善的认证和授权系统。
- **订阅管理**：处理用户订阅和服务开通。
- **支付集成**：支持多种支付网关。
- **订单管理**：跟踪和处理用户订单。
- **工单系统**：内置客户支持和问题跟踪。
- **节点管理**：监控和控制服务器节点。
- **API 框架**：提供全面的 RESTful API，供前端集成。

## 🚀 快速开始

### 前提条件

- **Go**：1.21 或更高版本
- **Docker**：可选，用于容器化部署
- **Git**：用于克隆仓库

### 通过源代码运行

1. **克隆仓库**：
   ```bash
   git clone https://github.com/perfect-panel/ppanel-server.git
   cd ppanel-server
   ```

2. **安装依赖**：
   ```bash
   go mod download
   ```

3. **生成代码**：
   ```bash
   chmod +x script/generate.sh
   ./script/generate.sh
   ```

4. **构建项目**：
   ```bash
   make linux-amd64
   ```

5. **启动服务器**：
   ```bash
   ./ppanel-server-linux-amd64 run --config etc/ppanel.yaml
   ```

### 🐳 Docker 部署

1. **构建 Docker 镜像**：
   ```bash
   docker buildx build --platform linux/amd64 -t ppanel-server:latest .
   ```

2. **运行容器**：
   ```bash
   docker run --rm -p 8080:8080 -v $(pwd)/etc:/app/etc ppanel-server:latest
   ```

3. **使用 Docker Compose**（创建 `docker-compose.yml`）：
   ```yaml
   version: '3.8'
   services:
     ppanel-server:
       image: ppanel-server:latest
       ports:
         - "8080:8080"
       volumes:
         - ./etc:/app/etc
       environment:
         - TZ=Asia/Shanghai
   ```
   运行：
   ```bash
   docker-compose up -d
   ```

4. **从 Docker Hub 拉取**（CI/CD 发布后）：
   ```bash
   docker pull ppanel/ppanel-server:latest
   docker run --rm -p 8080:8080 ppanel/ppanel-server:latest
   ```

## 📖 API 文档

查看完整的 API 文档：

- **Swagger 文档**：[https://ppanel.dev/zh-CN/swagger/ppanel](https://ppanel.dev/zh-CN/swagger/ppanel)

文档涵盖所有 API 端点、请求/响应格式及认证要求。

## 🔗 相关项目

| 项目               | 描述           | 链接                                                    |
|------------------|--------------|-------------------------------------------------------|
| PPanel Web       | PPanel 前端应用  | [GitHub](https://github.com/perfect-panel/ppanel-web) |
| PPanel User Web  | PPanel 用户界面  | [预览](https://user.ppanel.dev)                         |
| PPanel Admin Web | PPanel 管理员界面 | [预览](https://admin.ppanel.dev)                        |

## 🌐 官方网站

访问 [ppanel.dev](https://ppanel.dev) 获取更多信息。

## 🏛 系统架构

![Architecture Diagram](./doc/image/architecture-zh.png)

## 📁 目录结构

```
.
├── apis/             # API 定义文件
├── cmd/              # 应用程序入口
├── doc/              # 文档
├── etc/              # 配置文件（如 ppanel.yaml）
├── generate/         # 代码生成工具
├── initialize/       # 系统初始化
├── internal/         # 内部模块
│   ├── config/       # 配置文件解析
│   ├── handler/      # HTTP 处理程序
│   ├── middleware/   # HTTP 中间件
│   ├── logic/        # 业务逻辑
│   ├── model/        # 数据模型
│   ├── svc/          # 服务层
│   └── types/        # 类型定义
├── pkg/              # 公共工具代码
├── queue/            # 队列服务
├── scheduler/        # 定时任务
├── script/           # 构建脚本
├── go.mod            # Go 模块定义
├── Makefile          # 构建自动化
└── Dockerfile        # Docker 配置
```

## 💻 开发

### 格式化 API 文件
```bash
goctl api format --dir apis/user.api
```

### 添加新 API

1. 在 `apis/` 目录创建新的 API 文件。
2. 在 `apis/ppanel.api` 中导入新 API。
3. 重新生成代码：
   ```bash
   ./script/generate.sh
   ```

### 多平台构建

使用 `Makefile` 构建多种平台（如 Linux、Windows、macOS）：

```bash
make all  # 构建 linux-amd64、darwin-amd64、windows-amd64
make linux-arm64  # 构建特定平台
```

支持的平台包括：

- Linux：`386`、`amd64`、`arm64`、`armv5-v7`、`mips`、`riscv64`、`loong64` 等
- Windows：`386`、`amd64`、`arm64`、`armv7`
- macOS：`amd64`、`arm64`
- FreeBSD：`amd64`、`arm64`

## 🤝 贡献

欢迎各种贡献，包括功能开发、错误修复和文档改进。请查看[贡献指南](CONTRIBUTING_ZH.md)了解详情。

## ✨ 特别感谢

感谢以下优秀的开源项目，它们为本项目的开发提供了强大的支持！ 🚀

<div style="overflow-x: auto;">
<table style="width: 100%; border-collapse: collapse; margin: 20px 0;">
  <thead>
    <tr style="background-color: #f5f5f5;">
      <th style="padding: 10px; text-align: center;">项目</th>
      <th style="padding: 10px; text-align: left;">描述</th>
      <th style="padding: 10px; text-align: center;">项目</th>
      <th style="padding: 10px; text-align: left;">描述</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td align="center" style="padding: 15px; vertical-align: middle;">
        <a href="https://gin-gonic.com/" style="text-decoration: none;">
          <img src="https://raw.githubusercontent.com/gin-gonic/logo/master/color.png" width="25" alt="Gin" style="border-radius: 8px;" /><br/>
          <strong>Gin</strong><br/>
          <img src="https://img.shields.io/github/stars/gin-gonic/gin?style=social" alt="Gin Stars" />
        </a>
      </td>
      <td style="padding: 15px; vertical-align: middle;">
        高性能的 Go Web 框架<br/>
      </td>
      <td align="center" style="padding: 15px; vertical-align: middle;">
        <a href="https://gorm.io/" style="text-decoration: none;">
          <img src="https://gorm.io/gorm.svg" width="50" alt="Gorm" style="border-radius: 8px;" /><br/>
          <strong>Gorm</strong><br/>
          <img src="https://img.shields.io/github/stars/go-gorm/gorm?style=social" alt="Gorm Stars" />
        </a>
      </td>
      <td style="padding: 15px; vertical-align: middle;">
        功能强大的 Go ORM 框架<br/>
      </td>
    </tr>
    <tr>
      <td align="center" style="padding: 15px; vertical-align: middle;">
        <a href="https://github.com/hibiken/asynq" style="text-decoration: none;">
          <img src="https://user-images.githubusercontent.com/11155743/114697792-ffbfa580-9d26-11eb-8e5b-33bef69476dc.png" width="50" alt="Asynq" style="border-radius: 8px;" /><br/>
          <strong>Asynq</strong><br/>
          <img src="https://img.shields.io/github/stars/hibiken/asynq?style=social" alt="Asynq Stars" />
        </a>
      </td>
      <td style="padding: 15px; vertical-align: middle;">
        Go 语言的异步任务队列<br/>
      </td>
      <td align="center" style="padding: 15px; vertical-align: middle;">
        <a href="https://goswagger.io/" style="text-decoration: none;">
          <img src="https://goswagger.io/go-swagger/logo.png" width="30" alt="Go-Swagger" style="border-radius: 8px;" /><br/>
          <strong>Go-Swagger</strong><br/>
          <img src="https://img.shields.io/github/stars/go-swagger/go-swagger?style=social" alt="Go-Swagger Stars" />
        </a>
      </td>
      <td style="padding: 15px; vertical-align: middle;">
        完整的 Go Swagger 工具集<br/>
      </td>
    </tr>
    <tr>
      <td align="center" style="padding: 15px; vertical-align: middle;">
        <a href="https://go-zero.dev/" style="text-decoration: none;">
          <img src="https://raw.githubusercontent.com/zeromicro/zero-doc/main/doc/images/go-zero.png" width="30" alt="Go-Zero" style="border-radius: 8px;" /><br/>
          <strong>Go-Zero</strong><br/>
          <img src="https://img.shields.io/github/stars/zeromicro/go-zero?style=social" alt="Go-Zero Stars" />
        </a>
      </td>
      <td colspan="3" style="padding: 15px; vertical-align: middle;">
        Go 微服务框架（本项目的 API 生成器，基于 Go-Zero 实现）<br/>
      </td>
    </tr>
  </tbody>
</table>
</div>

---

🎉 **致敬开源**：感谢开源社区，让开发变得更简单、更高效！欢迎为这些项目点亮 ⭐，支持开源事业！

## 📄 许可证

本项目采用 [GPL-3.0 许可证](LICENSE) 授权。