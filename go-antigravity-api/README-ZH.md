# Go Antigravity API

一个独立的 Go 实现的 Antigravity API，提供 Gemini 兼容接口来访问 Google 内部的 Antigravity 服务，支持最新的模型，包括 Gemini 3 Pro、Claude 4.5 Opus 等。

## 特性

- **🔐 OAuth2 认证** - 自动 Google OAuth2 流程，支持令牌管理和刷新
- **🌐 多环境支持** - Daily 和 Autopush 环境自动故障转移
- **🔄 智能重试逻辑** - 指数退避与自动环境切换
- **📊 模型管理** - 动态模型发现和别名映射
- **💬 流式支持** - 服务器发送事件 (SSE) 实时响应
- **📈 配额监控** - 所有模型的实时配额跟踪
- **🔧 灵活配置** - JSON 配置文件或环境变量
- **⚡ 高性能** - 使用 Go 构建，性能卓越

## 支持的模型

该服务提供对以下前沿模型的访问：

- **Gemini 3 Pro** (`gemini-3-pro-preview`) - 最新 Gemini 架构
- **Gemini 3 Flash** (`gemini-3-flash-preview`) - 快速高效
- **Gemini 2.5 Flash** (`gemini-2.5-flash`) - 平衡性能
- **Gemini 2.5 Computer Use** (`gemini-2.5-computer-use-preview-10-2025`) - 计算机交互
- **Claude Sonnet 4.5** (`gemini-claude-sonnet-4-5`) - 通过 Antigravity 访问 Claude
- **Claude Opus 4.5 Thinking** (`gemini-claude-opus-4-5-thinking`) - 高级推理

## 快速开始

### 先决条件

- Go 1.21 或更高版本
- 有 Antigravity 访问权限的 Google 账户
- OAuth2 凭据（自动处理）

### 安装

```bash
# 进入项目目录
cd go-antigravity-api

# 安装依赖
go mod download

# 构建服务器
go build -o antigravity-api cmd/server/main.go
```

### 运行服务器

#### 使用配置文件

```bash
# 复制并编辑示例配置
cp example-config.json config.json
# 使用你的设置编辑 config.json

# 启动服务器
./antigravity-api -config config.json
```

#### 使用环境变量

```bash
export SERVER_PORT=3000
export SERVER_API_KEY=your-secret-key
export ANTIGRAVITY_CREDS_FILE=/path/to/oauth_creds.json
export ANTIGRAVITY_PROJECT_ID=your-project-id

./antigravity-api
```

#### 使用命令行参数

```bash
./antigravity-api -config config.json -port 8080 -project-id my-project
```

### 首次 OAuth 设置

当你第一次运行服务器时，它会引导你完成 OAuth 流程：

1. 服务器将打印授权 URL
2. 在浏览器中访问该 URL
3. 使用你的 Google 账户登录
4. 授权应用程序
5. 复制授权码
6. 将其粘贴到终端
7. 凭据将保存到 `~/.antigravity/oauth_creds.json`

后续运行将自动使用保存的凭据。

## 配置

### 配置文件格式

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 3000,
    "apiKey": "your-api-key-here"
  },
  "antigravity": {
    "oauthCredsFilePath": "/path/to/oauth_creds.json",
    "projectId": "",
    "baseUrlDaily": "https://daily-cloudcode-pa.sandbox.googleapis.com",
    "baseUrlAutopush": "https://autopush-cloudcode-pa.sandbox.googleapis.com",
    "userAgent": "antigravity/1.11.5 windows/amd64",
    "requestMaxRetries": 3,
    "requestBaseDelay": 1000,
    "cronNearMinutes": 50
  }
}
```

### 环境变量

| 变量 | 描述 | 默认值 |
|------|------|--------|
| `SERVER_HOST` | 服务器主机 | `0.0.0.0` |
| `SERVER_PORT` | 服务器端口 | `3000` |
| `SERVER_API_KEY` | 用于身份验证的 API 密钥 | (无) |
| `ANTIGRAVITY_CREDS_FILE` | OAuth 凭据文件路径 | `~/.antigravity/oauth_creds.json` |
| `ANTIGRAVITY_PROJECT_ID` | Google Cloud 项目 ID | (自动发现) |
| `ANTIGRAVITY_BASE_URL_DAILY` | Daily 环境 URL | (默认) |
| `ANTIGRAVITY_BASE_URL_AUTOPUSH` | Autopush 环境 URL | (默认) |
| `ANTIGRAVITY_USER_AGENT` | 用户代理字符串 | `antigravity/1.11.5 windows/amd64` |
| `ANTIGRAVITY_MAX_RETRIES` | 最大重试次数 | `3` |
| `ANTIGRAVITY_BASE_DELAY` | 重试基础延迟 (毫秒) | `1000` |

## API 使用

该服务提供 Gemini 兼容的 API 接口：

### 列出模型

```bash
curl http://localhost:3000/v1beta/models \
  -H "Authorization: Bearer your-api-key"
```

### 生成内容（非流式）

```bash
curl http://localhost:3000/v1beta/models/gemini-3-pro-preview:generateContent \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [
      {
        "role": "user",
        "parts": [{"text": "用简单的术语解释量子计算"}]
      }
    ]
  }'
```

### 流式生成内容

```bash
curl http://localhost:3000/v1beta/models/gemini-3-pro-preview:streamGenerateContent \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [
      {
        "role": "user",
        "parts": [{"text": "写一个关于 AI 的短篇故事"}]
      }
    ]
  }'
