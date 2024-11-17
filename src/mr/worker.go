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
	// Read the input file
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

	// Call the user-defined Map function
	intermediate := mapf(inputFile, string(content))

	// Partition the intermediate key-value pairs into buckets
	buckets := make([][]KeyValue, nReduce)
	for _, kv := range intermediate {
		bucket := ihash(kv.Key) % nReduce
		buckets[bucket] = append(buckets[bucket], kv)
	}

	// Write the buckets to intermediate files
	for i, bucket := range buckets {
		fileName := fmt.Sprintf("mr-%d-%d", taskID, i)
		file, err := os.Create(fileName)
		if err != nil {
			log.Printf("Worker: Failed to create file %s: %v", fileName, err)
			return true
		}
		enc := json.NewEncoder(file)
		for _, kv := range bucket {
			if err := enc.Encode(&kv); err != nil {
				file.Close()
				log.Printf("Worker: Failed to write to file %s: %v", fileName, err)
				return true
			}
		}
		file.Close()
	}

	return false
}

// Handle a Reduce task
func handleReduceTask(taskID int, nReduce int, reducef func(string, []string) string) bool {
	// Read intermediate files
	intermediate := []KeyValue{}
	for i := 0; i < nReduce; i++ {
		fileName := fmt.Sprintf("mr-%d-%d", i, taskID)
		file, err := os.Open(fileName)
		if err != nil {
			log.Printf("Worker: Failed to open file %s: %v", fileName, err)
			return true
		}

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

	// Sort intermediate results by key
	sortByKey(intermediate)

	// Group values by key and write the output
	outputFile := fmt.Sprintf("mr-out-%d", taskID)
	file, err := os.Create(outputFile)
	if err != nil {
		log.Printf("Worker: Failed to create output file %s: %v", outputFile, err)
		return true
	}
	defer file.Close()

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
		fmt.Fprintf(file, "%v %v\n", intermediate[i].Key, result)
		i = j
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
