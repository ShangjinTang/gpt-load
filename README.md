# GPT-Load

English | [中文](README_CN.md)

[![Release](https://img.shields.io/github/v/release/ShangjinTang/gpt-load)](https://github.com/ShangjinTang/gpt-load/releases) ![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg) [![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A high-performance, enterprise-grade AI API transparent proxy service designed specifically for enterprises and developers who need to integrate multiple AI services. Built with Go, featuring intelligent key management, advanced policy engine, model filtering, incremental validation, and comprehensive monitoring capabilities, designed for high-concurrency production environments.

For detailed documentation, please visit [Official Documentation](https://www.gpt-load.com/docs?lang=en)

<a href="https://trendshift.io/repositories/14880" target="_blank"><img src="https://trendshift.io/api/badge/repositories/14880" alt="ShangjinTang%2Fgpt-load | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a> <a href="https://hellogithub.com/repository/ShangjinTang/gpt-load" target="_blank"><img src="https://api.hellogithub.com/v1/widgets/recommend.svg?rid=554dc4c46eb14092b9b0c56f1eb9021c&claim_uid=Qlh8vzrWJ0HCneG" alt="Featured｜HelloGitHub" style="width: 250px; height: 54px;" width="250" height="54" /></a>

## 🚀 Features

### Core Proxy Features

- **Transparent Proxy**: Complete preservation of native API formats, supporting OpenAI, Google Gemini, and Anthropic Claude among other formats
- **High-Performance Design**: Zero-copy streaming, connection pool reuse, and atomic operations
- **Load Balancing**: Weighted load balancing across multiple upstream endpoints to enhance service availability
- **Graceful Shutdown**: Production-ready graceful shutdown and error recovery mechanisms

### 🔑 Advanced Key Management

- **Intelligent Key Pool**: Group-based management with automatic rotation and failure recovery
- **Smart State Machine**: 5-state key lifecycle (Pending → Active → Degraded → Disabled → Invalid)
- **Exponential Backoff**: Automatic retry with intelligent backoff for failed keys
- **Incremental Validation**: Time-window based validation to avoid redundant key checks
- **Batch Operations**: High-performance bulk key import, validation, and management

### 🎯 Policy Engine & Filtering

- **Composition-Based Policies**: Reusable policy templates that can be applied to multiple groups
- **Advanced Retry Policies**: Error-code and message-based retry strategies with customizable actions
- **Model Filtering**: Include/exclude model patterns with regex and wildcard support
- **Dynamic Policy Evaluation**: Real-time policy matching and execution based on request context

### 🏗️ Enterprise Architecture

- **Distributed Deployment**: Leader-follower architecture supporting horizontal scaling and high availability
- **Hot Configuration Reload**: System settings and group configurations support hot-reload without requiring restarts
- **Comprehensive Monitoring**: Real-time statistics, health checks, and detailed request logging
- **API Management**: RESTful API for configuration and monitoring
- **Dual Authentication**: Separate authentication for API management and proxy, with proxy authentication supporting global and group-level keys

### 🔒 Security & Reliability

- **Data Encryption**: Optional AES encryption for API keys at rest with migration support
- **Smart Failure Handling**: Automatic key blacklist management and recovery mechanisms
- **Rate Limiting**: Configurable request rate limiting and concurrent request control
- **Request Logging**: Detailed audit trails with configurable retention policies

## 🎯 Supported AI Services

GPT-Load serves as a transparent proxy service, completely preserving the native API formats of various AI service providers:

- **OpenAI Format**: Official OpenAI API, Azure OpenAI, and other OpenAI-compatible services
- **Google Gemini Format**: Native APIs for Gemini Pro, Gemini Pro Vision, and other models
- **Anthropic Claude Format**: Claude series models, supporting high-quality conversations and text generation

## 📋 System Requirements

- Go 1.23+ (for source builds)
- Docker (for containerized deployment)
- PostgreSQL or SQLite (for database storage)
- Redis (for caching and distributed coordination, optional)

## 🚀 Quick Start

### Method 1: Docker Quick Start

```bash
docker run -d --name gpt-load \
    -p 3001:3001 \
    -e AUTH_KEY=your-secure-key-here \
    -v "$(pwd)/data":/app/data \
    ghcr.io/ShangjinTang/gpt-load:latest
```

> Please change `your-secure-key-here` to a strong password (never use the default value), then you can access the management API at: <http://localhost:3001>

### Method 2: Using Docker Compose (Recommended)

**Installation Commands:**

```bash
# Create Directory
mkdir -p gpt-load && cd gpt-load

# Download configuration files
wget https://raw.githubusercontent.com/ShangjinTang/gpt-load/refs/heads/main/docker-compose.yml
wget -O .env https://raw.githubusercontent.com/ShangjinTang/gpt-load/refs/heads/main/.env.example

# Edit the .env file and change AUTH_KEY to a strong password. Never use default or simple keys like sk-123456.

# Start services
docker compose up -d
```

Before deployment, you must change the default admin key (AUTH_KEY). A recommended format is: sk-prod-[32-character random string].

The default installation uses the SQLite version, which is suitable for lightweight, single-instance applications.

If you need to install PostgreSQL and Redis, please uncomment the required services in the `docker-compose.yml` file, configure the corresponding environment variables, and restart.

**Other Commands:**

```bash
# Check service status
docker compose ps

# View logs
docker compose logs -f

# Restart Service
docker compose down && docker compose up -d

# Update to latest version
docker compose pull && docker compose down && docker compose up -d
```

After deployment:

- Access Management API: <http://localhost:3001>
- API Proxy Address: <http://localhost:3001/proxy>

> Use your modified AUTH_KEY to access the management API.

### Method 3: Source Build

Source build requires a locally installed database (SQLite or PostgreSQL) and Redis (optional).

```bash
# Clone and build
git clone https://github.com/ShangjinTang/gpt-load.git
cd gpt-load
go mod tidy

# Create configuration
cp .env.example .env

# Edit the .env file and change AUTH_KEY to a strong password. Never use default or simple keys like sk-123456.
# Modify DATABASE_DSN and REDIS_DSN configurations in .env
# REDIS_DSN is optional; if not configured, memory storage will be enabled

# Run
make run
```

After deployment:

- Access Management API: <http://localhost:3001>
- API Proxy Address: <http://localhost:3001/proxy>

> Use your modified AUTH_KEY to access the management API.

### Method 4: Cluster Deployment

Cluster deployment requires all nodes to connect to the same PostgreSQL and Redis, with Redis being mandatory. It's recommended to use unified distributed PostgreSQL and Redis clusters.

**Deployment Requirements:**

- All nodes must configure identical `AUTH_KEY`, `DATABASE_DSN`, `REDIS_DSN`
- Leader-follower architecture where follower nodes must configure environment variable: `IS_SLAVE=true`

For details, please refer to [Cluster Deployment Documentation](https://www.gpt-load.com/docs/cluster?lang=en)

## ⚙️ Configuration System

### Configuration Architecture Overview

GPT-Load adopts a dual-layer configuration architecture:

#### 1. Static Configuration (Environment Variables)

- **Characteristics**: Read at application startup, immutable during runtime, requires application restart to take effect
- **Purpose**: Infrastructure configuration such as database connections, server ports, authentication keys, etc.
- **Management**: Set via `.env` files or system environment variables

#### 2. Dynamic Configuration (Hot-Reload)

- **System Settings**: Stored in database, providing unified behavioral standards for the entire application
- **Group Configuration**: Behavior parameters customized for specific groups, can override system settings
- **Configuration Priority**: Group Configuration > System Settings > Environment Configuration
- **Characteristics**: Supports hot-reload, takes effect immediately after modification without application restart

<details>
<summary>Static Configuration (Environment Variables)</summary>

**Server Configuration:**

| Setting | Environment Variable | Default | Description |
| --- | --- | --- | --- |
| Service Port | `PORT` | 3001 | HTTP server listening port |
| Service Address | `HOST` | 0.0.0.0 | HTTP server binding address |
| Read Timeout | `SERVER_READ_TIMEOUT` | 60 | HTTP server read timeout (seconds) |
| Write Timeout | `SERVER_WRITE_TIMEOUT` | 600 | HTTP server write timeout (seconds) |
| Idle Timeout | `SERVER_IDLE_TIMEOUT` | 120 | HTTP connection idle timeout (seconds) |
| Graceful Shutdown Timeout | `SERVER_GRACEFUL_SHUTDOWN_TIMEOUT` | 10 | Service graceful shutdown wait time (seconds) |
| Follower Mode | `IS_SLAVE` | false | Follower node identifier for cluster deployment |
| Timezone | `TZ` | `Asia/Shanghai` | Specify timezone |

**Security Configuration:**

| Setting | Environment Variable | Default | Description |
| --- | --- | --- | --- |
| Admin Key | `AUTH_KEY` | - | Access authentication key for the **management API**, please change it to a strong password |
| Encryption Key | `ENCRYPTION_KEY` | - | Encrypts API keys at rest. Supports any string or leave empty to disable encryption. ⚠️ **Once set, do not change directly. See [Encryption Key Migration](#encryption-key-migration) if you need to change it.** |

**Database Configuration:**

| Setting | Environment Variable | Default | Description |
| --- | --- | --- | --- |
| Database Connection | `DATABASE_DSN` | `./data/gpt-load.db` | Database connection string (DSN) or file path |
| Redis Connection | `REDIS_DSN` | - | Redis connection string, uses memory storage when empty |

**Performance & CORS Configuration:**

| Setting | Environment Variable | Default | Description |
| --- | --- | --- | --- |
| Max Concurrent Requests | `MAX_CONCURRENT_REQUESTS` | 100 | Maximum concurrent requests allowed by system |
| Enable CORS | `ENABLE_CORS` | false | Whether to enable Cross-Origin Resource Sharing |
| Allowed Origins | `ALLOWED_ORIGINS` | - | Allowed origins, comma-separated |
| Allowed Methods | `ALLOWED_METHODS` | `GET,POST,PUT,DELETE,OPTIONS` | Allowed HTTP methods |
| Allowed Headers | `ALLOWED_HEADERS` | `*` | Allowed request headers, comma-separated |
| Allow Credentials | `ALLOW_CREDENTIALS` | false | Whether to allow sending credentials |

**Logging Configuration:**

| Setting | Environment Variable | Default | Description |
| --- | --- | --- | --- |
| Log Level | `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| Log Format | `LOG_FORMAT` | `text` | Log format: text, json |
| Enable File Logging | `LOG_ENABLE_FILE` | false | Whether to enable file log output |
| Log File Path | `LOG_FILE_PATH` | `./data/logs/app.log` | Log file storage path |

**Proxy Configuration:**

GPT-Load automatically reads proxy settings from environment variables to make requests to upstream AI providers.

| Setting | Environment Variable | Default | Description |
| --- | --- | --- | --- |
| HTTP Proxy | `HTTP_PROXY` | - | Proxy server address for HTTP requests |
| HTTPS Proxy | `HTTPS_PROXY` | - | Proxy server address for HTTPS requests |
| No Proxy | `NO_PROXY` | - | Comma-separated list of hosts or domains to bypass the proxy |

Supported Proxy Protocol Formats:

- **HTTP**: `http://user:pass@host:port`
- **HTTPS**: `https://user:pass@host:port`
- **SOCKS5**: `socks5://user:pass@host:port`
</details>

<details>
<summary>Dynamic Configuration (Hot-Reload)</summary>

**Basic Settings:**

| Setting | Field Name | Default | Group Override | Description |
| --- | --- | --- | --- | --- |
| Project URL | `app_url` | `http://localhost:3001` | ❌ | Project base URL |
| Global Proxy Keys | `proxy_keys` | Initial value from `AUTH_KEY` | ❌ | Globally effective proxy keys, comma-separated |
| Log Retention Days | `request_log_retention_days` | 7 | ❌ | Request log retention days, 0 for no cleanup |
| Log Write Interval | `request_log_write_interval_minutes` | 1 | ❌ | Log write to database cycle (minutes) |
| Enable Request Body Logging | `enable_request_body_logging` | false | ✅ | Whether to log complete request body content in request logs |

**Request Settings:**

| Setting | Field Name | Default | Group Override | Description |
| --- | --- | --- | --- | --- |
| Request Timeout | `request_timeout` | 600 | ✅ | Forward request complete lifecycle timeout (seconds) |
| Connection Timeout | `connect_timeout` | 15 | ✅ | Timeout for establishing connection with upstream service (seconds) |
| Idle Connection Timeout | `idle_conn_timeout` | 120 | ✅ | HTTP client idle connection timeout (seconds) |
| Response Header Timeout | `response_header_timeout` | 600 | ✅ | Timeout for waiting upstream response headers (seconds) |
| Max Idle Connections | `max_idle_conns` | 100 | ✅ | Connection pool maximum total idle connections |
| Max Idle Connections Per Host | `max_idle_conns_per_host` | 50 | ✅ | Maximum idle connections per upstream host |
| Proxy URL | `proxy_url` | - | ✅ | HTTP/HTTPS proxy for forwarding requests, uses environment if empty |

**Key Configuration:**

| Setting | Field Name | Default | Group Override | Description |
| --- | --- | --- | --- | --- |
| Max Retries | `max_retries` | 3 | ✅ | Maximum retry count using different keys for single request |
| Blacklist Threshold | `blacklist_threshold` | 3 | ✅ | Number of consecutive failures before key enters blacklist |
| Key Validation Interval | `key_validation_interval_minutes` | 60 | ✅ | Background scheduled key validation cycle (minutes) |
| Key Validation Concurrency | `key_validation_concurrency` | 10 | ✅ | Concurrency for background validation of invalid keys |
| Key Validation Timeout | `key_validation_timeout_seconds` | 20 | ✅ | API request timeout for validating individual keys in background (seconds) |

</details>

## 🔐 Encryption Key Migration

If you need to change your encryption key after it has been set, follow these steps carefully:

<details>
<summary>View Encryption Key Migration Guide</summary>

### ⚠️ Important Warnings

- **Service Downtime Required**: The service must be stopped during migration
- **Backup First**: Always backup your database before migration
- **No Recovery**: If you lose the current encryption key, encrypted data cannot be recovered
- **Test First**: Test the migration process on a copy of your database first

### Migration Steps

1. **Stop the Service**

   ```bash
   # Docker Compose
   docker compose down

   # Or if running directly
   # Stop the running process (Ctrl+C or kill process)
   ```

2. **Backup Database**

   ```bash
   # For SQLite (default)
   cp ./data/gpt-load.db ./data/gpt-load.db.backup

   # For PostgreSQL
   pg_dump your_database > backup.sql
   ```

3. **Prepare Migration Script**

   Copy the migration script from `scripts/migrate_encryption_key.go` and modify the configuration:

   ```go
   const (
       DATABASE_DSN = "./data/gpt-load.db"  // Your database connection
       OLD_ENCRYPTION_KEY = "current-key"   // Your current key
       NEW_ENCRYPTION_KEY = "new-32-char-secret-key"  // Your new key
   )
   ```

4. **Run Migration**

   ```bash
   cd scripts
   go run migrate_encryption_key.go
   ```

5. **Update Configuration**

   ```bash
   # Update .env file
   echo "ENCRYPTION_KEY=new-32-char-secret-key" >> .env
   ```

6. **Restart Service**

   ```bash
   # Docker Compose
   docker compose up -d

   # Or direct run
   make run
   ```

### Migration Scenarios

- **Enable Encryption**: Set `OLD_ENCRYPTION_KEY = ""` and provide `NEW_ENCRYPTION_KEY`
- **Disable Encryption**: Provide current key as `OLD_ENCRYPTION_KEY` and set `NEW_ENCRYPTION_KEY = ""`
- **Change Key**: Provide both old and new encryption keys

### Troubleshooting

If migration fails:

1. Restore from backup: `cp ./data/gpt-load.db.backup ./data/gpt-load.db`
2. Check the error message in the migration script output
3. Verify your database connection string and encryption keys
4. Ensure the service is completely stopped

### Key Generation

Generate a secure 32-character key:

```bash
openssl rand -base64 32 | tr -d "=+/" | cut -c1-32
```

</details>

## 🏗️ Architecture Design

GPT-Load is built with a modular, high-performance architecture designed for enterprise-grade AI proxy services.

<details>
<summary>View Architecture Overview</summary>

### 🎯 Core Design Principles

- **High Performance**: Asynchronous processing, connection pooling, and intelligent caching
- **Scalability**: Horizontal scaling support with stateless design
- **Reliability**: Circuit breaker patterns, graceful degradation, and comprehensive error handling
- **Security**: Encryption at rest, secure key management, and audit logging
- **Maintainability**: Clean architecture, dependency injection, and comprehensive testing

### 🏛️ System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client Apps   │    │  Load Balancer  │    │   Monitoring    │
└─────────┬───────┘    └─────────┬───────┘    └─────────────────┘
          │                      │
          └──────────┐          ┌┴──────────┐
                     │          │           │
                ┌────▼──────────▼───┐   ┌───▼────┐
                │    GPT-Load       │   │ Metrics│
                │   Proxy Server    │   │ & Logs │
                └────┬──────────────┘   └────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
   ┌────▼───┐   ┌───▼────┐   ┌───▼────┐
   │ OpenAI │   │ Claude │   │ Others │
   │   API  │   │   API  │   │  APIs  │
   └────────┘   └────────┘   └────────┘
```

### 🔧 Component Architecture

#### 1. **HTTP Layer** (`internal/router`, `internal/handler`)
- **Gin Framework**: High-performance HTTP routing
- **Middleware Stack**: CORS, authentication, rate limiting, logging
- **API Versioning**: RESTful API design with version support

#### 2. **Proxy Engine** (`internal/proxy`)
- **Request Routing**: Intelligent upstream selection
- **Load Balancing**: Round-robin and weighted algorithms
- **Circuit Breaker**: Automatic failure detection and recovery
- **Response Processing**: Stream handling and format conversion

#### 3. **Key Management** (`internal/keypool`)
- **State Machine**: 5-state lifecycle (Pending → Active → Degraded → Disabled → Invalid)
- **Rotation Algorithm**: Atomic key selection and rotation
- **Health Monitoring**: Background validation and automatic recovery
- **Encryption**: AES-256-GCM encryption for keys at rest

#### 4. **Policy Engine** (`internal/policy`)
- **Rule-Based Logic**: Flexible retry and degradation policies
- **Pattern Matching**: Regex-based error classification
- **Action Execution**: Automatic key state transitions
- **Composition Strategy**: Layered policy inheritance

#### 5. **Data Layer** (`internal/db`, `internal/models`)
- **ORM**: GORM for database abstraction
- **Multi-Database**: SQLite and PostgreSQL support
- **Migrations**: Automatic schema management
- **Connection Pooling**: Optimized database connections

#### 6. **Caching Layer** (`internal/store`)
- **Multi-Backend**: Memory and Redis support
- **Key Distribution**: Consistent hashing for scalability
- **TTL Management**: Automatic expiration and cleanup
- **Pub/Sub**: Real-time configuration updates

#### 7. **Background Services** (`internal/services`)
- **Validation Service**: Periodic key health checks
- **Log Service**: Asynchronous request logging
- **State Service**: Key lifecycle management
- **Incremental Validation**: Smart validation with skip logic

### 📊 Data Flow

#### Request Processing Flow
```
1. Client Request → 2. Authentication → 3. Rate Limiting
         ↓                    ↓                 ↓
4. Key Selection ← 5. Load Balancing ← 6. Upstream Routing
         ↓                    ↓                 ↓
7. API Call → 8. Response Processing → 9. Client Response
         ↓                    ↓                 ↓
10. Logging ← 11. Key Status Update ← 12. Policy Evaluation
```

#### Key Lifecycle Flow
```
New Key → Validation → Active → [Success/Failure] → Policy Check
    ↓         ↓          ↓              ↓              ↓
  Pending → Invalid → Degraded → Disabled → Recovery/Removal
```

### 🔄 State Management

#### API Key States
- **Pending**: Newly added, awaiting validation
- **Active**: Validated and available for requests
- **Degraded**: Experiencing issues but still usable
- **Disabled**: Temporarily disabled due to failures
- **Invalid**: Permanently invalid, excluded from rotation

#### State Transitions
- **Smart Backoff**: Exponential backoff for failed keys
- **Auto Recovery**: Automatic re-validation and state restoration
- **Policy-Driven**: Custom rules for state transitions

### 🚀 Performance Features

#### Concurrency & Parallelism
- **Goroutine Pools**: Controlled concurrency for validation
- **Channel-Based**: Non-blocking communication
- **Context Cancellation**: Graceful shutdown support

#### Caching Strategy
- **Multi-Level**: L1 (memory) + L2 (Redis) caching
- **Smart Invalidation**: Event-driven cache updates
- **Prefetching**: Predictive key loading

#### Resource Optimization
- **Connection Reuse**: HTTP client pooling
- **Memory Management**: Efficient data structures
- **Garbage Collection**: Optimized object lifecycle

### 🛡️ Security Architecture

#### Encryption
- **AES-256-GCM**: Industry-standard encryption
- **Key Derivation**: PBKDF2 for key strengthening
- **Salt Generation**: Cryptographically secure randomness

#### Access Control
- **API Authentication**: Bearer token validation
- **Role-Based**: Admin and user access levels
- **Audit Logging**: Complete request traceability

### 🔧 Configuration Management

#### Hot Reload
- **File Watching**: Automatic configuration updates
- **Graceful Updates**: Zero-downtime configuration changes
- **Validation**: Schema validation before applying changes

#### Environment Support
- **Multi-Environment**: Development, staging, production
- **Override Chain**: Environment → File → Defaults
- **Sensitive Data**: Secure handling of secrets

</details>

## 🎯 Advanced Features

### 🔄 API Key State Machine

GPT-Load implements a sophisticated 5-state key lifecycle management system:

#### State Definitions

- **Pending**: Newly added keys awaiting first validation
- **Active**: Validated and fully operational keys
- **Degraded**: Keys with partial failures but still usable
- **Disabled**: Keys temporarily disabled due to consecutive failures (exponential backoff)
- **Invalid**: Permanently invalid keys requiring manual intervention

#### State Transitions

- **Automatic Transitions**: Based on validation results and request success/failure patterns
- **Manual Transitions**: Support for manual enable/disable/invalidate operations
- **Smart Logic**: Considers failure count, error types, and historical performance

#### Exponential Backoff

- **Progressive Delays**: 1min → 2min → 4min → ... → max 30min
- **Automatic Recovery**: Disabled keys automatically transition to Degraded after backoff period
- **Configurable Thresholds**: Customizable failure counts and backoff parameters

### 🎯 Policy Engine & Composition Strategy

#### Composition-Based Architecture

- **Reusable Policies**: Create policy templates that can be applied to multiple groups
- **Policy Types**: Retry policies, degradation policies, model filtering policies
- **Priority System**: Policies evaluated by priority order with first-match semantics

#### Advanced Retry Policies

```json
{
  "name": "Smart Retry Policy",
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

#### Model Filtering

- **Include/Exclude Patterns**: Support both whitelist and blacklist approaches
- **Regex & Wildcards**: Flexible pattern matching for model names
- **Per-Group Configuration**: Different filtering rules for different groups

### 📊 Incremental Validation System

#### Smart Filtering

- **Time Window Filtering**: Only validate keys added within specified timeframe
- **State Filtering**: Target specific key states (pending, invalid, etc.)
- **Deduplication**: Exclude recently validated keys to avoid redundant checks

#### High-Performance Processing

- **Batch Processing**: Configurable batch sizes for efficient database operations
- **Concurrent Validation**: Parallel key validation with configurable concurrency
- **Context Management**: Support for cancellation and timeout controls

#### RESTful API

```bash
# Validate specific group with custom parameters
POST /api/validation/groups/{groupId}
{
  "time_window_hours": 24,
  "include_states": ["pending", "invalid"],
  "exclude_recently_validated": true,
  "recent_validation_hours": 1,
  "concurrency": 5,
  "batch_size": 100
}

# Get validation history
GET /api/validation/groups/{groupId}/history?hours=24
```

## 🔌 API Usage Guide

<details>
<summary>Proxy Interface Invocation</summary>

GPT-Load routes requests to different AI services through group names. Usage is as follows:

### 1. Proxy Endpoint Format

```text
http://localhost:3001/proxy/{group_name}/{original_api_path}
```

- `{group_name}`: Group name created via the management API
- `{original_api_path}`: Maintain complete consistency with original AI service paths

### 2. Authentication Methods

Configure **Proxy Keys** via the management API, which supports system-level and group-level proxy keys.

- **Authentication Method**: Consistent with the native API, but replace the original key with the configured proxy key.
- **Key Scope**: **Global Proxy Keys** configured in system settings can be used in all groups. **Group Proxy Keys** configured in a group are only valid for the current group.
- **Format**: Multiple keys are separated by commas.

### 3. OpenAI Interface Example

Assuming a group named `openai` was created:

**Original invocation:**

```bash
curl -X POST https://api.openai.com/v1/chat/completions \
  -H "Authorization: Bearer sk-your-openai-key" \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4o-mini", "messages": [{"role": "user", "content": "Hello"}]}'
```

**Proxy invocation:**

```bash
curl -X POST http://localhost:3001/proxy/openai/v1/chat/completions \
  -H "Authorization: Bearer your-proxy-key" \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4o-mini", "messages": [{"role": "user", "content": "Hello"}]}'
```

**Changes required:**

- Replace `https://api.openai.com` with `http://localhost:3001/proxy/openai`
- Replace original API Key with the **Proxy Key**

### 4. Gemini Interface Example

Assuming a group named `gemini` was created:

**Original invocation:**

```bash
curl -X POST https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-exp:generateContent?key=your-gemini-key \
  -H "Content-Type: application/json" \
  -d '{"contents": [{"parts": [{"text": "Hello"}]}]}'
```

**Proxy invocation:**

```bash
curl -X POST http://localhost:3001/proxy/gemini/v1beta/models/gemini-2.0-flash-exp:generateContent?key=your-proxy-key \
  -H "Content-Type: application/json" \
  -d '{"contents": [{"parts": [{"text": "Hello"}]}]}'
```

**Changes required:**

- Replace `https://generativelanguage.googleapis.com` with `http://localhost:3001/proxy/gemini`
- Replace `key=your-gemini-key` in URL parameter with the **Proxy Key**

### 5. Anthropic Interface Example

Assuming a group named `anthropic` was created:

**Original invocation:**

```bash
curl -X POST https://api.anthropic.com/v1/messages \
  -H "x-api-key: sk-ant-api03-your-anthropic-key" \
  -H "anthropic-version: 2023-06-01" \
  -H "Content-Type: application/json" \
  -d '{"model": "claude-3-5-sonnet-20241022", "messages": [{"role": "user", "content": "Hello"}]}'
```

**Proxy invocation:**

```bash
curl -X POST http://localhost:3001/proxy/anthropic/v1/messages \
  -H "x-api-key: your-proxy-key" \
  -H "anthropic-version: 2023-06-01" \
  -H "Content-Type: application/json" \
  -d '{"model": "claude-3-5-sonnet-20241022", "messages": [{"role": "user", "content": "Hello"}]}'
```

**Changes required:**

- Replace `https://api.anthropic.com` with `http://localhost:3001/proxy/anthropic`
- Replace the original API Key in `x-api-key` header with the **Proxy Key**

### 6. Supported Interfaces

**OpenAI Format:**

- `/v1/chat/completions` - Chat conversations
- `/v1/completions` - Text completion
- `/v1/embeddings` - Text embeddings
- `/v1/models` - Model list
- And all other OpenAI-compatible interfaces

**Gemini Format:**

- `/v1beta/models/*/generateContent` - Content generation
- `/v1beta/models` - Model list
- And all other Gemini native interfaces

**Anthropic Format:**

- `/v1/messages` - Message conversations
- `/v1/models` - Model list (if available)
- And all other Anthropic native interfaces

### 7. Client SDK Configuration

**OpenAI Python SDK:**

```python
from openai import OpenAI

client = OpenAI(
    api_key="your-proxy-key",  # Use the proxy key
    base_url="http://localhost:3001/proxy/openai"  # Use proxy endpoint
)

response = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[{"role": "user", "content": "Hello"}]
)
```

**Google Gemini SDK (Python):**

```python
import google.generativeai as genai

# Configure API key and base URL
genai.configure(
    api_key="your-proxy-key",  # Use the proxy key
    client_options={"api_endpoint": "http://localhost:3001/proxy/gemini"}
)

model = genai.GenerativeModel('gemini-2.0-flash-exp')
response = model.generate_content("Hello")
```

**Anthropic SDK (Python):**

```python
from anthropic import Anthropic

client = Anthropic(
    api_key="your-proxy-key",  # Use the proxy key
    base_url="http://localhost:3001/proxy/anthropic"  # Use proxy endpoint
)

response = client.messages.create(
    model="claude-3-5-sonnet-20241022",
    messages=[{"role": "user", "content": "Hello"}]
)
```

> **Important Note**: As a transparent proxy service, GPT-Load completely preserves the native API formats and authentication methods of various AI services. You only need to replace the endpoint address and use the **Proxy Key** configured via the management API for seamless migration.

</details>

## 🔧 Management API Reference

<details>
<summary>Management API Endpoints</summary>

### Authentication

All management APIs require authentication using the `AUTH_KEY` configured in your environment.

```bash
curl -H "Authorization: Bearer your-auth-key" http://localhost:3001/api/...
```

### Key Validation APIs

#### Incremental Validation

```bash
# Validate specific group
POST /api/validation/groups/{groupId}
Content-Type: application/json
{
  "time_window_hours": 24,           // Only validate keys added in last 24 hours
  "include_states": ["pending", "invalid"],  // Target specific key states
  "exclude_recently_validated": true, // Skip recently validated keys
  "recent_validation_hours": 1,      // Define "recent" as 1 hour
  "concurrency": 5,                  // Concurrent validation workers
  "batch_size": 100                  // Batch processing size
}

# Response
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

#### Validation History

```bash
# Get validation history for a group
GET /api/validation/groups/{groupId}/history?hours=24

# Response
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

#### Validate All Groups

```bash
# Trigger validation for all groups
POST /api/validation/groups
Content-Type: application/json
{
  "time_window_hours": 1,
  "include_states": ["pending"],
  "concurrency": 3,
  "batch_size": 50
}

# Response
{
  "code": 0,
  "message": "Success",
  "data": {
    "results": [...],  // Individual group results
    "summary": {
      "total_groups": 5,
      "total_keys": 250,
      "total_validated": 200,
      "total_successful": 180
    }
  }
}
```

### Group Management APIs

#### List Groups

```bash
GET /api/groups

# Response
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

#### Group Statistics

```bash
GET /api/groups/{id}/stats

# Response
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

### Key Management APIs

#### Bulk Key Operations

```bash
# Add multiple keys
POST /api/keys/add-multiple
{
  "group_id": 1,
  "keys": ["sk-key1", "sk-key2", "sk-key3"]
}

# Delete multiple keys
POST /api/keys/delete-multiple
{
  "group_id": 1,
  "key_values": ["sk-key1", "sk-key2"]
}

# Export keys
GET /api/keys/export?group_id=1&format=json
```

### System APIs

#### Dashboard Statistics

```bash
GET /api/dashboard/stats

# Response
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

#### Health Check

```bash
GET /health

# Response
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "1.0.0"
}
```

</details>

## 🏗️ Architecture & Design

### 🔧 Core Architecture

GPT-Load is built with a modern, scalable architecture designed for enterprise environments:

#### Layered Architecture

- **Presentation Layer**: RESTful APIs for configuration and monitoring
- **Business Layer**: Policy engine, key management, validation services
- **Data Layer**: Database abstraction with support for multiple backends
- **Infrastructure Layer**: Caching, logging, monitoring, and networking

#### Dependency Injection

- **Container-based DI**: All services managed through dependency injection
- **Interface Abstraction**: Clean separation between interfaces and implementations
- **Testability**: Easy mocking and unit testing through interface contracts

#### Microservice-Ready Design

- **Service Separation**: Clear boundaries between different functional areas
- **Event-Driven**: Asynchronous processing for non-blocking operations
- **Stateless Components**: Horizontal scaling support with shared storage

### 🔄 Key Management Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Pending: Add Key
    Pending --> Active: Validation Success
    Pending --> Invalid: Validation Failed (3x)
    Active --> Degraded: Request Failed
    Degraded --> Active: Request Success
    Degraded --> Disabled: Consecutive Failures
    Disabled --> Degraded: Backoff Expired
    Active --> Invalid: Manual Action
    Degraded --> Invalid: Manual Action
    Invalid --> [*]: Remove Key
```

### 🎯 Policy Engine Flow

```mermaid
flowchart TD
    A[Request Failed] --> B[Get Group Policies]
    B --> C[Sort by Priority]
    C --> D[Evaluate Conditions]
    D --> E{Match Found?}
    E -->|Yes| F[Execute Action]
    E -->|No| G[Default Handling]
    F --> H[Update Key State]
    G --> H
    H --> I[Log Result]
```

### 📊 Performance Characteristics

#### Throughput

- **Concurrent Requests**: 1000+ concurrent connections supported
- **Key Validation**: 100+ keys/second validation throughput
- **Database Operations**: Optimized batch processing for bulk operations

#### Latency

- **Proxy Latency**: <5ms overhead for request forwarding
- **Key Selection**: <1ms for key pool operations
- **State Updates**: Asynchronous processing to avoid blocking

#### Scalability

- **Horizontal Scaling**: Leader-follower architecture
- **Database Scaling**: Support for read replicas and connection pooling
- **Caching**: Redis-based caching for hot data

## 🧪 Development & Testing

### 🔧 Development Setup

```bash
# Clone repository
git clone https://github.com/ShangjinTang/gpt-load.git
cd gpt-load

# Install dependencies
go mod tidy

# Set up environment
cp .env.example .env
# Edit .env with your configuration

# Run in development mode
make dev

# Run tests
make test

# Check test coverage
make test-coverage
```

### 🧪 Testing Strategy

#### Unit Tests

- **Service Layer**: Comprehensive unit tests for business logic
- **Mock Dependencies**: Interface-based mocking for isolated testing
- **Coverage Target**: >80% code coverage for critical components

#### Integration Tests

- **Database Tests**: In-memory SQLite for fast integration testing
- **API Tests**: Full HTTP API testing with test fixtures
- **Policy Tests**: Complex policy evaluation scenarios

#### Performance Tests

- **Load Testing**: Concurrent request handling validation
- **Memory Testing**: Memory leak detection and optimization
- **Benchmark Tests**: Performance regression detection

### 🔍 Code Quality

#### Static Analysis

- **Go Vet**: Built-in Go static analysis
- **Linting**: golangci-lint for comprehensive code quality checks
- **Security**: gosec for security vulnerability scanning

#### Code Organization

- **Clean Architecture**: Clear separation of concerns
- **Interface Design**: Dependency inversion principle
- **Error Handling**: Comprehensive error types and handling

## 🤝 Contributing

Thanks to all the developers who have contributed to GPT-Load!

[![Contributors](https://contrib.rocks/image?repo=ShangjinTang/gpt-load)](https://github.com/ShangjinTang/gpt-load/graphs/contributors)

### 📋 Contribution Guidelines

We welcome contributions! Please follow these guidelines:

#### 🐛 Bug Reports

- Use the issue template
- Include reproduction steps
- Provide system information
- Add relevant logs

#### ✨ Feature Requests

- Describe the use case
- Explain the expected behavior
- Consider backward compatibility
- Discuss implementation approach

#### 🔧 Pull Requests

- Fork the repository
- Create a feature branch
- Follow coding standards
- Add tests for new features
- Update documentation
- Ensure CI passes

#### 📚 Documentation

- Update README for new features
- Add code comments for complex logic
- Include API documentation
- Provide usage examples

## 🔒 Security

### 🛡️ Security Features

- **Data Encryption**: AES encryption for sensitive data at rest
- **Authentication**: Secure token-based authentication
- **Input Validation**: Comprehensive input sanitization
- **Rate Limiting**: Protection against abuse and DoS attacks

### 🚨 Reporting Security Issues

Please report security vulnerabilities privately to the maintainers. Do not use public issues for security reports.

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## ⭐ Star History

[![Stargazers over time](https://starchart.cc/ShangjinTang/gpt-load.svg?variant=adaptive)](https://starchart.cc/ShangjinTang/gpt-load)

---

## 🙏 Acknowledgments

Special thanks to:

- All contributors who have helped improve GPT-Load
- The Go community for excellent tools and libraries
- Users who provide feedback and feature requests

---

_Built with ❤️ by the GPT-Load team_

## 📌 Fork Information

This repository is forked from [https://github.com/tbphp/gpt-load](https://github.com/tbphp/gpt-load) at version v1.2.1.