```

### 获取使用限制

```bash
curl http://localhost:3000/usage \
  -H "Authorization: Bearer your-api-key"
```

### 健康检查

```bash
curl http://localhost:3000/health
```

## 架构

### 项目结构

```
go-antigravity-api/
├── cmd/
│   └── server/
│       └── main.go              # 主入口点
├── internal/
│   ├── antigravity/
│   │   ├── auth.go              # OAuth2 认证
│   │   ├── client.go            # 带重试逻辑的 API 客户端
│   │   ├── constants.go         # 常量和模型映射
│   │   ├── converter.go         # 格式转换
│   │   ├── discovery.go         # 项目发现
│   │   ├── models.go            # 模型管理
│   │   ├── quota.go             # 配额跟踪
│   │   ├── service.go           # 主服务编排
│   │   ├── stream.go            # SSE 流式处理
│   │   └── utils.go             # 实用函数
│   ├── api/
│   │   └── handlers.go          # HTTP 处理器
│   └── config/
│       └── config.go            # 配置管理
├── pkg/
│   └── models/
│       └── types.go             # 共享类型定义
├── go.mod                       # Go 模块定义
├── example-config.json          # 示例配置
├── .gitignore                   # Git 忽略规则
└── README.md                    # 英文文档
```

### 核心组件

1. **认证管理器** - 处理 OAuth2 流程、令牌刷新和凭据存储
2. **API 客户端** - 使用智能重试和故障转移发起 HTTP 请求
3. **发现管理器** - 发现或创建 Google Cloud 项目 ID
4. **模型管理器** - 管理可用模型及其功能
5. **配额管理器** - 跟踪和报告使用配额
6. **转换器** - 在 Gemini 和 Antigravity 格式之间转换
7. **流处理器** - 解析 SSE 流以实现实时响应

### 多环境故障转移

服务自动按顺序尝试多个环境：

1. **Daily 环境** - 主端点
2. **Autopush 环境** - 备用端点

在以下情况下切换：
- 429（速率限制）错误
- 网络故障
- 5xx 服务器错误

### 错误处理

- **400/401** - 刷新 OAuth 令牌并重试
- **429** - 切换环境或指数退避
- **500-599** - 指数退避并重试
- **网络错误** - 立即切换环境

## 高级功能

### 令牌管理

- 过期前 50 分钟自动刷新令牌
- 持久化凭据存储
- 优雅处理过期令牌

### 项目 ID 管理

服务按以下优先级处理项目 ID：

1. 配置文件
2. 命令行参数
3. 通过 `loadCodeAssist` API 自动发现
4. 通过 `onboardUser` API 自动创建
5. 作为后备随机生成

### 模型别名

友好的模型名称自动映射到内部名称：

- `gemini-3-pro-preview` → `gemini-3-pro-high`
- `gemini-claude-sonnet-4-5` → `claude-sonnet-4-5`

## 与 Node.js 版本的对比

| 特性 | Go 版本 | Node.js 版本 |
|------|---------|--------------|
| 性能 | ⚡ 更快 | 标准 |
| 内存使用 | 💾 更低 | 更高 |
| 并发 | 🔀 原生 goroutines | 事件循环 |
| 编译 | ✅ 单一二进制文件 | 需要运行时 |
| 依赖 | 最小化 | 多个包 |
| 类型安全 | ✅ 强类型 | 动态类型 |

## 故障排除

### OAuth 问题

**问题**："Failed to get access token"
- **解决方案**：删除 `~/.antigravity/oauth_creds.json` 并重新认证

**问题**："Token refresh failed"
- **解决方案**：确保你的 Google 账户有 Antigravity 访问权限

### API 错误

**问题**："All Antigravity base URLs failed"
- **解决方案**：检查网络连接和防火墙设置

**问题**："Rate limit exceeded"
- **解决方案**：服务将自动使用退避重试

### 项目 ID 问题

**问题**："Failed to discover Project ID"
- **解决方案**：在配置中指定 `projectId` 或使用 `-project-id` 标志

## 开发

### 构建

```bash
# 为当前平台构建
go build -o antigravity-api cmd/server/main.go

# 为 Linux 构建
GOOS=linux GOARCH=amd64 go build -o antigravity-api-linux cmd/server/main.go

# 为 Windows 构建
GOOS=windows GOARCH=amd64 go build -o antigravity-api.exe cmd/server/main.go

# 为 macOS 构建
GOOS=darwin GOARCH=amd64 go build -o antigravity-api-macos cmd/server/main.go
```

### 测试

```bash
# 运行测试
go test ./...

# 运行覆盖率测试
go test -cover ./...

# 运行详细输出测试
go test -v ./...
```

### 代码格式化

```bash
# 格式化代码
go fmt ./...

# 运行 linter
go vet ./...
```

## 安全考虑

- 安全存储 OAuth 凭据
- 使用强 API 密钥
- 生产环境使用 HTTPS
- 适当限制网络访问
- 定期轮换凭据

## 许可证

此项目遵循父项目 AIClient-2-API 的相同许可证。

## 贡献

欢迎贡献！请确保：

- 代码遵循 Go 最佳实践
- 新功能包含测试
- 文档已更新
- 提交说明清晰

## 支持

对于问题和疑问：
- 在 GitHub 仓库中打开 issue
- 检查现有 issue 寻找解决方案
- 查看故障排除部分

## 致谢

此实现基于 AIClient-2-API 中的 Node.js Antigravity 服务，并针对 Go 的性能和并发模型进行了增强。
