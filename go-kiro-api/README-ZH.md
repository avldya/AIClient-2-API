# Go Kiro API

Kiro API 客户端的高性能 Go 实现，提供 OpenAI 和 Claude 兼容接口以访问 CodeWhisperer/Amazon Q 模型。

## 功能特性

- ✅ **多种认证方式**：支持 Base64 凭据、文件路径和目录扫描
- ✅ **自动刷新 Token**：OAuth token 管理，在过期前自动刷新
- ✅ **OpenAI 兼容**：完全支持 `/v1/chat/completions` 端点
- ✅ **Claude 兼容**：完全支持 `/v1/messages` 端点
- ✅ **流式支持**：OpenAI 和 Claude 格式的实时流式响应
- ✅ **工具调用**：完整支持函数/工具调用
- ✅ **重试逻辑**：失败请求自动重试，采用指数退避策略
- ✅ **错误处理**：全面的错误处理，403 错误时自动刷新 token
- ✅ **用量跟踪**：查询用量限制和剩余配额

## 支持的模型

- `claude-opus-4-5`
- `claude-opus-4-5-20251101`
- `claude-haiku-4-5`
- `claude-sonnet-4-5`
- `claude-sonnet-4-5-20250929`
- `claude-sonnet-4-20250514`
- `claude-3-7-sonnet-20250219`

## 快速开始

### 安装

```bash
# 克隆仓库
git clone https://github.com/avldya/AIClient-2-API.git
cd AIClient-2-API/go-kiro-api

# 编译二进制文件
go build -o kiro-api cmd/server/main.go

# 运行服务器
./kiro-api -config config.json
```

### 配置

创建 `config.json` 文件：

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 3000,
    "apiKey": "your-api-key"
  },
  "kiro": {
    "credPath": "/path/to/.aws/sso/cache",
    "credsBase64": "",
    "credsFilePath": "",
    "region": "us-east-1",
    "requestMaxRetries": 3,
    "requestBaseDelay": 1000,
    "cronNearMinutes": 10
  }
}
```

### 环境变量

也可以使用环境变量进行配置：

```bash
export SERVER_HOST=0.0.0.0
export SERVER_PORT=3000
export SERVER_API_KEY=your-api-key
export KIRO_CREDS_BASE64=your-base64-encoded-credentials
export KIRO_REGION=us-east-1
```

### 运行服务器

```bash
# 使用配置文件
./kiro-api -config config.json

# 使用命令行参数
./kiro-api -host 0.0.0.0 -port 3000 -api-key your-api-key

# 使用环境变量
export KIRO_CREDS_BASE64="..."
export SERVER_PORT=3000
./kiro-api
```

## API 使用

### OpenAI 格式（非流式）

```bash
curl http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-5",
    "messages": [
      {"role": "user", "content": "你好，最近怎么样？"}
    ]
  }'
```

### OpenAI 格式（流式）

```bash
curl http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-5",
    "messages": [
      {"role": "user", "content": "给我讲个故事"}
    ],
    "stream": true
  }'
```

### Claude 格式（非流式）

```bash
curl http://localhost:3000/v1/messages \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-5",
    "messages": [
      {"role": "user", "content": "你好！"}
    ],
    "max_tokens": 1024
  }'
```

### Claude 格式（流式）

```bash
curl http://localhost:3000/v1/messages \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-5",
    "messages": [
      {"role": "user", "content": "写一首诗"}
    ],
    "stream": true
  }'
```

### 列出模型

```bash
curl http://localhost:3000/v1/models \
  -H "Authorization: Bearer your-api-key"
```

### 计算 Token 数量

```bash
curl http://localhost:3000/v1/messages/count_tokens \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {"role": "user", "content": "你好，世界！"}
    ]
  }'
```

### 获取用量限制

```bash
curl http://localhost:3000/getUsageLimits \
  -H "Authorization: Bearer your-api-key"
