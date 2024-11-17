package mr

import (
	"os"
	"strconv"
)

// 消息类型的枚举定义
type MsgType int

// 消息类型常量
const (
	RequestTask         MsgType = iota // Worker 请求任务
	AssignMapTask                      // Coordinator 分配 Map 任务
	AssignReduceTask                   // Coordinator 分配 Reduce 任务
	MapTaskCompleted                   // Worker 通知 Map 任务完成
	MapTaskFailed                      // Worker 通知 Map 任务失败
	ReduceTaskCompleted                // Worker 通知 Reduce 任务完成
	ReduceTaskFailed                   // Worker 通知 Reduce 任务失败
	CoordinatorEnd                     // Coordinator 通知 Worker 系统关闭
	CoordinatorWait                    // Coordinator 通知 Worker 等待下一任务
)

// 用于 Worker 向 Coordinator 发送消息
type WorkerRequest struct {
	Type   MsgType // 消息类型
	TaskID int     // 任务 ID，适用于任务完成或失败通知
}

// 用于 Coordinator 向 Worker 回复消息
type CoordinatorResponse struct {
	Type      MsgType // 消息类型
	TaskID    int     // 分配的任务 ID
	InputFile string  // Map 任务的输入文件路径
	NumReduce int     // Reduce 任务总数
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
