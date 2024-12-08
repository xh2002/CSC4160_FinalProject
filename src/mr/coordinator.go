package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

// 定义任务状态
type taskStatus int

const (
	idle     taskStatus = iota // 未分配
	running                    // 正在运行
	finished                   // 已完成
	failed                     // 失败
)

// Map 任务信息
type MapTaskInfo struct {
	TaskID    int        // 任务 ID
	Status    taskStatus // 任务状态
	StartTime time.Time  // 任务开始时间
	InputFile string     // 输入文件名
}

// Reduce 任务信息
type ReduceTaskInfo struct {
	TaskID    int        // 任务 ID
	Status    taskStatus // 任务状态
	StartTime time.Time  // 任务开始时间
}

// Coordinator 结构体
type Coordinator struct {
	mu          sync.Mutex              // 保护共享数据的锁
	mapTasks    map[int]*MapTaskInfo    // 所有 Map 任务
	reduceTasks map[int]*ReduceTaskInfo // 所有 Reduce 任务
	nReduce     int                     // Reduce 任务数量
	mapDone     bool                    // 是否所有 Map 任务完成
	reduceDone  bool                    // 是否所有 Reduce 任务完成
	allDone     bool                    // 是否所有任务完成
	taskTimeout time.Duration           // 任务超时时间
}

// 初始化任务
func (c *Coordinator) initTasks(files []string) {
	c.mapTasks = make(map[int]*MapTaskInfo)
	c.reduceTasks = make(map[int]*ReduceTaskInfo)

	// 初始化 Map 任务
	for i, file := range files {
		c.mapTasks[i] = &MapTaskInfo{
			TaskID:    i,
			Status:    idle,
			InputFile: file,
		}
	}

	// 初始化 Reduce 任务
	for i := 0; i < c.nReduce; i++ {
		c.reduceTasks[i] = &ReduceTaskInfo{
			TaskID: i,
			Status: idle,
		}
	}
}

func (c *Coordinator) AskForTask(args *WorkerRequest, reply *CoordinatorResponse) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 分配 Map 任务
	if !c.mapDone {
		for _, task := range c.mapTasks {
			if task.Status == idle || (task.Status == running && time.Since(task.StartTime) > c.taskTimeout) {
				// 分配任务
				task.Status = running
				task.StartTime = time.Now()
				reply.Type = AssignMapTask
				reply.TaskID = task.TaskID
				reply.InputFile = task.InputFile
				reply.NumReduce = c.nReduce

				// 添加日志
				log.Printf("Coordinator: Assigned Map Task %d with file %s to Worker", task.TaskID, task.InputFile)
				return nil
			}
		}

		// 如果 Map 任务还未完成，但没有可分配的任务
		reply.Type = CoordinatorWait
		return nil
	}

	// 分配 Reduce 任务
	if c.mapDone && !c.reduceDone {
		for _, task := range c.reduceTasks {
			if task.Status == idle || (task.Status == running && time.Since(task.StartTime) > c.taskTimeout) {
				// 分配任务
				task.Status = running
				task.StartTime = time.Now()
				reply.Type = AssignReduceTask
				reply.TaskID = task.TaskID
				reply.NumReduce = c.nReduce
				return nil
			}
		}

		// 如果 Reduce 任务还未完成，但没有可分配的任务
		reply.Type = CoordinatorWait
		return nil
	}

	// 通知 Worker 退出
	reply.Type = CoordinatorEnd
	c.allDone = true
	return nil
}

func (c *Coordinator) monitorTimeouts() {
	for {
		time.Sleep(time.Second) // 每秒检查一次任务状态
		c.mu.Lock()

		// 检查 Map 任务是否超时
		if !c.mapDone {
			for _, task := range c.mapTasks {
				if task.Status == running && time.Since(task.StartTime) > c.taskTimeout {
					task.Status = idle // 超时，重置任务为未分配状态
				}
			}
		}

		// 检查 Reduce 任务是否超时
		if c.mapDone && !c.reduceDone {
			for _, task := range c.reduceTasks {
				if task.Status == running && time.Since(task.StartTime) > c.taskTimeout {
					task.Status = idle // 超时，重置任务为未分配状态
				}
			}
		}

		c.mu.Unlock()
	}
}

// RPC 处理：Worker 汇报任务状态
func (c *Coordinator) NoticeResult(args *WorkerRequest, reply *struct{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if args.Type == MapTaskCompleted {
		// 更新 Map 任务状态
		if task, ok := c.mapTasks[args.TaskID]; ok && task.Status == running {
			task.Status = finished
		}

		// 检查是否所有 Map 任务完成
		c.mapDone = c.checkAllMapTasksDone()
	} else if args.Type == ReduceTaskCompleted {
		// 更新 Reduce 任务状态
		if task, ok := c.reduceTasks[args.TaskID]; ok && task.Status == running {
			task.Status = finished
		}

		// 检查是否所有 Reduce 任务完成
		c.reduceDone = c.checkAllReduceTasksDone()
	}

	return nil
}

// 检查是否所有 Map 任务完成
func (c *Coordinator) checkAllMapTasksDone() bool {
	for _, task := range c.mapTasks {
		if task.Status != finished {
			return false
		}
	}
	return true
}

// 检查是否所有 Reduce 任务完成
func (c *Coordinator) checkAllReduceTasksDone() bool {
	for _, task := range c.reduceTasks {
		if task.Status != finished {
			return false
		}
	}
	return true
}

// RPC 服务启动
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// 检查是否所有任务完成
func (c *Coordinator) Done() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.allDone
}

// 创建 Coordinator
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		nReduce:     nReduce,
		taskTimeout: 10 * time.Second, // 任务超时时间为 10 秒
	}

	// 初始化任务
	c.initTasks(files)

	// 启动 RPC 服务
	go c.monitorTimeouts()
	c.server()
	return &c
}