```

## 配置选项

### 服务器配置

- `server.host`: 服务器监听地址（默认：`0.0.0.0`）
- `server.port`: 服务器监听端口（默认：`3000`）
- `server.apiKey`: 认证 API 密钥（可选）

### Kiro 配置

- `kiro.credPath`: 凭据目录路径（默认：`~/.aws/sso/cache`）
- `kiro.credsBase64`: Base64 编码的凭据（最高优先级）
- `kiro.credsFilePath`: 特定凭据文件路径
- `kiro.region`: AWS 区域（默认：`us-east-1`）
- `kiro.proxy`: HTTP 代理 URL（可选）
- `kiro.requestMaxRetries`: 最大重试次数（默认：`3`）
- `kiro.requestBaseDelay`: 重试基础延迟（毫秒，默认：`1000`）
- `kiro.cronNearMinutes`: token 过期前多少分钟触发刷新（默认：`10`）

## 认证方法

客户端支持三种加载凭据的方法（按优先级排序）：

1. **Base64 凭据**：通过 `credsBase64` 提供 Base64 编码的 JSON 字符串
2. **文件路径**：通过 `credsFilePath` 指定凭据文件的直接路径
3. **目录扫描**：通过 `credPath` 自动扫描目录（类似 AWS CLI 的 `.aws/sso/cache`）

凭据从所有可用来源合并，较早的来源优先。

## 工具调用支持

API 支持 OpenAI 和 Claude 格式的函数/工具调用：

```bash
curl http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-5",
    "messages": [
      {"role": "user", "content": "旧金山的天气怎么样？"}
    ],
    "tools": [
      {
        "name": "get_weather",
        "description": "获取某地的当前天气",
        "input_schema": {
          "type": "object",
          "properties": {
            "location": {
              "type": "string",
              "description": "城市名称"
            }
          },
          "required": ["location"]
        }
      }
    ]
  }'
```

## 部署

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o kiro-api cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/kiro-api .
EXPOSE 3000
CMD ["./kiro-api"]
```

构建和运行：

```bash
docker build -t kiro-api .
docker run -p 3000:3000 \
  -e KIRO_CREDS_BASE64=your-base64-creds \
  kiro-api
```

### 二进制部署

为不同平台构建：

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o kiro-api-linux cmd/server/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o kiro-api-windows.exe cmd/server/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o kiro-api-macos cmd/server/main.go
```

## 性能

Go 实现相比 Node.js 版本提供了显著的性能提升：

- **更低的内存使用**：~10-20MB vs Node.js 的 50-100MB
- **更快的启动时间**：~100ms vs Node.js 的 500ms
- **更好的并发性**：原生 goroutine vs 事件循环
- **更小的二进制文件**：~10MB vs 50MB+ (含 node_modules)

## 错误处理

客户端实现了全面的错误处理：

- **403 错误**：自动刷新 token 并重试
- **429 错误**：指数退避重试
- **5xx 错误**：指数退避重试
- **Token 过期**：在过期前 10 分钟主动刷新

## 架构

```
go-kiro-api/
├── cmd/
│   └── server/
│       └── main.go              # 应用程序入口
├── internal/
│   ├── kiro/
│   │   ├── client.go           # 带重试逻辑的 HTTP 客户端
│   │   ├── auth.go             # 认证和 token 管理
│   │   ├── converter.go        # 格式转换逻辑
│   │   ├── stream.go           # 流式响应解析
│   │   ├── constants.go        # 常量和模型映射
│   │   └── utils.go            # 工具函数
│   ├── api/
│   │   ├── handlers.go         # 通用 HTTP 处理器
│   │   ├── openai.go           # OpenAI 兼容端点
│   │   └── claude.go           # Claude 兼容端点
│   └── config/
│       └── config.go            # 配置管理
└── pkg/
    └── models/
        └── types.go             # 类型定义
```

## 与 Node.js 版本对比

| 特性 | Go 版本 | Node.js 版本 |
|------|---------|--------------|
| 内存使用 | 10-20MB | 50-100MB |
| 启动时间 | ~100ms | ~500ms |
| 二进制大小 | ~10MB | 50MB+ (含 node_modules) |
| 并发性 | 原生 goroutine | 事件循环 |
| 类型安全 | 强静态类型 | 动态类型 |
| 依赖项 | 最小化（仅 uuid） | 多个 npm 包 |

## 许可证

本项目采用 GPL v3 许可证 - 详见 LICENSE 文件。

## 贡献

欢迎贡献！请随时提交 Pull Request。

## 支持

如有问题，请在 GitHub 上提 issue。
