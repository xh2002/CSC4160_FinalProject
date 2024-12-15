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

// Define task status
type taskStatus int

const (
	idle     taskStatus = iota // Not assigned
	running                    // Running
	finished                   // Completed
	failed                     // Failed
)

// Map task information
type MapTaskInfo struct {
	TaskID    int        // Task ID
	Status    taskStatus // Task status
	StartTime time.Time  // Task start time
	InputFile string     // Input file name
}

// Reduce task information
type ReduceTaskInfo struct {
	TaskID    int        // Task ID
	Status    taskStatus // Task status
	StartTime time.Time  // Task start time
}

// Coordinator struct
type Coordinator struct {
	mu          sync.Mutex              // Lock to protect shared data
	mapTasks    map[int]*MapTaskInfo    // All Map tasks
	reduceTasks map[int]*ReduceTaskInfo // All Reduce tasks
	nReduce     int                     // Number of Reduce tasks
	mapDone     bool                    // Whether all Map tasks are completed
	reduceDone  bool                    // Whether all Reduce tasks are completed
	allDone     bool                    // Whether all tasks are completed
	taskTimeout time.Duration           // Task timeout duration
}

// Initialize tasks
func (c *Coordinator) initTasks(files []string) {
	c.mapTasks = make(map[int]*MapTaskInfo)
	c.reduceTasks = make(map[int]*ReduceTaskInfo)

	// Initialize Map tasks
	for i, file := range files {
		c.mapTasks[i] = &MapTaskInfo{
			TaskID:    i,
			Status:    idle,
			InputFile: file,
		}
	}

	// Initialize Reduce tasks
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

	// Assign Map tasks
	if !c.mapDone {
		for _, task := range c.mapTasks {
			if task.Status == idle || (task.Status == running && time.Since(task.StartTime) > c.taskTimeout) {
				// Assign task
				task.Status = running
				task.StartTime = time.Now()
				reply.Type = AssignMapTask
				reply.TaskID = task.TaskID
				reply.InputFile = task.InputFile
				reply.NumReduce = c.nReduce

				// Log assignment
				log.Printf("Coordinator: Assigned Map Task %d with file %s to Worker", task.TaskID, task.InputFile)
				return nil
			}
		}

		// If Map tasks are not completed but no tasks are available to assign
		reply.Type = CoordinatorWait
		return nil
	}

	// Assign Reduce tasks
	if c.mapDone && !c.reduceDone {
		for _, task := range c.reduceTasks {
			if task.Status == idle || (task.Status == running && time.Since(task.StartTime) > c.taskTimeout) {
				// Assign task
				task.Status = running
				task.StartTime = time.Now()
				reply.Type = AssignReduceTask
				reply.TaskID = task.TaskID
				reply.NumReduce = c.nReduce
				return nil
			}
		}

		// If Reduce tasks are not completed but no tasks are available to assign
		reply.Type = CoordinatorWait
		return nil
	}

	// Notify Worker to exit
	reply.Type = CoordinatorEnd
	c.allDone = true
	return nil
}

func (c *Coordinator) monitorTimeouts() {
	for {
		time.Sleep(time.Second) // Check task status every second
		c.mu.Lock()

		// Check if Map tasks are timed out
		if !c.mapDone {
			for _, task := range c.mapTasks {
				if task.Status == running && time.Since(task.StartTime) > c.taskTimeout {
					task.Status = idle // Timeout, reset task to idle
				}
			}
		}

		// Check if Reduce tasks are timed out
		if c.mapDone && !c.reduceDone {
			for _, task := range c.reduceTasks {
				if task.Status == running && time.Since(task.StartTime) > c.taskTimeout {
					task.Status = idle // Timeout, reset task to idle
				}
			}
		}

		c.mu.Unlock()
	}
}

// RPC handler: Worker reports task status
func (c *Coordinator) NoticeResult(args *WorkerRequest, reply *struct{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if args.Type == MapTaskCompleted {
		// Update Map task status
		if task, ok := c.mapTasks[args.TaskID]; ok && task.Status == running {
			task.Status = finished
		}

		// Check if all Map tasks are completed
		c.mapDone = c.checkAllMapTasksDone()
	} else if args.Type == ReduceTaskCompleted {
		// Update Reduce task status
		if task, ok := c.reduceTasks[args.TaskID]; ok && task.Status == running {
			task.Status = finished
		}

		// Check if all Reduce tasks are completed
		c.reduceDone = c.checkAllReduceTasksDone()
	}

	return nil
}

// Check if all Map tasks are completed
func (c *Coordinator) checkAllMapTasksDone() bool {
	for _, task := range c.mapTasks {
		if task.Status != finished {
			return false
		}
	}
	return true
}

// Check if all Reduce tasks are completed
func (c *Coordinator) checkAllReduceTasksDone() bool {
	for _, task := range c.reduceTasks {
		if task.Status != finished {
			return false
		}
	}
	return true
}

// Start RPC server
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

// Check if all tasks are completed
func (c *Coordinator) Done() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.allDone
}

// Create Coordinator
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		nReduce:     nReduce,
		taskTimeout: 10 * time.Second, // Task timeout duration is 10 seconds
	}

	// Initialize tasks
	c.initTasks(files)

	// Start RPC server
	go c.monitorTimeouts()
	c.server()
	return &c
}
