# GPT-Load

[English](README.md) | 中文

[![Release](https://img.shields.io/github/v/release/ShangjinTang/gpt-load)](https://github.com/ShangjinTang/gpt-load/releases) ![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg) [![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

一个高性能、企业级的 AI 接口透明代理服务，专门为需要集成多种 AI 服务的企业和开发者设计。采用 Go 语言开发，具备智能密钥管理、高级策略引擎、模型过滤、增量验证和完善的监控功能，专为高并发生产环境而设计。

详细请查看[官方文档](https://www.gpt-load.com/docs?lang=zh)

<a href="https://trendshift.io/repositories/14880" target="_blank"><img src="https://trendshift.io/api/badge/repositories/14880" alt="ShangjinTang%2Fgpt-load | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a> <a href="https://hellogithub.com/repository/ShangjinTang/gpt-load" target="_blank"><img src="https://api.hellogithub.com/v1/widgets/recommend.svg?rid=554dc4c46eb14092b9b0c56f1eb9021c&claim_uid=Qlh8vzrWJ0HCneG" alt="Featured｜HelloGitHub" style="width: 250px; height: 54px;" width="250" height="54" /></a>

## 🚀 功能特性

### 核心代理功能

- **透明代理**: 完全保留原生 API 格式，支持 OpenAI、Google Gemini 和 Anthropic Claude 等多种格式
- **高性能设计**: 零拷贝流式传输、连接池复用、原子操作
- **负载均衡**: 支持多上游端点的加权负载均衡，提升服务可用性
- **优雅关闭**: 生产就绪的优雅关闭和错误恢复机制

### 🔑 高级密钥管理

- **智能密钥池**: 分组管理，支持自动轮换和故障恢复
- **智能状态机**: 5 状态密钥生命周期（待验证 → 活跃 → 降级 → 禁用 → 无效）
- **指数退避**: 失败密钥的智能退避重试机制
- **增量验证**: 基于时间窗口的验证，避免重复检查
- **批量操作**: 高性能的批量密钥导入、验证和管理

### 🎯 策略引擎与过滤

- **组合式策略**: 可复用的策略模板，可应用于多个分组
- **高级重试策略**: 基于错误码和消息的重试策略，支持自定义动作
- **模型过滤**: 支持包含/排除模式的正则表达式和通配符
- **动态策略评估**: 基于请求上下文的实时策略匹配和执行

### 🏗️ 企业级架构

- **分布式部署**: 主从架构，支持水平扩展和高可用
- **热配置重载**: 系统设置和分组配置支持热重载，无需重启
- **全面监控**: 实时统计、健康检查、详细请求日志
- **API 管理**: RESTful API 用于配置和监控
- **双重认证体系**: 管理端与代理端认证分离，代理认证支持全局和分组级别密钥

### 🔒 安全性与可靠性

- **数据加密**: API 密钥的可选 AES 加密存储，支持迁移
- **智能故障处理**: 自动密钥黑名单管理和恢复机制
- **速率限制**: 可配置的请求速率限制和并发请求控制
- **请求日志**: 详细的审计跟踪，可配置保留策略

## 🎯 支持的 AI 服务

GPT-Load 作为透明代理服务，完整保留各 AI 服务商的原生 API 格式：

- **OpenAI 格式**: 官方 OpenAI API、Azure OpenAI、以及其他 OpenAI 兼容服务
- **Google Gemini 格式**: Gemini Pro、Gemini Pro Vision 等模型的原生 API
- **Anthropic Claude 格式**: Claude 系列模型，支持高质量的对话和文本生成

## 📋 系统要求

- Go 1.23+ (源码构建)
- Docker (容器化部署)
- PostgreSQL 或 SQLite (数据库存储)
- Redis (缓存和分布式协调，可选)

## 🚀 快速开始

### 方式一：Docker 快速开始

```bash
docker run -d --name gpt-load \
    -p 3001:3001 \
    -e AUTH_KEY=your-secure-key-here \
    -v "$(pwd)/data":/app/data \
    ghcr.io/ShangjinTang/gpt-load:latest
```

> 请将 `your-secure-key-here` 修改为强密码（切勿使用默认值），然后访问管理 API：<http://localhost:3001>

### 方式二：使用 Docker Compose（推荐）

**安装命令：**

```bash
# 创建目录
mkdir -p gpt-load && cd gpt-load

# 下载配置文件
wget https://raw.githubusercontent.com/ShangjinTang/gpt-load/refs/heads/main/docker-compose.yml
wget -O .env https://raw.githubusercontent.com/ShangjinTang/gpt-load/refs/heads/main/.env.example

# 编辑 .env 文件，将 AUTH_KEY 修改为强密码，切勿使用默认或简单密钥如 sk-123456

# 启动服务
docker compose up -d
```

部署前必须修改默认管理密钥（AUTH_KEY），推荐格式：sk-prod-[32 位随机字符串]。

默认安装使用 SQLite 版本，适合轻量级单实例应用。

如需安装 PostgreSQL 和 Redis，请取消 `docker-compose.yml` 文件中相应服务的注释，配置对应环境变量并重启。

**其他命令：**

```bash
# 检查服务状态
docker compose ps

# 查看日志
docker compose logs -f

# 重启服务
docker compose down && docker compose up -d

# 更新到最新版本
docker compose pull && docker compose down && docker compose up -d
```

部署完成后：

- 访问管理 API：<http://localhost:3001>
- API 代理地址：<http://localhost:3001/proxy>

> 使用您修改的 AUTH_KEY 访问管理 API。

### 方式三：源码构建

源码构建需要本地安装数据库（SQLite 或 PostgreSQL）和 Redis（可选）。

```bash
# 克隆和构建
git clone https://github.com/ShangjinTang/gpt-load.git
cd gpt-load
go mod tidy

# 创建配置
cp .env.example .env

# 编辑 .env 文件，将 AUTH_KEY 修改为强密码，切勿使用默认或简单密钥如 sk-123456
# 修改 .env 中的 DATABASE_DSN 和 REDIS_DSN 配置
# REDIS_DSN 为可选，如未配置将启用内存存储

# 运行
make run
```

部署完成后：

- 访问管理 API：<http://localhost:3001>
- API 代理地址：<http://localhost:3001/proxy>

> 使用您修改的 AUTH_KEY 访问管理 API。

### 方式四：集群部署

集群部署需要所有节点连接相同的 PostgreSQL 和 Redis，Redis 为必须。建议使用统一的分布式 PostgreSQL 和 Redis 集群。

**部署要求：**

- 所有节点必须配置相同的 `AUTH_KEY`、`DATABASE_DSN`、`REDIS_DSN`
- 主从架构，从节点必须配置环境变量：`IS_SLAVE=true`

详情请参考[集群部署文档](https://www.gpt-load.com/docs/cluster?lang=zh)

## ⚙️ 配置系统

### 配置架构概述

GPT-Load 采用双层配置架构：

#### 1. 静态配置（环境变量）

- **特性**：应用启动时读取，运行时不可变，需要重启应用才能生效
- **用途**：基础设施配置，如数据库连接、服务器端口、认证密钥等
- **管理**：通过 `.env` 文件或系统环境变量设置

#### 2. 动态配置（热重载）

- **系统设置**：存储在数据库中，为整个应用提供统一的行为标准
- **分组配置**：针对特定分组定制的行为参数，可覆盖系统设置
- **配置优先级**：分组配置 > 系统设置 > 环境配置
- **特性**：支持热重载，修改后立即生效，无需重启应用

<details>
<summary>静态配置（环境变量）</summary>

**服务器配置：**

| 设置项 | 环境变量 | 默认值 | 描述 |
| --- | --- | --- | --- |
| 服务端口 | `PORT` | 3001 | HTTP 服务器监听端口 |
| 服务地址 | `HOST` | 0.0.0.0 | HTTP 服务器绑定地址 |
| 读取超时 | `SERVER_READ_TIMEOUT` | 60 | HTTP 服务器读取超时（秒） |
| 写入超时 | `SERVER_WRITE_TIMEOUT` | 600 | HTTP 服务器写入超时（秒） |
| 空闲超时 | `SERVER_IDLE_TIMEOUT` | 120 | HTTP 连接空闲超时（秒） |
| 优雅关闭超时 | `SERVER_GRACEFUL_SHUTDOWN_TIMEOUT` | 10 | 服务优雅关闭等待时间（秒） |
| 从节点模式 | `IS_SLAVE` | false | 集群部署的从节点标识 |
| 时区 | `TZ` | `Asia/Shanghai` | 指定时区 |

**安全配置：**

| 设置项 | 环境变量 | 默认值 | 描述 |
| --- | --- | --- | --- |
| 管理密钥 | `AUTH_KEY` | - | **管理端**访问认证密钥，请修改为强密码 |
| 加密密钥 | `ENCRYPTION_KEY` | - | 静态加密 API 密钥。支持任意字符串或留空禁用加密。⚠️ **设置后请勿直接修改。如需更改请参见[加密密钥迁移](#加密密钥迁移)。** |

**数据库配置：**

| 设置项 | 环境变量 | 默认值 | 描述 |
| --- | --- | --- | --- |
| 数据库连接 | `DATABASE_DSN` | `./data/gpt-load.db` | 数据库连接字符串（DSN）或文件路径 |
| Redis 连接 | `REDIS_DSN` | - | Redis 连接字符串，为空时使用内存存储 |

**性能与跨域配置：**

| 设置项 | 环境变量 | 默认值 | 描述 |
| --- | --- | --- | --- |
| 最大并发请求数 | `MAX_CONCURRENT_REQUESTS` | 100 | 系统允许的最大并发请求数 |
| 启用跨域 | `ENABLE_CORS` | false | 是否启用跨域资源共享 |
| 允许的源 | `ALLOWED_ORIGINS` | - | 允许的来源，逗号分隔 |
| 允许的方法 | `ALLOWED_METHODS` | `GET,POST,PUT,DELETE,OPTIONS` | 允许的 HTTP 方法 |
| 允许的请求头 | `ALLOWED_HEADERS` | `*` | 允许的请求头，逗号分隔 |
| 允许发送凭证 | `ALLOW_CREDENTIALS` | false | 是否允许发送凭证 |

**日志配置：**

| 设置项 | 环境变量 | 默认值 | 描述 |
| --- | --- | --- | --- |
| 日志级别 | `LOG_LEVEL` | `info` | 日志级别：debug、info、warn、error |
| 日志格式 | `LOG_FORMAT` | `text` | 日志格式：text、json |
| 启用文件日志 | `LOG_ENABLE_FILE` | false | 是否启用文件日志输出 |
| 日志文件路径 | `LOG_FILE_PATH` | `./data/logs/app.log` | 日志文件存储路径 |

**代理配置：**

GPT-Load 自动从环境变量读取代理设置，用于向上游 AI 服务商发起请求。

| 设置项     | 环境变量      | 默认值 | 描述                               |
| ---------- | ------------- | ------ | ---------------------------------- |
| HTTP 代理  | `HTTP_PROXY`  | -      | HTTP 请求的代理服务器地址          |
| HTTPS 代理 | `HTTPS_PROXY` | -      | HTTPS 请求的代理服务器地址         |
| 无代理     | `NO_PROXY`    | -      | 绕过代理的主机或域名列表，逗号分隔 |

支持的代理协议格式：

- **HTTP**: `http://user:pass@host:port`
- **HTTPS**: `https://user:pass@host:port`
- **SOCKS5**: `socks5://user:pass@host:port`
</details>

<details>
<summary>动态配置（热重载）</summary>

**基础设置：**

| 设置项 | 字段名 | 默认值 | 分组覆盖 | 描述 |
| --- | --- | --- | --- | --- |
| 项目地址 | `app_url` | `http://localhost:3001` | ❌ | 项目基础 URL |
| 全局代理密钥 | `proxy_keys` | `AUTH_KEY` 初始值 | ❌ | 全局生效的代理密钥，逗号分隔 |
| 日志保留天数 | `request_log_retention_days` | 7 | ❌ | 请求日志保留天数，0 为不清理 |
| 日志写入间隔 | `request_log_write_interval_minutes` | 1 | ❌ | 日志写入数据库的周期（分钟） |
| 启用请求体日志 | `enable_request_body_logging` | false | ✅ | 是否在请求日志中记录完整请求体内容 |

**请求设置：**

| 设置项 | 字段名 | 默认值 | 分组覆盖 | 描述 |
| --- | --- | --- | --- | --- |
| 请求超时 | `request_timeout` | 600 | ✅ | 转发请求完整生命周期超时（秒） |
| 连接超时 | `connect_timeout` | 15 | ✅ | 与上游服务建立连接的超时（秒） |
| 空闲连接超时 | `idle_conn_timeout` | 120 | ✅ | HTTP 客户端空闲连接超时（秒） |
| 响应头超时 | `response_header_timeout` | 600 | ✅ | 等待上游响应头的超时（秒） |
| 最大空闲连接数 | `max_idle_conns` | 100 | ✅ | 连接池最大总空闲连接数 |
| 每主机最大空闲连接数 | `max_idle_conns_per_host` | 50 | ✅ | 每个上游主机的最大空闲连接数 |
| 代理 URL | `proxy_url` | - | ✅ | 转发请求的 HTTP/HTTPS 代理，为空时使用环境变量 |

**密钥配置：**

| 设置项 | 字段名 | 默认值 | 分组覆盖 | 描述 |
| --- | --- | --- | --- | --- |
| 最大重试次数 | `max_retries` | 3 | ✅ | 单个请求使用不同密钥的最大重试次数 |
| 黑名单阈值 | `blacklist_threshold` | 3 | ✅ | 密钥连续失败多少次后进入黑名单 |
| 密钥验证间隔 | `key_validation_interval_minutes` | 60 | ✅ | 后台定时密钥验证的周期（分钟） |
| 密钥验证并发数 | `key_validation_concurrency` | 10 | ✅ | 后台验证无效密钥的并发数 |
| 密钥验证超时 | `key_validation_timeout_seconds` | 20 | ✅ | 后台验证单个密钥时 API 请求的超时（秒） |

</details>

## 🔐 加密密钥迁移

如果您需要在设置加密密钥后进行更改，请仔细按照以下步骤操作：

<details>
<summary>查看加密密钥迁移指南</summary>

### ⚠️ 重要警告

- **需要停机**：迁移期间必须停止服务
- **先备份**：迁移前务必备份数据库
- **无法恢复**：如果丢失当前加密密钥，加密数据将无法恢复
- **先测试**：请先在数据库副本上测试迁移过程

### 迁移步骤

1. **停止服务**

   ```bash
   # Docker Compose
   docker compose down

   # 或直接运行的情况
   # 停止运行中的进程（Ctrl+C 或 kill 进程）
   ```

2. **备份数据库**

   ```bash
   # SQLite（默认）
   cp ./data/gpt-load.db ./data/gpt-load.db.backup

   # PostgreSQL
   pg_dump your_database > backup.sql
   ```

3. **准备迁移脚本**

   复制 `scripts/migrate_encryption_key.go` 迁移脚本并修改配置：

   ```go
   const (
       DATABASE_DSN = "./data/gpt-load.db"  // 您的数据库连接
       OLD_ENCRYPTION_KEY = "current-key"   // 当前密钥
       NEW_ENCRYPTION_KEY = "new-32-char-secret-key"  // 新密钥
   )
   ```

4. **运行迁移**

   ```bash
   cd scripts
   go run migrate_encryption_key.go
   ```

5. **更新配置**

   ```bash
   # 更新 .env 文件
   echo "ENCRYPTION_KEY=new-32-char-secret-key" >> .env
   ```

6. **重启服务**

   ```bash
   # Docker Compose
   docker compose up -d

   # 或直接运行
   make run
   ```

### 迁移场景

- **启用加密**：设置 `OLD_ENCRYPTION_KEY = ""` 并提供 `NEW_ENCRYPTION_KEY`
- **禁用加密**：提供当前密钥作为 `OLD_ENCRYPTION_KEY` 并设置 `NEW_ENCRYPTION_KEY = ""`
- **更改密钥**：同时提供旧密钥和新密钥

### 故障排除

如果迁移失败：

1. 从备份恢复：`cp ./data/gpt-load.db.backup ./data/gpt-load.db`
2. 检查迁移脚本输出中的错误信息
3. 验证数据库连接字符串和加密密钥
4. 确保服务已完全停止

### 密钥生成

生成安全的 32 位密钥：

```bash
openssl rand -base64 32 | tr -d "=+/" | cut -c1-32
```

</details>

## 🏗️ 架构设计

GPT-Load 采用模块化、高性能的架构设计，专为企业级 AI 代理服务而构建。

<details>
<summary>查看架构概览</summary>

### 🎯 核心设计原则

- **高性能**：异步处理、连接池和智能缓存
- **可扩展性**：无状态设计支持水平扩展
- **可靠性**：熔断器模式、优雅降级和全面错误处理
- **安全性**：静态加密、安全密钥管理和审计日志
- **可维护性**：清洁架构、依赖注入和全面测试

### 🏛️ 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   客户端应用    │    │   负载均衡器    │    │     监控        │
└─────────┬───────┘    └─────────┬───────┘    └─────────────────┘
          │                      │
          └──────────┐          ┌┴──────────┐
                     │          │           │
                ┌────▼──────────▼───┐   ┌───▼────┐
                │    GPT-Load       │   │ 指标   │
                │    代理服务器     │   │ & 日志 │
                └────┬──────────────┘   └────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
   ┌────▼───┐   ┌───▼────┐   ┌───▼────┐
   │ OpenAI │   │ Claude │   │ 其他   │
   │   API  │   │   API  │   │  APIs  │
   └────────┘   └────────┘   └────────┘
```

### 🔧 组件架构

#### 1. **HTTP 层** (`internal/router`, `internal/handler`)

- **Gin 框架**：高性能 HTTP 路由
- **中间件栈**：CORS、认证、限流、日志
- **API 版本控制**：RESTful API 设计与版本支持

#### 2. **代理引擎** (`internal/proxy`)

- **请求路由**：智能上游选择
- **负载均衡**：轮询和加权算法
- **熔断器**：自动故障检测和恢复
- **响应处理**：流处理和格式转换

#### 3. **密钥管理** (`internal/keypool`)

- **状态机**：5 状态生命周期（待验证 → 活跃 → 降级 → 禁用 → 无效）
- **轮换算法**：原子性密钥选择和轮换
- **健康监控**：后台验证和自动恢复
- **加密**：AES-256-GCM 静态密钥加密

#### 4. **策略引擎** (`internal/policy`)

- **基于规则的逻辑**：灵活的重试和降级策略
- **模式匹配**：基于正则表达式的错误分类
- **动作执行**：自动密钥状态转换
- **组合策略**：分层策略继承

#### 5. **数据层** (`internal/db`, `internal/models`)

- **ORM**：GORM 数据库抽象
- **多数据库**：SQLite 和 PostgreSQL 支持
- **迁移**：自动模式管理
- **连接池**：优化的数据库连接

#### 6. **缓存层** (`internal/store`)

- **多后端**：内存和 Redis 支持
- **密钥分发**：一致性哈希实现可扩展性
- **TTL 管理**：自动过期和清理
- **发布/订阅**：实时配置更新

#### 7. **后台服务** (`internal/services`)

- **验证服务**：定期密钥健康检查
- **日志服务**：异步请求日志记录
- **状态服务**：密钥生命周期管理
- **增量验证**：智能验证与跳过逻辑

### 📊 数据流

#### 请求处理流程

```
1. 客户端请求 → 2. 身份验证 → 3. 速率限制
         ↓                    ↓                 ↓
4. 密钥选择 ← 5. 负载均衡 ← 6. 上游路由
         ↓                    ↓                 ↓
7. API 调用 → 8. 响应处理 → 9. 客户端响应
         ↓                    ↓                 ↓
10. 日志记录 ← 11. 密钥状态更新 ← 12. 策略评估
```

#### 密钥生命周期流程

```
新密钥 → 验证 → 活跃 → [成功/失败] → 策略检查
    ↓         ↓          ↓              ↓              ↓
  待验证 → 无效 → 降级 → 禁用 → 恢复/移除
```

### 🔄 状态管理

#### API 密钥状态

- **待验证**：新添加，等待验证
- **活跃**：已验证，可用于请求
- **降级**：存在问题但仍可使用
- **禁用**：因失败而临时禁用
- **无效**：永久无效，从轮换中排除

#### 状态转换

- **智能退避**：失败密钥的指数退避
- **自动恢复**：自动重新验证和状态恢复
- **策略驱动**：状态转换的自定义规则

### 🚀 性能特性

#### 并发与并行

- **Goroutine 池**：验证的受控并发
- **基于通道**：非阻塞通信
- **上下文取消**：优雅关闭支持

#### 缓存策略

- **多级**：L1（内存）+ L2（Redis）缓存
- **智能失效**：事件驱动的缓存更新
- **预取**：预测性密钥加载

#### 资源优化

- **连接复用**：HTTP 客户端池
- **内存管理**：高效的数据结构
- **垃圾回收**：优化的对象生命周期

### 🛡️ 安全架构

#### 加密

- **AES-256-GCM**：行业标准加密
- **密钥派生**：PBKDF2 密钥强化
- **盐生成**：密码学安全随机数

#### 访问控制

- **API 认证**：Bearer token 验证
- **基于角色**：管理员和用户访问级别
- **审计日志**：完整的请求可追溯性

### 🔧 配置管理

#### 热重载

- **文件监控**：自动配置更新
- **优雅更新**：零停机配置变更
- **验证**：应用前的模式验证

#### 环境支持

- **多环境**：开发、测试、生产
- **覆盖链**：环境 → 文件 → 默认值
- **敏感数据**：机密信息的安全处理

</details>

## 🎯 高级功能

### 🔄 API 密钥状态机

GPT-Load 实现了复杂的 5 状态密钥生命周期管理系统：

#### 状态定义

- **待验证**：新添加的密钥，等待首次验证
- **活跃**：已验证且完全可用的密钥
- **降级**：部分失败但仍可使用的密钥
- **禁用**：连续失败被暂时禁用的密钥（指数退避）
- **无效**：永久无效需要手动干预的密钥

#### 状态转换

- **自动转换**：基于验证结果和请求成功/失败模式
- **手动转换**：支持手动启用/禁用/无效化操作
- **智能逻辑**：考虑失败次数、错误类型和历史性能

#### 指数退避

- **渐进延迟**：1 分钟 → 2 分钟 → 4 分钟 → ... → 最大 30 分钟
- **自动恢复**：禁用密钥在退避期后自动转换为降级状态
- **可配置阈值**：自定义失败次数和退避参数

### 🎯 策略引擎与组合策略

#### 组合式架构

- **可复用策略**：创建可应用于多个分组的策略模板
- **策略类型**：重试策略、降级策略、模型过滤策略
- **优先级系统**：按优先级顺序评估策略，采用首次匹配语义

#### 高级重试策略

```json
{
  "name": "智能重试策略",
  "type": "retry",
  "rules": [
    {
      "conditions": [
        {
          "type": "error_message",
          "operator": "contains",
          "value": "rate limit"
        }
      ],
      "action": "retry",
      "max_retries": 3,
      "backoff_ms": 2000
    },
    {
      "conditions": [
        {
          "type": "error_message",
          "operator": "contains",
          "value": "invalid api key"
        }
      ],
      "action": "invalidate"
    }
  ]
}
```

#### 模型过滤

- **包含/排除模式**：支持白名单和黑名单方式
- **正则表达式和通配符**：灵活的模型名称模式匹配
- **分组配置**：不同分组的不同过滤规则

### 📊 增量验证系统

#### 智能过滤

- **时间窗口过滤**：仅验证指定时间范围内添加的密钥
- **状态过滤**：针对特定密钥状态（待验证、无效等）
- **去重处理**：排除最近已验证的密钥，避免重复检查

#### 高性能处理

- **批量处理**：可配置的批大小，实现高效的数据库操作
- **并发验证**：并行密钥验证，可配置并发数
- **上下文管理**：支持取消和超时控制

#### RESTful API

```bash
# 使用自定义参数验证特定分组
POST /api/validation/groups/{groupId}
{
  "time_window_hours": 24,
  "include_states": ["pending", "invalid"],
  "exclude_recently_validated": true,
  "recent_validation_hours": 1,
  "concurrency": 5,
  "batch_size": 100
}

# 获取验证历史
GET /api/validation/groups/{groupId}/history?hours=24
```

## 🔌 API 使用指南

<details>
<summary>代理接口调用</summary>

GPT-Load 通过分组名称将请求路由到不同的 AI 服务。使用方法如下：

### 1. 代理端点格式

```text
http://localhost:3001/proxy/{group_name}/{original_api_path}
```

- `{group_name}`：在管理界面中创建的分组名称
- `{original_api_path}`：与原始 AI 服务路径保持完全一致

### 2. 认证方式

在 Web 管理界面中配置**代理密钥**，支持系统级和分组级代理密钥。

- **认证方式**：与原生 API 一致，但将原始密钥替换为配置的代理密钥
- **密钥范围**：系统设置中配置的**全局代理密钥**可在所有分组中使用。分组中配置的**分组代理密钥**仅对当前分组有效
- **格式**：多个密钥用逗号分隔

### 3. OpenAI 接口示例

假设创建了名为 `openai` 的分组：

**原始调用：**

```bash
curl -X POST https://api.openai.com/v1/chat/completions \
  -H "Authorization: Bearer sk-your-openai-key" \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4o-mini", "messages": [{"role": "user", "content": "Hello"}]}'
```

**代理调用：**

```bash
curl -X POST http://localhost:3001/proxy/openai/v1/chat/completions \
  -H "Authorization: Bearer your-proxy-key" \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4o-mini", "messages": [{"role": "user", "content": "Hello"}]}'
```

**需要修改的地方：**

- 将 `https://api.openai.com` 替换为 `http://localhost:3001/proxy/openai`
- 将原始 API Key 替换为**代理密钥**

### 4. Gemini 接口示例

假设创建了名为 `gemini` 的分组：

**原始调用：**

```bash
curl -X POST https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent?key=your-gemini-key \
  -H "Content-Type: application/json" \
  -d '{"contents": [{"parts": [{"text": "Hello"}]}]}'
```

**代理调用：**

```bash
curl -X POST http://localhost:3001/proxy/gemini/v1beta/models/gemini-2.0-flash-exp:generateContent?key=your-proxy-key \
  -H "Content-Type: application/json" \
  -d '{"contents": [{"parts": [{"text": "Hello"}]}]}'
```

**需要修改的地方：**

- 将 `https://generativelanguage.googleapis.com` 替换为 `http://localhost:3001/proxy/gemini`
- 将 URL 参数中的 `key=your-gemini-key` 替换为**代理密钥**

### 5. Anthropic 接口示例

假设创建了名为 `anthropic` 的分组：

**原始调用：**

```bash
curl -X POST https://api.anthropic.com/v1/messages \
  -H "x-api-key: sk-ant-api03-your-anthropic-key" \
  -H "anthropic-version: 2023-06-01" \
  -H "Content-Type: application/json" \
  -d '{"model": "claude-3-5-sonnet-20241022", "messages": [{"role": "user", "content": "Hello"}]}'
```

**代理调用：**

```bash
curl -X POST http://localhost:3001/proxy/anthropic/v1/messages \
  -H "x-api-key: your-proxy-key" \
  -H "anthropic-version: 2023-06-01" \
  -H "Content-Type: application/json" \
  -d '{"model": "claude-3-5-sonnet-20241022", "messages": [{"role": "user", "content": "Hello"}]}'
```

**需要修改的地方：**

- 将 `https://api.anthropic.com` 替换为 `http://localhost:3001/proxy/anthropic`
- 将 `x-api-key` 头中的原始 API Key 替换为**代理密钥**

### 6. 支持的接口

**OpenAI 格式：**

- `/v1/chat/completions` - 聊天对话
- `/v1/completions` - 文本补全
- `/v1/embeddings` - 文本嵌入
- `/v1/models` - 模型列表
- 以及所有其他 OpenAI 兼容接口

**Gemini 格式：**

- `/v1beta/models/*/generateContent` - 内容生成
- `/v1beta/models` - 模型列表
- 以及所有其他 Gemini 原生接口

**Anthropic 格式：**

- `/v1/messages` - 消息对话
- `/v1/models` - 模型列表（如可用）
- 以及所有其他 Anthropic 原生接口

### 7. 客户端 SDK 配置

**OpenAI Python SDK：**

```python
from openai import OpenAI

client = OpenAI(
    api_key="your-proxy-key",  # 使用代理密钥
    base_url="http://localhost:3001/proxy/openai"  # 使用代理端点
)

response = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "Hello"}]
)
```

**Google Gemini SDK (Python)：**

```python
import google.generativeai as genai

# 配置 API 密钥和基础 URL
genai.configure(
    api_key="your-proxy-key",  # 使用代理密钥
    client_options={"api_endpoint": "http://localhost:3001/proxy/gemini"}
)

model = genai.GenerativeModel('gemini-2.0-flash-exp')
response = model.generate_content("Hello")
```

**Anthropic SDK (Python)：**

```python
from anthropic import Anthropic

client = Anthropic(
    api_key="your-proxy-key",  # 使用代理密钥
    base_url="http://localhost:3001/proxy/anthropic"  # 使用代理端点
)

response = client.messages.create(
    model="claude-3-5-sonnet-20241022",
    messages=[{"role": "user", "content": "Hello"}]
)
```

> **重要说明**：作为透明代理服务，GPT-Load 完全保留各种 AI 服务的原生 API 格式和认证方式。您只需要替换端点地址并使用管理界面中配置的**代理密钥**即可无缝迁移。

</details>

## 🔧 管理 API 参考

<details>
<summary>管理 API 端点</summary>

### 认证

所有管理 API 都需要使用环境中配置的 `AUTH_KEY` 进行认证。

```bash
curl -H "Authorization: Bearer your-auth-key" http://localhost:3001/api/...
```

### 密钥验证 API

#### 增量验证

```bash
# 验证特定分组
POST /api/validation/groups/{groupId}
Content-Type: application/json
{
  "time_window_hours": 24,           // 仅验证最近24小时添加的密钥
  "include_states": ["pending", "invalid"],  // 针对特定密钥状态
  "exclude_recently_validated": true, // 跳过最近验证过的密钥
  "recent_validation_hours": 1,      // 定义"最近"为1小时
  "concurrency": 5,                  // 并发验证工作线程
  "batch_size": 100                  // 批处理大小
}

# 响应
{
  "code": 0,
  "message": "Success",
  "data": {
    "group_id": 1,
    "group_name": "OpenAI Group",
    "total_keys": 50,
    "validated_keys": 45,
    "successful_keys": 40,
    "failed_keys": 5,
    "duration_ms": 2500
  }
}
```

#### 验证历史

```bash
# 获取分组的验证历史
GET /api/validation/groups/{groupId}/history?hours=24

# 响应
{
  "code": 0,
  "message": "Success",
  "data": {
    "group_id": 1,
    "time_range": 24,
    "history": {
      "total_keys": 100,
      "validated_keys": 85,
      "pending_keys": 10,
      "active_keys": 70,
      "invalid_keys": 20,
      "validation_rate": 0.85
    }
  }
}
```

#### 验证所有分组

```bash
# 触发所有分组的验证
POST /api/validation/groups
Content-Type: application/json
{
  "time_window_hours": 1,
  "include_states": ["pending"],
  "concurrency": 3,
  "batch_size": 50
}

# 响应
{
  "code": 0,
  "message": "Success",
  "data": {
    "results": [...],  // 各分组结果
    "summary": {
      "total_groups": 5,
      "total_keys": 250,
      "total_validated": 200,
      "total_successful": 180
    }
  }
}
```

### 分组管理 API

#### 列出分组

```bash
GET /api/groups

# 响应
{
  "code": 0,
  "data": {
    "groups": [
      {
        "id": 1,
        "name": "openai",
        "channel_type": "openai",
        "upstreams": [...],
        "key_count": 10,
        "active_key_count": 8
      }
    ]
  }
}
```

#### 分组统计

```bash
GET /api/groups/{id}/stats

# 响应
{
  "code": 0,
  "data": {
    "total_requests": 1000,
    "successful_requests": 950,
    "failed_requests": 50,
    "success_rate": 0.95,
    "avg_response_time": 1200,
    "key_stats": {
      "total": 10,
      "active": 8,
      "degraded": 1,
      "disabled": 0,
      "invalid": 1
    }
  }
}
```

### 密钥管理 API

#### 批量密钥操作

```bash
# 添加多个密钥
POST /api/keys/add-multiple
{
  "group_id": 1,
  "keys": ["sk-key1", "sk-key2", "sk-key3"]
}

# 删除多个密钥
POST /api/keys/delete-multiple
{
  "group_id": 1,
  "key_values": ["sk-key1", "sk-key2"]
}

# 导出密钥
GET /api/keys/export?group_id=1&format=json
```

### 系统 API

#### 仪表板统计

```bash
GET /api/dashboard/stats

# 响应
{
  "code": 0,
  "data": {
    "total_requests": 10000,
    "success_rate": 0.95,
    "active_groups": 5,
    "total_keys": 50,
    "system_health": "healthy"
  }
}
```

#### 健康检查

```bash
GET /health

# 响应
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "1.0.0"
}
```

</details>

## 🏗️ 架构与设计

### 🔧 核心架构

GPT-Load 采用现代化、可扩展的架构，专为企业环境设计：

#### 分层架构

- **表示层**：Vue 3 Web 界面和 RESTful API
- **业务层**：策略引擎、密钥管理、验证服务
- **数据层**：数据库抽象，支持多种后端
- **基础设施层**：缓存、日志、监控和网络

#### 依赖注入

- **基于容器的 DI**：所有服务通过依赖注入管理
- **接口抽象**：接口与实现的清晰分离
- **可测试性**：通过接口契约实现简单的模拟和单元测试

#### 微服务就绪设计

- **服务分离**：不同功能区域之间的清晰边界
- **事件驱动**：非阻塞操作的异步处理
- **无状态组件**：共享存储支持的水平扩展

### 🔄 密钥管理生命周期

```mermaid
stateDiagram-v2
    [*] --> Pending: 添加密钥
    Pending --> Active: 验证成功
    Pending --> Invalid: 验证失败（3次）
    Active --> Degraded: 请求失败
    Degraded --> Active: 请求成功
    Degraded --> Disabled: 连续失败
    Disabled --> Degraded: 退避过期
    Active --> Invalid: 手动操作
    Degraded --> Invalid: 手动操作
    Invalid --> [*]: 删除密钥
```

### 🎯 策略引擎流程

```mermaid
flowchart TD
    A[请求失败] --> B[获取分组策略]
    B --> C[按优先级排序]
    C --> D[评估条件]
    D --> E{找到匹配?}
    E -->|是| F[执行动作]
    E -->|否| G[默认处理]
    F --> H[更新密钥状态]
    G --> H
    H --> I[记录结果]
```

### 📊 性能特性

#### 吞吐量

- **并发请求**：支持 1000+ 并发连接
- **密钥验证**：100+ 密钥/秒验证吞吐量
- **数据库操作**：批量操作的优化批处理

#### 延迟

- **代理延迟**：请求转发 <5ms 开销
- **密钥选择**：密钥池操作 <1ms
- **状态更新**：异步处理避免阻塞

#### 扩展性

- **水平扩展**：主从架构
- **数据库扩展**：支持只读副本和连接池
- **缓存**：基于 Redis 的热数据缓存

## 🧪 开发与测试

### 🔧 开发设置

```bash
# 克隆仓库
git clone https://github.com/ShangjinTang/gpt-load.git
cd gpt-load

# 安装依赖
go mod tidy

# 设置环境
cp .env.example .env
# 编辑 .env 配置

# 开发模式运行
make dev

# 运行测试
make test

# 检查测试覆盖率
make test-coverage
```

### 🧪 测试策略

#### 单元测试

- **服务层**：业务逻辑的全面单元测试
- **模拟依赖**：基于接口的模拟，实现隔离测试
- **覆盖率目标**：关键组件 >80% 代码覆盖率

#### 集成测试

- **数据库测试**：快速集成测试的内存 SQLite
- **API 测试**：使用测试夹具的完整 HTTP API 测试
- **策略测试**：复杂策略评估场景

#### 性能测试

- **负载测试**：并发请求处理验证
- **内存测试**：内存泄漏检测和优化
- **基准测试**：性能回归检测

### 🔍 代码质量

#### 静态分析

- **Go Vet**：内置 Go 静态分析
- **Linting**：golangci-lint 全面代码质量检查
- **安全**：gosec 安全漏洞扫描

#### 代码组织

- **清洁架构**：关注点的清晰分离
- **接口设计**：依赖倒置原则
- **错误处理**：全面的错误类型和处理

## 🤝 贡献

感谢所有为 GPT-Load 做出贡献的开发者！

[![Contributors](https://contrib.rocks/image?repo=ShangjinTang/gpt-load)](https://github.com/ShangjinTang/gpt-load/graphs/contributors)

### 📋 贡献指南

我们欢迎贡献！请遵循以下指南：

#### 🐛 错误报告

- 使用问题模板
- 包含重现步骤
- 提供系统信息
- 添加相关日志

#### ✨ 功能请求

- 描述使用场景
- 解释预期行为
- 考虑向后兼容性
- 讨论实现方法

#### 🔧 拉取请求

- Fork 仓库
- 创建功能分支
- 遵循编码标准
- 为新功能添加测试
- 更新文档
- 确保 CI 通过

#### 📚 文档

- 为新功能更新 README
- 为复杂逻辑添加代码注释
- 包含 API 文档
- 提供使用示例

## 🔒 安全

### 🛡️ 安全功能

- **数据加密**：敏感数据的 AES 静态加密
- **认证**：安全的基于令牌的认证
- **输入验证**：全面的输入清理
- **速率限制**：防止滥用和 DoS 攻击

### 🚨 报告安全问题

请私下向维护者报告安全漏洞。不要使用公共问题报告安全问题。

## 📄 许可证

MIT 许可证 - 详情请参见 [LICENSE](LICENSE) 文件。

## ⭐ Star 历史

[![Stargazers over time](https://starchart.cc/ShangjinTang/gpt-load.svg?variant=adaptive)](https://starchart.cc/ShangjinTang/gpt-load)

---

## 🙏 致谢

特别感谢：

- 所有帮助改进 GPT-Load 的贡献者
- Go 社区提供的优秀工具和库
- 提供反馈和功能请求的用户

---

_由 GPT-Load 团队用 ❤️ 构建_

## 📌 Fork 信息

本仓库 fork 自 [https://github.com/tbphp/gpt-load](https://github.com/tbphp/gpt-load) 的 v1.2.1 版本。
