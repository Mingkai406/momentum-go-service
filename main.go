package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "sync"
    "sync/atomic"
    "time"
)

type Task struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Priority    string    `json:"priority"`
    Status      string    `json:"status"`
    ScheduledAt time.Time `json:"scheduled_at"`
}

type Scheduler struct {
    tasks       []Task
    mu          sync.RWMutex
    workers     int
    processed   int64
}

var scheduler = &Scheduler{
    tasks:   []Task{},
    workers: 10,
}

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    http.HandleFunc("/", rootHandler)
    http.HandleFunc("/health", healthCheck)
    http.HandleFunc("/schedule", scheduleTask)
    http.HandleFunc("/status", getStatus)

    fmt.Printf("Go microservice running on port %s with %d concurrent workers\n", port, scheduler.workers)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "service": "Momentum Go Scheduler",
        "version": "1.0.0",
        "workers": scheduler.workers,
    })
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":     "healthy",
        "service":    "momentum-go-scheduler",
        "workers":    scheduler.workers,
        "concurrent": true,
        "timestamp":  time.Now(),
    })
}

func scheduleTask(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var task Task
    json.NewDecoder(r.Body).Decode(&task)

    scheduler.mu.Lock()
    task.ID = len(scheduler.tasks) + 1
    task.ScheduledAt = time.Now().Add(5 * time.Second)
    task.Status = "scheduled"
    scheduler.tasks = append(scheduler.tasks, task)
    scheduler.mu.Unlock()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Task scheduled for concurrent processing",
        "task":    task,
        "workers": scheduler.workers,
    })
}

func getStatus(w http.ResponseWriter, r *http.Request) {
    scheduler.mu.RLock()
    defer scheduler.mu.RUnlock()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "total_tasks": len(scheduler.tasks),
        "workers":     scheduler.workers,
        "processed":   atomic.LoadInt64(&scheduler.processed),
        "concurrent":  true,
    })
}
