# PPanel Server

<div align="center">

[![License](https://img.shields.io/github/license/perfect-panel/server)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/perfect-panel/server)](https://goreportcard.com/report/github.com/perfect-panel/server)
[![Docker](https://img.shields.io/badge/Docker-Available-blue)](Dockerfile)
[![CI/CD](https://img.shields.io/github/actions/workflow/status/perfect-panel/server/release.yml)](.github/workflows/release.yml)

**PPanel is a pure, professional, and perfect open-source proxy panel tool, designed for learning and practical use.**

[English](README.md) | [中文](readme_zh.md) | [Report Bug](https://github.com/perfect-panel/server/issues/new) | [Request Feature](https://github.com/perfect-panel/server/issues/new)

</div>

> **Article 1.**  
> All human beings are born free and equal in dignity and rights.  
> They are endowed with reason and conscience and should act towards one another in a spirit of brotherhood.
>
> **Article 12.**  
> No one shall be subjected to arbitrary interference with his privacy, family, home or correspondence, nor to attacks upon his honour and reputation.  
> Everyone has the right to the protection of the law against such interference or attacks.
>
> **Article 19.**  
> Everyone has the right to freedom of opinion and expression; this right includes freedom to hold opinions without interference and to seek, receive and impart information and ideas through any media and regardless of frontiers.
>
> *Source: [United Nations – Universal Declaration of Human Rights (UN.org)](https://www.un.org/sites/un2.un.org/files/2021/03/udhr.pdf)*

## 📋 Overview

PPanel Server is the backend component of the PPanel project, providing robust APIs and core functionality for managing
proxy services. Built with Go, it emphasizes performance, security, and scalability.

### Key Features

- **Multi-Protocol Support**: Supports Shadowsocks, V2Ray, Trojan, and more.
- **Privacy First**: No user logs are collected, ensuring privacy and security.
- **Minimalist Design**: Simple yet powerful, with complete business logic.
- **User Management**: Full authentication and authorization system.
- **Subscription System**: Manage user subscriptions and service provisioning.
- **Payment Integration**: Supports multiple payment gateways.
- **Order Management**: Track and process user orders.
- **Ticket System**: Built-in customer support and issue tracking.
- **Node Management**: Monitor and control server nodes.
- **API Framework**: Comprehensive RESTful APIs for frontend integration.

## 🚀 Quick Start

### Prerequisites

- **Go**: 1.21 or higher
- **Docker**: Optional, for containerized deployment
- **Git**: For cloning the repository

### Installation from Source

1. **Clone the repository**:
   ```bash
   git clone https://github.com/perfect-panel/ppanel-server.git
   cd ppanel-server
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Generate code**:
   ```bash
   chmod +x script/generate.sh
   ./script/generate.sh
   ```

4. **Build the project**:
   ```bash
   make linux-amd64
   ```

5. **Run the server**:
   ```bash
   ./ppanel-server-linux-amd64 run --config etc/ppanel.yaml
   ```

### 🐳 Docker Deployment

1. **Build the Docker image**:
   ```bash
   docker buildx build --platform linux/amd64 -t ppanel-server:latest .
   ```

2. **Run the container**:
   ```bash
   docker run --rm -p 8080:8080 -v $(pwd)/etc:/app/etc ppanel-server:latest
   ```

3. **Use Docker Compose** (create `docker-compose.yml`):
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
   Run:
   ```bash
   docker-compose up -d
   ```

4. **Pull from Docker Hub** (after CI/CD publishes):
   ```bash
   docker pull ppanel/ppanel-server:latest
   docker run --rm -p 8080:8080 ppanel/ppanel-server:latest
   ```

## 📖 API Documentation

Explore the full API documentation:

- **Swagger**: [https://ppanel.dev/en-US/swagger/ppanel](https://ppanel.dev/swagger/ppanel)

The documentation covers all endpoints, request/response formats, and authentication details.

## 🔗 Related Projects

| Project          | Description                | Link                                                  |
|------------------|----------------------------|-------------------------------------------------------|
| PPanel Web       | Frontend for PPanel        | [GitHub](https://github.com/perfect-panel/ppanel-web) |
| PPanel User Web  | User interface for PPanel  | [Preview](https://user.ppanel.dev)                    |
| PPanel Admin Web | Admin interface for PPanel | [Preview](https://admin.ppanel.dev)                   |

## 🌐 Official Website

Visit [ppanel.dev](https://ppanel.dev/) for more details.

## 🏛 Architecture

![Architecture Diagram](./doc/image/architecture-en.png)

## 📁 Directory Structure

```
.
├── apis/             # API definition files
├── cmd/              # Application entry point
├── doc/              # Documentation
├── etc/              # Configuration files (e.g., ppanel.yaml)
├── generate/         # Code generation tools
├── initialize/       # System initialization
├── internal/         # Internal modules
│   ├── config/       # Configuration parsing
│   ├── handler/      # HTTP handlers
│   ├── middleware/   # HTTP middleware
│   ├── logic/        # Business logic
│   ├── model/        # Data models
│   ├── svc/          # Service layer
│   └── types/        # Type definitions
├── pkg/              # Utility code
├── queue/            # Queue services
├── scheduler/        # Scheduled tasks
├── script/           # Build scripts
├── go.mod            # Go module definition
├── Makefile          # Build automation
└── Dockerfile        # Docker configuration
```

## 💻 Development

### Format API Files
```bash
goctl api format --dir apis/user.api
```

### Add a New API

1. Create a new API file in `apis/`.
2. Import it in `apis/ppanel.api`.
3. Regenerate code:
   ```bash
   ./script/generate.sh
   ```

### Build for Multiple Platforms

Use the `Makefile` to build for various platforms (e.g., Linux, Windows, macOS):

```bash
make all  # Builds linux-amd64, darwin-amd64, windows-amd64
make linux-arm64  # Build for specific platform
```

Supported platforms include:

- Linux: `386`, `amd64`, `arm64`, `armv5-v7`, `mips`, `riscv64`, `loong64`, etc.
- Windows: `386`, `amd64`, `arm64`, `armv7`
- macOS: `amd64`, `arm64`
- FreeBSD: `amd64`, `arm64`

## 🤝 Contributing

Contributions are welcome! Please follow the [Contribution Guidelines](CONTRIBUTING.md) for bug fixes, features, or
documentation improvements.

## ✨ Special Thanks

A huge thank you to the following outstanding open-source projects that have provided invaluable support for this
project's development! 🚀

<div style="overflow-x: auto;">
<table style="width: 100%; border-collapse: collapse; margin: 20px 0;">
  <thead>
    <tr style="background-color: #f5f5f5;">
      <th style="padding: 10px; text-align: center;">Project</th>
      <th style="padding: 10px; text-align: left;">Description</th>
      <th style="padding: 10px; text-align: center;">Project</th>
      <th style="padding: 10px; text-align: left;">Description</th>
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
        High-performance Go Web framework<br/>
      </td>
      <td align="center" style="padding: 15px; vertical-align: middle;">
        <a href="https://gorm.io/" style="text-decoration: none;">
          <img src="https://gorm.io/gorm.svg" width="50" alt="Gorm" style="border-radius: 8px;" /><br/>
          <strong>Gorm</strong><br/>
          <img src="https://img.shields.io/github/stars/go-gorm/gorm?style=social" alt="Gorm Stars" />
        </a>
      </td>
      <td style="padding: 15px; vertical-align: middle;">
        Powerful Go ORM framework<br/>
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
        Asynchronous task queue for Go<br/>
      </td>
      <td align="center" style="padding: 15px; vertical-align: middle;">
        <a href="https://goswagger.io/" style="text-decoration: none;">
          <img src="https://goswagger.io/go-swagger/logo.png" width="30" alt="Go-Swagger" style="border-radius: 8px;" /><br/>
          <strong>Go-Swagger</strong><br/>
          <img src="https://img.shields.io/github/stars/go-swagger/go-swagger?style=social" alt="Go-Swagger Stars" />
        </a>
      </td>
      <td style="padding: 15px; vertical-align: middle;">
        Comprehensive Go Swagger toolkit<br/>
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
        Go microservices framework (this project's API generator is built on Go-Zero)<br/>
      </td>
    </tr>
  </tbody>
</table>
</div>

---

🎉 **Salute to Open Source**: Thank you to the open-source community for making development simpler and more efficient!
Please give these projects a ⭐ to support the open-source movement!
## 📄 License

This project is licensed under the [GPL-3.0 License](LICENSE).