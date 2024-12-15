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

// KeyValue is a type used by Map functions to return a slice of key/value pairs.
type KeyValue struct {
    Key   string
    Value string
}

// ihash uses ihash(key) % NReduce to choose the reduce task number for each KeyValue emitted by Map.
func ihash(key string) int {
    h := fnv.New32a()
    h.Write([]byte(key))
    return int(h.Sum32() & 0x7fffffff)
}

// Worker is called by main/mrworker.go.
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

// handleMapTask handles a Map task.
func handleMapTask(taskID int, inputFile string, nReduce int, mapf func(string, string) []KeyValue) bool {
    // Read input file
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

    // Call user-defined Map function
    intermediate := mapf(inputFile, string(content))

    // Partition intermediate results into nReduce buckets
    buckets := make([][]KeyValue, nReduce)
    for _, kv := range intermediate {
        bucket := ihash(kv.Key) % nReduce
        buckets[bucket] = append(buckets[bucket], kv)
    }

    // Create intermediate files for each bucket
    for i, bucket := range buckets {
        // Use temporary files
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

        // Rename temporary file to final file
        if err := os.Rename(tmpFileName, fileName); err != nil {
            log.Printf("Worker: Failed to rename tmp file %s to %s: %v", tmpFileName, fileName, err)
            return true
        }
    }

    return false
}

// handleReduceTask handles a Reduce task.
func handleReduceTask(taskID int, nReduce int, reducef func(string, []string) string) bool {
    // Read intermediate files
    intermediate := []KeyValue{}
    for i := 0; i < nReduce; i++ {
        fileName := fmt.Sprintf("mr-%d-%d", i, taskID)
        file, err := os.Open(fileName)
        if err != nil {
            // If file does not exist, skip
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

    // If no intermediate data, skip
    if len(intermediate) == 0 {
        log.Printf("Worker: No intermediate data for Reduce Task %d, skipping...", taskID)
        return false
    }

    // Sort by key
    sortByKey(intermediate)

    // Create temporary output file
    tmpOutputFileName := fmt.Sprintf("mr-out-%d-tmp", taskID)
    outputFileName := fmt.Sprintf("mr-out-%d", taskID)
    tmpFile, err := os.Create(tmpOutputFileName)
    if err != nil {
        log.Printf("Worker: Failed to create tmp output file %s: %v", tmpOutputFileName, err)
        return true
    }
    defer tmpFile.Close()

    // Call Reduce function for each key and write results
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
        // Write to temporary file
        fmt.Fprintf(tmpFile, "%v %v\n", intermediate[i].Key, result)
        i = j
    }

    // Rename temporary file to final file
    if err := os.Rename(tmpOutputFileName, outputFileName); err != nil {
        log.Printf("Worker: Failed to rename tmp output file %s to %s: %v", tmpOutputFileName, outputFileName, err)
        return true
    }

    return false
}

// sortByKey sorts intermediate results by key.
func sortByKey(kva []KeyValue) {
    sort.Slice(kva, func(i, j int) bool {
        return kva[i].Key < kva[j].Key
    })
}

// CallForReportStatus reports task status to the Coordinator.
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

// CallForTask requests a task from the Coordinator.
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

// call sends an RPC request to the coordinator and waits for the response.
// It usually returns true, and returns false if something goes wrong.
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