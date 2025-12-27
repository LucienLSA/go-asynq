# Go Asynq Demo - 异步任务队列学习项目

一个完整的 Go 异步任务队列演示项目，展示如何使用 [Asynq](https://github.com/hibiken/asynq) 库构建生产级的任务处理系统。

## 🎯 项目目标

本项目旨在通过一个简洁的示例，展示 Go 语言中异步任务队列的核心概念和最佳实践，帮助开发者快速上手 Asynq 库。

## 📚 核心知识点

### 1. 异步任务队列的基本概念

- **生产者 (Producer)**: 创建和提交任务的组件
- **消费者 (Consumer)**: 处理任务的工作进程
- **任务队列**: 存储待处理任务的数据结构
- **任务类型**: 不同业务逻辑的任务分类
- **任务载荷**: 任务执行所需的数据

### 2. Asynq 核心组件

#### Client (客户端)
```go
client := asynq.NewClient(redisConnOpt)
defer client.Close()
```
- 负责将任务提交到队列
- 支持多种任务类型：立即执行、延迟执行、定时执行

#### Server (服务器)
```go
srv := asynq.NewServer(redisConnOpt, asynq.Config{
    Concurrency: 5,  // 并发处理任务数
    Queues: map[string]int{
        "critical": 6, // 队列优先级权重
        "default":  3,
        "low":      1,
    },
})
```
- 管理工作进程
- 处理任务分发和执行
- 支持优雅关闭

#### Task & Handler (任务和处理器)
```go
// 任务定义
type WelcomePayload struct {
    UserID   int    `json:"user_id"`
    Username string `json:"username"`
    Message  string `json:"message"`
}

// 任务处理器
func HandleWelcomeTask(ctx context.Context, t *asynq.Task) error {
    var p WelcomePayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return err
    }
    return processWelcomeTask(ctx, &p)
}
```

### 3. 任务类型和执行模式

#### 立即任务 (Immediate Tasks)
```go
info, err := client.Enqueue(asynq.NewTask(TypeWelcomeMessage, payload))
```

#### 延迟任务 (Delayed Tasks)
```go
info, err := client.Enqueue(
    asynq.NewTask(TypeEmailTask, payload),
    asynq.ProcessIn(5*time.Second), // 5秒后执行
)
```

#### 定时任务 (Scheduled Tasks)
```go
info, err := client.Enqueue(
    asynq.NewTask(TypeEmailTask, payload),
    asynq.ProcessAt(specificTime), // 在指定时间执行
)
```

#### 周期性任务 (Periodic Tasks)
```go
// 使用调度器创建周期性任务
scheduler := asynq.NewScheduler(redisConnOpt, nil)
scheduler.Register("@every 30s", asynq.NewTask(TypeServerInfo, payload))
scheduler.Start()
```

### 4. 队列优先级系统

Asynq 支持多队列优先级调度：
- **Critical**: 权重 6，最高优先级
- **Default**: 权重 3，默认优先级
- **Low**: 权重 1，最低优先级

队列权重影响任务的处理顺序，权重高的队列优先处理。

### 5. 服务器监控任务示例

项目包含一个每30秒执行的服务器信息收集任务：

```go
// 服务器信息载荷
type ServerInfoPayload struct {
    Timestamp int64  `json:"timestamp"`
    Source    string `json:"source"`
}

// 处理器收集系统信息
func HandleServerInfoTask(ctx context.Context, p *ServerInfoPayload) error {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    fmt.Printf("🖥️  [Server Info] %s - 系统状态报告\n", time.Now().Format("2006-01-02 15:04:05"))
    fmt.Printf("   🔢 CPU核心数: %d\n", runtime.NumCPU())
    fmt.Printf("   💾 分配内存: %.2f MB\n", float64(m.Alloc)/1024/1024)
    fmt.Printf("   🧵 当前Goroutines: %d\n", runtime.NumGoroutine())
    return nil
}
```

### 5. JSON 序列化与反序列化

任务载荷使用 JSON 格式传输：
```go
// 序列化
payload, err := json.Marshal(taskData)

// 反序列化
var p PayloadType
err := json.Unmarshal(t.Payload(), &p)
```

### 6. 并发处理和资源管理

- **并发控制**: 通过 `Concurrency` 参数控制同时处理的任务数
- **上下文管理**: 使用 `context.Context` 进行超时和取消控制
- **优雅关闭**: 正确处理程序退出时的资源清理

### 7. 错误处理和重试机制

Asynq 内置重试机制：
- 任务失败后自动重试
- 可配置最大重试次数
- 支持自定义错误处理逻辑

### 8. 信号处理和进程管理

```go
// 优雅关闭处理
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
<-sigChan

srv.Shutdown() // 优雅关闭服务器
wg.Wait()      // 等待所有 goroutine 完成
```

## 🏗️ 项目架构分析

### 目录结构（已简化）
```
go-asynq/
├── main.go         # 主程序入口，包含生产者和消费者逻辑
├── common/
│   └── task.go     # 任务类型定义和业务处理逻辑
├── run.sh          # 一键运行脚本（启动 Redis + 编译并运行）
├── go.mod          # Go 模块依赖管理
├── go.sum          # 依赖校验文件
└── README.md       # 项目文档和知识点总结
```

### 代码设计模式

#### 1. 工厂模式 (Factory Pattern)
```go
client := asynq.NewClient(redisConnOpt)
srv := asynq.NewServer(redisConnOpt, config)
```

#### 2. 策略模式 (Strategy Pattern)
```go
mux := asynq.NewServeMux()
mux.HandleFunc(TypeWelcomeMessage, HandleWelcomeTask)
mux.HandleFunc(TypeEmailTask, HandleEmailTask)
```

#### 3. 模板方法模式 (Template Method Pattern)
```go
func HandleWelcomeTask(ctx context.Context, t *asynq.Task) error {
    var p WelcomePayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return fmt.Errorf("failed to unmarshal: %v", err)
    }
    return common.HandleWelcomeTask(ctx, &p)
}
```

## 🚀 快速开始

### 环境要求
- Go 1.21+
- Redis 6.0+

### 一键运行（推荐）
```bash
./run.sh
```
自动启动 Redis、编译并运行演示程序。

### 手动运行
```bash
# 启动 Redis
docker run -d -p 6380:6379 redis:7-alpine

# 运行演示
go run main.go
```

## 🎯 功能演示

程序会演示三种任务类型：

1. **欢迎消息任务** - 模拟用户注册欢迎
2. **邮件发送任务** - 模拟发送邮件通知
3. **服务器信息监控** - 每30秒自动打印系统状态

运行示例输出：
```
🚀 Starting Asynq Demo...
📍 Redis: localhost:6380
🐰 Consumer started, waiting for tasks...
⏰ Server info scheduler registered - runs every 30 seconds
📤 Creating welcome message tasks...
✅ Enqueued welcome task for Alice (ID: ...)
👋 [Welcome] Hello Alice (ID: 1)! Welcome to our amazing platform!
🖥️  [Server Info] 系统状态报告
   🔢 CPU核心数: 24
   💾 分配内存: 1.03 MB
   🧵 当前Goroutines: 16
   ✅ 服务器信息收集完成
```

## 🔧 配置说明

### Redis 连接配置
```go
redisConnOpt := asynq.RedisClientOpt{
    Addr: "localhost:6380",
    // Password: "your-password",     // 可选：Redis 密码
    // DB: 1,                        // 可选：Redis 数据库
}
```

### 服务器配置
```go
config := asynq.Config{
    Concurrency: 5,  // 并发处理任务数
    Queues: map[string]int{
        "critical": 6, // 队列优先级权重
        "default":  3,
        "low":      1,
    },
    // Logger: customLogger,          // 可选：自定义日志器
    // ShutdownTimeout: 30*time.Second, // 可选：关闭超时时间
}
```

## 🛠️ 扩展开发指南

### 添加新任务类型

1. **定义任务载荷结构体**
```go
type NotificationPayload struct {
    UserID  int    `json:"user_id"`
    Title   string `json:"title"`
    Content string `json:"content"`
    Type    string `json:"type"` // email, sms, push
}
```

2. **定义任务类型常量**
```go
const TypeNotification = "notification:send"
```

3. **实现任务处理器**
```go
func HandleNotificationTask(ctx context.Context, p *NotificationPayload) error {
    fmt.Printf("🔔 Sending %s notification to user %d\n", p.Type, p.UserID)
    // 实现具体的通知发送逻辑
    return nil
}
```

#### 添加周期性任务

1. **创建调度器**
```go
scheduler := asynq.NewScheduler(redisConnOpt, nil)
```

2. **注册周期性任务**
```go
// 每30秒执行一次
scheduler.Register("@every 30s", asynq.NewTask(TypeServerInfo, payload))

// Cron 表达式
scheduler.Register("0 */5 * * *", asynq.NewTask(TypeCleanup, payload)) // 每5分钟
```

3. **启动调度器**
```go
scheduler.Start()
defer scheduler.Shutdown()
```

4. **注册任务处理器**
```go
mux.HandleFunc(TypeNotification, func(ctx context.Context, t *asynq.Task) error {
    var p NotificationPayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return err
    }
    return HandleNotificationTask(ctx, &p)
})
```

5. **创建任务**
```go
payload := NotificationPayload{
    UserID:  123,
    Title:   "Welcome!",
    Content: "Welcome to our platform",
    Type:    "email",
}

data, _ := json.Marshal(payload)
client.Enqueue(asynq.NewTask(TypeNotification, data))
```

## 📊 监控和调试

### 启动网页 UI（可选）

推荐使用 `run.sh` 启动演示并观察控制台输出。若需可视化监控，可单独安装并运行 `asynqmon`：

```bash
# 启动演示（包含启动本地 Redis）
./run.sh

# 可选：安装并运行 asynqmon（手动方式）
go install github.com/hibiken/asynqmon/cmd/asynqmon@latest
asynqmon -redis-addr=localhost:6380
```

访问监控界面： http://localhost:8080  
监控界面可以查看：队列状态、活跃任务、延迟/重试/失败任务和任务详情。

### 监控界面功能

启动后访问 `http://localhost:8080` 查看：

#### 📈 仪表板 (Dashboard)
- **队列状态总览**：各队列的任务数量和状态
- **系统指标**：内存使用、CPU占用、goroutine数量
- **实时统计**：处理速度、成功率、失败率

#### 📋 队列管理 (Queues)
- **活跃任务 (Active)**：正在处理的任务
- **等待任务 (Pending)**：队列中的待处理任务
- **延迟任务 (Scheduled)**：定时执行的任务
- **重试任务 (Retry)**：失败后等待重试的任务
- **死信队列 (Dead)**：多次失败的任务

#### 📝 任务详情 (Tasks)
- **任务执行历史**：所有已完成任务的记录
- **任务详细信息**：载荷内容、执行时间、错误信息
- **任务重试记录**：失败原因和重试历史

#### ⚠️ 失败任务 (Failures)
- **失败统计**：按任务类型和错误类型的统计
- **错误详情**：具体的错误信息和堆栈跟踪
- **手动重试**：支持手动重新执行失败任务

#### 🔍 实时监控 (Live)
- **实时任务流**：新任务的实时显示
- **性能指标**：处理延迟、吞吐量等

### Redis 命令行监控
```bash
# 连接到 Redis
redis-cli -p 6380

# 查看队列信息
KEYS "asynq:*"

# 查看队列长度
LLEN "asynq:{critical}:active"
LLEN "asynq:{default}:pending"
```

## 🔧 故障排除

### 常见问题

#### 1. Redis 连接失败
**错误**: `dial tcp [::1]:6380: connect: connection refused`
**解决**:
```bash
# 检查 Redis 是否运行
docker ps | grep redis

# 重新启动 Redis
docker run -d -p 6380:6379 redis:7-alpine
```

#### 2. 任务处理器未注册
**错误**: `task not registered`
**解决**: 确保在 `main.go` 中正确注册了所有任务处理器

#### 3. JSON 序列化错误
**错误**: `failed to unmarshal payload`
**解决**: 检查结构体标签和 JSON 字段名是否匹配

### 调试技巧

1. **启用详细日志**
```go
import "log"

srv.Run(mux) // Asynq 会自动输出详细日志
```

2. **添加自定义日志**
```go
func HandleTask(ctx context.Context, t *asynq.Task) error {
    log.Printf("Processing task: %s", t.Type())
    // 处理逻辑
}
```

## 🎯 最佳实践

### 1. 错误处理
```go
func HandleTask(ctx context.Context, t *asynq.Task) error {
    if err := validatePayload(t.Payload()); err != nil {
        return fmt.Errorf("invalid payload: %w", err)
    }

    if err := processTask(t); err != nil {
        return fmt.Errorf("failed to process task: %w", err)
    }

    return nil
}
```

### 2. 资源管理
```go
func main() {
    client := asynq.NewClient(redisConnOpt)
    defer client.Close()

    srv := asynq.NewServer(redisConnOpt, config)

    // 优雅关闭处理
    go func() {
        <-sigChan
        srv.Shutdown()
    }()

    srv.Run(mux)
}
```

### 3. 任务设计
- 任务载荷保持精简
- 使用有意义的类型名称
- 包含必要的上下文信息
- 支持幂等操作

### 4. 性能优化
- 合理设置并发数
- 使用连接池
- 监控队列积压
- 定期清理过期任务

## 📈 生产环境部署

### Docker 部署
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

### Kubernetes 部署
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: asynq-worker
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: worker
        image: your-registry/asynq-worker:latest
        env:
        - name: REDIS_ADDR
          value: "redis-service:6379"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

## 🤝 贡献指南

欢迎贡献！请遵循以下步骤：

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📚 相关链接

- [Asynq 官方文档](https://github.com/hibiken/asynq)
- [Redis 官方文档](https://redis.io/documentation)
- [Go 官方文档](https://golang.org/doc/)

---

**学习建议**: 通过运行 `./run.sh` 脚本开始你的 Asynq 学习之旅，逐步深入了解每个知识点，然后尝试添加自己的任务类型！