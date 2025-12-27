package common

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// Task types
const (
	TypeWelcomeMessage = "welcome:message"
	TypeEmailTask      = "email:send"
	TypeServerInfo     = "server:info"
)

// WelcomePayload represents the payload for welcome message tasks
type WelcomePayload struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

// EmailPayload represents the payload for email tasks
type EmailPayload struct {
	UserID  int    `json:"user_id"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// ServerInfoPayload represents the payload for server info tasks
type ServerInfoPayload struct {
	Timestamp int64  `json:"timestamp"`
	Source    string `json:"source"`
}

// HandleWelcomeTask processes welcome message tasks
func HandleWelcomeTask(ctx context.Context, p *WelcomePayload) error {
	fmt.Printf("👋 [Welcome] Hello %s (ID: %d)! %s\n", p.Username, p.UserID, p.Message)
	// Simulate processing time
	time.Sleep(200 * time.Millisecond)
	return nil
}

// HandleEmailTask processes email sending tasks
func HandleEmailTask(ctx context.Context, p *EmailPayload) error {
	fmt.Printf("📧 [Email] Sending email to %s (UserID: %d)\n", p.Email, p.UserID)
	fmt.Printf("   Subject: %s\n", p.Subject)
	fmt.Printf("   Message: %s\n", p.Message)
	fmt.Println("   ✅ Email sent successfully!")

	// Simulate processing time
	time.Sleep(300 * time.Millisecond)
	return nil
}

// HandleServerInfoTask processes server info tasks and prints current server information
func HandleServerInfoTask(ctx context.Context, p *ServerInfoPayload) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("🖥️  [Server Info] %s - 系统状态报告\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("   📅 时间戳: %d\n", p.Timestamp)
	fmt.Printf("   🔢 CPU核心数: %d\n", runtime.NumCPU())
	fmt.Printf("   🧵 当前Goroutines: %d\n", runtime.NumGoroutine())
	fmt.Printf("   💾 分配内存: %.2f MB\n", float64(m.Alloc)/1024/1024)
	fmt.Printf("   🔄 系统内存: %.2f MB\n", float64(m.Sys)/1024/1024)
	fmt.Printf("   🗑️  GC次数: %d\n", m.NumGC)

	// 避免除零错误
	if m.NumGC > 0 {
		fmt.Printf("   ⏱️  平均GC暂停时间: %v\n", time.Duration(m.PauseTotalNs/uint64(m.NumGC)))
	} else {
		fmt.Printf("   ⏱️  平均GC暂停时间: N/A\n")
	}

	fmt.Printf("   📊 堆使用: %.2f MB\n", float64(m.HeapAlloc)/1024/1024)
	fmt.Printf("   📈 堆系统: %.2f MB\n", float64(m.HeapSys)/1024/1024)
	fmt.Printf("   🏗️  堆对象数: %d\n", m.HeapObjects)
	fmt.Printf("   📋 来源: %s\n", p.Source)
	fmt.Println("   ✅ 服务器信息收集完成")

	return nil
}
