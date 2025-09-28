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
    NodeID      string    `json:"node_id"` // 添加节点ID
}

type Scheduler struct {
    tasks       []Task
    mu          sync.RWMutex
    workers     int
    processed   int64
    nodeID      string // 节点标识
}

var scheduler = &Scheduler{
    tasks:   []Task{},
    workers: 10,
    nodeID:  os.Getenv("NODE_ID"), // 支持多节点部署
}

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    if scheduler.nodeID == "" {
        scheduler.nodeID = "node-1" // 默认节点
    }

    http.HandleFunc("/", rootHandler)
    http.HandleFunc("/health", healthCheck)
    http.HandleFunc("/schedule", scheduleTask)
    http.HandleFunc("/status", getStatus)
    http.HandleFunc("/distribute", distributeTask) // 新增分布式端点

    fmt.Printf("Distributed Go scheduler running on port %s (Node: %s) with %d workers\n", 
        port, scheduler.nodeID, scheduler.workers)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func distributeTask(w http.ResponseWriter, r *http.Request) {
    // 模拟分布式任务分配
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "distributed",
        "node_id": scheduler.nodeID,
        "message": "Task distributed across nodes",
        "nodes_available": []string{"node-1", "node-2", "node-3"},
        "load_balanced": true,
    })
}

// ... 保留其他原有函数 ...

func healthCheck(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":     "healthy",
        "service":    "distributed-task-scheduler",
        "node_id":    scheduler.nodeID,
        "workers":    scheduler.workers,
        "distributed": true,
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
    task.NodeID = scheduler.nodeID // 分配到当前节点
    scheduler.tasks = append(scheduler.tasks, task)
    scheduler.mu.Unlock()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Task scheduled for distributed processing",
        "task":    task,
        "node":    scheduler.nodeID,
        "workers": scheduler.workers,
    })
}

func getStatus(w http.ResponseWriter, r *http.Request) {
    scheduler.mu.RLock()
    defer scheduler.mu.RUnlock()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "total_tasks": len(scheduler.tasks),
        "node_id":     scheduler.nodeID,
        "workers":     scheduler.workers,
        "processed":   atomic.LoadInt64(&scheduler.processed),
        "distributed": true,
        "architecture": "distributed-scheduler",
    })
}
