# PPanel Server

<div align="center">

[![License](https://img.shields.io/github/license/perfect-panel/ppanel-server)](LICENSE)
![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)
[![Docker](https://img.shields.io/badge/Docker-Available-blue)](Dockerfile)
[![CI/CD](https://img.shields.io/github/actions/workflow/status/perfect-panel/ppanel-server/release.yml)](.github/workflows/release.yml)

**PPanel is a pure, professional, and perfect open-source proxy panel tool, designed for learning and practical use.**

[English](README.md) | [中文](README_zh.md) | [Report Bug](https://github.com/perfect-panel/ppanel-server/issues/new) | [Request Feature](https://github.com/perfect-panel/ppanel-server/issues/new)

</div>

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
   docker pull yourusername/ppanel-server:latest
   docker run --rm -p 8080:8080 yourusername/ppanel-server:latest
   ```

## 📖 API Documentation

Explore the full API documentation:

- **Swagger**: [https://ppanel.dev/swagger/ppanel](https://ppanel.dev/swagger/ppanel)

The documentation covers all endpoints, request/response formats, and authentication details.

## 🔗 Related Projects

| Project          | Description                | Link                                                  |
|------------------|----------------------------|-------------------------------------------------------|
| PPanel Web       | Frontend for PPanel        | [GitHub](https://github.com/perfect-panel/ppanel-web) |
| PPanel User Web  | User interface for PPanel  | [Preview](https://user.ppanel.dev)                    |
| PPanel Admin Web | Admin interface for PPanel | [Preview](https://admin.ppanel.dev)                   |

## 🌐 Official Website

Visit [ppanel.dev](https://ppanel.dev/) for more details.

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

## 📄 License

This project is licensed under the [GPL-3.0 License](LICENSE).