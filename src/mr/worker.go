package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	for {
		// Request a task from the coordinator
		task := CallForTask()
		if task == nil {
			log.Println("Worker: No task received. Exiting...")
			break
		}

		switch task.Type {
		case AssignMapTask:
			// Handle a Map task
			err := handleMapTask(task.TaskID, task.InputFile, task.NumReduce, mapf)
			if err {
				CallForReportStatus(MapTaskFailed, task.TaskID)
			} else {
				CallForReportStatus(MapTaskCompleted, task.TaskID)
			}
		case AssignReduceTask:
			// Handle a Reduce task
			err := handleReduceTask(task.TaskID, task.NumReduce, reducef)
			if err {
				CallForReportStatus(ReduceTaskFailed, task.TaskID)
			} else {
				CallForReportStatus(ReduceTaskCompleted, task.TaskID)
			}
		case CoordinatorWait:
			// No task available, wait before retrying
			time.Sleep(time.Second)
		case CoordinatorEnd:
			// All tasks are completed, exit the worker
			log.Println("Worker: All tasks completed. Exiting...")
			return
		default:
			log.Printf("Worker: Unknown task type %v", task.Type)
		}
	}
}

// Handle a Map task
func handleMapTask(taskID int, inputFile string, nReduce int, mapf func(string, string) []KeyValue) bool {
	// 读取输入文件
	file, err := os.Open(inputFile)
	if err != nil {
		log.Printf("Worker: Failed to open file %s: %v", inputFile, err)
		return true
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("Worker: Failed to read file %s: %v", inputFile, err)
		return true
	}

	// 调用用户定义的 Map 函数
	intermediate := mapf(inputFile, string(content))

	// 按键分区，将中间结果分成 nReduce 个桶
	buckets := make([][]KeyValue, nReduce)
	for _, kv := range intermediate {
		bucket := ihash(kv.Key) % nReduce
		buckets[bucket] = append(buckets[bucket], kv)
	}

	// 为每个桶创建对应的中间文件
	for i, bucket := range buckets {
		// 使用临时文件
		tmpFileName := fmt.Sprintf("mr-%d-%d-tmp", taskID, i)
		fileName := fmt.Sprintf("mr-%d-%d", taskID, i)
		tmpFile, err := os.Create(tmpFileName)
		if err != nil {
			log.Printf("Worker: Failed to create tmp file %s: %v", tmpFileName, err)
			return true
		}
		enc := json.NewEncoder(tmpFile)
		for _, kv := range bucket {
			if err := enc.Encode(&kv); err != nil {
				tmpFile.Close()
				log.Printf("Worker: Failed to write to tmp file %s: %v", tmpFileName, err)
				return true
			}
		}
		tmpFile.Close()

		// 重命名临时文件为正式文件
		if err := os.Rename(tmpFileName, fileName); err != nil {
			log.Printf("Worker: Failed to rename tmp file %s to %s: %v", tmpFileName, fileName, err)
			return true
		}
	}

	return false
}

func handleReduceTask(taskID int, nReduce int, reducef func(string, []string) string) bool {
	// 读取中间文件
	intermediate := []KeyValue{}
	for i := 0; i < nReduce; i++ {
		fileName := fmt.Sprintf("mr-%d-%d", i, taskID)
		file, err := os.Open(fileName)
		if err != nil {
			// 如果文件不存在，跳过
			log.Printf("Worker: File %s does not exist, skipping...", fileName)
			continue
		}
		log.Printf("Worker: Successfully opened file %s for Reduce Task %d", fileName, taskID)

		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			intermediate = append(intermediate, kv)
		}
		file.Close()
	}

	// 如果没有中间数据，则直接返回
	if len(intermediate) == 0 {
		log.Printf("Worker: No intermediate data for Reduce Task %d, skipping...", taskID)
		return false
	}

	// 按键排序
	sortByKey(intermediate)

	// 创建临时输出文件
	tmpOutputFileName := fmt.Sprintf("mr-out-%d-tmp", taskID)
	outputFileName := fmt.Sprintf("mr-out-%d", taskID)
	tmpFile, err := os.Create(tmpOutputFileName)
	if err != nil {
		log.Printf("Worker: Failed to create tmp output file %s: %v", tmpOutputFileName, err)
		return true
	}
	defer tmpFile.Close()

	// 按键调用 Reduce 函数并写入结果
	for i := 0; i < len(intermediate); {
		j := i + 1
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}
		result := reducef(intermediate[i].Key, values)
		// 写入临时文件
		fmt.Fprintf(tmpFile, "%v %v\n", intermediate[i].Key, result)
		i = j
	}

	// 重命名临时文件为正式文件
	if err := os.Rename(tmpOutputFileName, outputFileName); err != nil {
		log.Printf("Worker: Failed to rename tmp output file %s to %s: %v", tmpOutputFileName, outputFileName, err)
		return true
	}

	return false
}

// Sort intermediate results by key
func sortByKey(kva []KeyValue) {
	sort.Slice(kva, func(i, j int) bool {
		return kva[i].Key < kva[j].Key
	})
}

// CallForReportStatus 用于向 Coordinator 报告任务状态
func CallForReportStatus(successType MsgType, taskID int) {
	args := WorkerRequest{
		Type:   successType,
		TaskID: taskID,
	}

	ok := call("Coordinator.NoticeResult", &args, nil)
	if !ok {
		log.Printf("Worker: Failed to report task status %v for task %v", successType, taskID)
	} else {
		log.Printf("Worker: Successfully reported task status %v for task %v", successType, taskID)
	}
}

// CallForTask 向 Coordinator 请求任务
func CallForTask() *CoordinatorResponse {
	args := WorkerRequest{Type: RequestTask}
	reply := CoordinatorResponse{}

	ok := call("Coordinator.AskForTask", &args, &reply)
	if !ok {
		log.Println("Worker: Failed to request task.")
		return nil
	}

	return &reply
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	log.Println(err)
	return false
}
