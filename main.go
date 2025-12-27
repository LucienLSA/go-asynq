package main

import (
	"asynqdemo/common"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
)

// HandleWelcomeTask wraps the common handler for Asynq
func HandleWelcomeTask(ctx context.Context, t *asynq.Task) error {
	var p common.WelcomePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal welcome payload: %v", err)
	}
	return common.HandleWelcomeTask(ctx, &p)
}

// HandleEmailTask wraps the common handler for Asynq
func HandleEmailTask(ctx context.Context, t *asynq.Task) error {
	var p common.EmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal email payload: %v", err)
	}
	return common.HandleEmailTask(ctx, &p)
}

// HandleServerInfoTask wraps the common handler for Asynq
func HandleServerInfoTask(ctx context.Context, t *asynq.Task) error {
	var p common.ServerInfoPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal server info payload: %v", err)
	}
	return common.HandleServerInfoTask(ctx, &p)
}

func main() {
	// Redis connection config
	redisAddr := "localhost:6380"
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		redisAddr = addr
	}

	redisConnOpt := asynq.RedisClientOpt{
		Addr: redisAddr,
	}
	// Support Redis password via environment variable
	if pwd := os.Getenv("REDIS_PASSWORD"); pwd != "" {
		redisConnOpt.Password = pwd
	}

	// Create client for enqueuing tasks
	client := asynq.NewClient(redisConnOpt)
	defer client.Close()

	// Create server for processing tasks
	srv := asynq.NewServer(
		redisConnOpt,
		asynq.Config{
			Concurrency: 5,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	// Register task handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(common.TypeWelcomeMessage, HandleWelcomeTask)
	mux.HandleFunc(common.TypeEmailTask, HandleEmailTask)
	mux.HandleFunc(common.TypeServerInfo, HandleServerInfoTask)

	fmt.Println("🚀 Starting Asynq Demo...")
	fmt.Printf("📍 Redis: %s\n", redisAddr)

	// Start consumer in background
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("🐰 Consumer started, waiting for tasks...")
		if err := srv.Run(mux); err != nil {
			log.Printf("❌ Consumer error: %v", err)
		}
	}()

	// Give consumer time to start
	time.Sleep(1 * time.Second)

	// Start scheduler for periodic tasks
	scheduler := asynq.NewScheduler(redisConnOpt, nil)

	// Register periodic server info task every 30 seconds
	serverInfoPayload := &common.ServerInfoPayload{
		Timestamp: time.Now().Unix(),
		Source:    "periodic-monitor",
	}

	payload, err := json.Marshal(serverInfoPayload)
	if err != nil {
		log.Printf("❌ Failed to marshal server info payload: %v", err)
	} else {
		if _, err := scheduler.Register("@every 30s", asynq.NewTask(common.TypeServerInfo, payload)); err != nil {
			log.Printf("❌ Failed to register server info scheduler: %v", err)
		} else {
			fmt.Println("⏰ Server info scheduler registered - runs every 30 seconds")
		}
	}

	// Start scheduler in background
	if err := scheduler.Start(); err != nil {
		log.Printf("❌ Failed to start scheduler: %v", err)
	}

	// Producer: Create sample welcome message tasks
	fmt.Println("📤 Creating welcome message tasks...")

	welcomeTasks := []common.WelcomePayload{
		{UserID: 1, Username: "Alice", Message: "Welcome to our amazing platform!"},
		{UserID: 2, Username: "Bob", Message: "Thanks for joining our community!"},
		{UserID: 3, Username: "Charlie", Message: "We're excited to have you here!"},
	}

	for i, task := range welcomeTasks {
		payload, err := json.Marshal(task)
		if err != nil {
			log.Printf("❌ Failed to marshal welcome task for %s: %v", task.Username, err)
			continue
		}

		var info *asynq.TaskInfo

		if i%2 == 0 {
			// Immediate task
			info, err = client.Enqueue(asynq.NewTask(common.TypeWelcomeMessage, payload))
		} else {
			// Delayed task
			info, err = client.Enqueue(asynq.NewTask(common.TypeWelcomeMessage, payload), asynq.ProcessIn(time.Duration(i)*300*time.Millisecond))
		}

		if err != nil {
			log.Printf("❌ Failed to enqueue welcome task for %s: %v", task.Username, err)
			continue
		}
		fmt.Printf("✅ Enqueued welcome task for %s (ID: %s)\n", task.Username, info.ID)
	}

	// Producer: Create sample email tasks
	fmt.Println("📤 Creating email tasks...")

	emailTasks := []common.EmailPayload{
		{UserID: 4, Email: "alice@example.com", Subject: "Welcome!", Message: "Welcome to our platform, Alice!"},
		{UserID: 5, Email: "bob@example.com", Subject: "Getting Started", Message: "Here are some tips to get you started, Bob."},
		{UserID: 6, Email: "charlie@example.com", Subject: "Weekly Newsletter", Message: "Check out this week's highlights!"},
	}

	for i, task := range emailTasks {
		payload, err := json.Marshal(task)
		if err != nil {
			log.Printf("❌ Failed to marshal email task for %s: %v", task.Email, err)
			continue
		}

		var info *asynq.TaskInfo

		if i%2 == 0 {
			// Immediate task
			info, err = client.Enqueue(asynq.NewTask(common.TypeEmailTask, payload))
		} else {
			// Delayed task
			info, err = client.Enqueue(asynq.NewTask(common.TypeEmailTask, payload), asynq.ProcessIn(time.Duration(i+1)*500*time.Millisecond))
		}

		if err != nil {
			log.Printf("❌ Failed to enqueue email task for %s: %v", task.Email, err)
			continue
		}
		fmt.Printf("✅ Enqueued email task for %s (ID: %s)\n", task.Email, info.ID)
	}

	fmt.Println("🎉 All tasks created! Consumer will process them shortly.")
	fmt.Println("Press Ctrl+C to stop...")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛑 Shutting down...")

	// Shutdown scheduler
	scheduler.Shutdown()

	// Shutdown server
	srv.Shutdown()

	// Wait for consumer to finish
	wg.Wait()

	fmt.Println("✅ Shutdown complete")
}
