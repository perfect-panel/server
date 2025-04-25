# PPanel Server

<div align="center">

[![License](https://img.shields.io/github/license/perfect-panel/ppanel-server)](LICENSE)
![Go Version](https://img.shields.io/badge/Go-1.16%2B-blue)
[![Docker](https://img.shields.io/badge/Docker-Available-blue)](Dockerfile)

**PPanel is a pure, professional, and perfect open-source proxy panel tool, designed to be your ideal choice for
learning and practical use.**

[English](README.md) | [中文](readme_zh.md) | [Report Bug](https://github.com/perfect-panel/server/issues/new) | [Request Feature](https://github.com/perfect-panel/server/issues/new)

</div>

## 📋 Overview

PPanel Server is the backend component of the PPanel project, providing robust APIs and core functionality for the
PPanel system. It's built with Go and designed with performance, security, and scalability in mind.

### Key Features

- **Multi-Protocol Support**: Manages various encryption protocols including Shadowsocks, V2Ray, Trojan, and more
- **Privacy Protection**: No user logs are collected, ensuring user privacy and security
- **Minimalist Design**: Easy-to-use product while maintaining the integrity of business logic
- **User System**: Complete user management with authentication and authorization
- **Subscription Management**: Handle user subscriptions and service provisions
- **Payment Integration**: Multiple payment gateway support
- **Order Management**: Process and track user orders
- **Ticket System**: Customer support and issue tracking
- **Node Management**: Server node monitoring and control
- **API Framework**: Comprehensive API interfaces for frontend applications

## 🚀 Quick Start

### Prerequisites

- Go 1.16+
- Docker (optional, for containerized deployment)

### Running from Source Code

1. Clone the repository

```bash
git clone https://github.com/perfect-panel/server.git
cd ppanel-server
```

2. Install dependencies

```bash
go mod download
```

3. Generate code

```bash
chmod +x script/generate.sh
./script/generate.sh
```

4. Build the project

```bash
go build -o ppanel ppanel.go
```

5. Start the server

```bash
./ppanel run --config etc/ppanel.yaml
```

### 🐳 Docker Deployment

1. Build Docker image

```bash
docker build -t ppanel-server .
```

2. Run container

```bash
docker run -p 8080:8080 -v $(pwd)/etc/ppanel.yaml:/app/etc/ppanel.yaml ppanel-server
```

Or use Docker Compose:

```bash
docker-compose up -d
```

## 📖 API Documentation

PPanel provides comprehensive API documentation available online:

- **Official Swagger Documentation**: [https://ppanel.dev/swagger/ppanel](https://ppanel.dev/swagger/ppanel)

The documentation includes all available API endpoints, request/response formats, and authentication requirements.

## 🔗 Related Projects

| Project          | Description                      | Link                                                  |
|------------------|----------------------------------|-------------------------------------------------------|
| PPanel Web       | Frontend applications for PPanel | [GitHub](https://github.com/perfect-panel/ppanel-web) |
| PPanel User Web  | User interface for PPanel        | [Preview](https://user.ppanel.dev)                    |
| PPanel Admin Web | Admin interface for PPanel       | [Preview](https://admin.ppanel.dev)                   |

## 🌐 Official Website

For more information, visit our official website: [ppanel.dev](https://ppanel.dev/)

## 📁 Directory Structure

```
.
├── etc               # Configuration files directory
├── cmd               # Application entry point
├── queue             # Queue consumption service
├── generate          # Code generation tools
├── initialize        # System initialization configuration
├── go.mod            # Go module definition
├── internal          # Internal modules
│   ├── config        # Configuration file parsing
│   ├── handler       # HTTP interface handling
│   ├── middleware    # HTTP middleware
│   ├── logic         # Business logic processing
│   ├── svc           # Service layer encapsulation
│   ├── types         # Type definitions
│   └── model         # Data models
├── scheduler         # Scheduled tasks
├── pkg               # Common utility code
├── apis              # API definition files
├── script            # Build scripts
└── doc               # Documentation
```

## 💻 Development

### Format API File

```bash
goctl api format --dir api/user.api
```

### Adding New API

1. Create a new API definition file in the `apis` directory
2. Import the new API definition in the `ppanel.api` file
3. Run the generation script to regenerate the code

```bash
./script/generate.sh
```

## 🤝 Contributing

We welcome all forms of contribution, whether it's feature development, bug fixes, or documentation improvements. Please
check the [Contribution Guidelines](CONTRIBUTING.md) for more details.

## 📄 License

This project is licensed under the [GPL-3.0 License](LICENSE).