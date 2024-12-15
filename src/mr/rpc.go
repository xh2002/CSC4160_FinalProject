package mr

import (
    "os"
    "strconv"
)

// Enumeration definition for message types
type MsgType int

// Constants for message types
const (
    RequestTask         MsgType = iota // Worker requests a task
    AssignMapTask                      // Coordinator assigns a Map task
    AssignReduceTask                   // Coordinator assigns a Reduce task
    MapTaskCompleted                   // Worker notifies that a Map task is completed
    MapTaskFailed                      // Worker notifies that a Map task has failed
    ReduceTaskCompleted                // Worker notifies that a Reduce task is completed
    ReduceTaskFailed                   // Worker notifies that a Reduce task has failed
    CoordinatorEnd                     // Coordinator notifies Worker to shut down
    CoordinatorWait                    // Coordinator notifies Worker to wait for the next task
)

// Used by Worker to send messages to Coordinator
type WorkerRequest struct {
    Type   MsgType // Message type
    TaskID int     // Task ID, applicable for task completion or failure notifications
}

// Used by Coordinator to reply to Worker
type CoordinatorResponse struct {
    Type      MsgType // Message type
    TaskID    int     // Assigned task ID
    InputFile string  // Input file path for Map task
    NumReduce int     // Total number of Reduce tasks
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
    s := "/var/tmp/5840-mr-"
    s += strconv.Itoa(os.Getuid())
    return s
}