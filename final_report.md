![alt text](image.png)

环境：
EC2: m5.large, Ubuntu 24.04 LTS
S3: same region as EC2
gcc: version (Ubuntu 13.3.0-6ubuntu2~24.04) 13.3.0
go: version go1.21.1 linux/amd64


1. 云环境集成（完成）
EC2新建实例，写好的代码上传到EC2

S3新建桶，上传需要进行MapReduce的数据
（没有权限访问，因为IAM没法新建Users）
写了个python脚本 ./py/download.py 从开放的s3桶里抓取下载需要跑的文件保存在./src/main

go构建：go build -buildmode=plugin wc.go
删除之前的输出：rm mr-out*
coordinator 运行需要MapReduce的：go run mrcoordinator.go pg-*.txt
另开一个命令行，作为worker，运行go run mrworker.go ../mrapps/wc.so
检查输出结果：ls mr-out- *

2. 自动扩展与容错

  (1) 实现 Worker 容错机制 (未实现)
    在 Coordinator 中实现任务重新分配逻辑。当某个 Worker 失败时，未完成的任务会被重新分配给其他 Worker。
    在 Coordinator 的任务管理逻辑中，定期检查 Worker 的心跳，如果 Worker 长时间未完成任务，标记该任务为失败并重新分配。
  ```
      func checkWorkerHealth() {
          for {
              time.Sleep(10 * time.Second)
              for taskID, task := range tasks {
                  if time.Since(task.StartTime) > TaskTimeout && !task.Completed {
                      fmt.Printf("任务 %d 超时，重新分配\n", taskID)
                      reassignTask(taskID)
                  }
              }
          }
      }
      启动 checkWorkerHealth 协程：

      go
      复制代码
      go checkWorkerHealth()
  ```
  (2) 集成 AWS Auto Recovery
    配置 EC2 实例的自动恢复机制，以应对硬件故障或实例宕机
    ![alt text](image-1.png)
  
  (3) 动态启动/停止 Worker 的优化
    不行，前提条件是确保 EC2 实例关联了一个 IAM 角色，没有权限访问

3. 安全性
    如前文所述，配置不了IAM所以也干不了，淦。
  
4. 分布式存储
    如前文所述，配置不了IAM，用不了S3所以也干不了，淦。

5. 性能测试与优化

   (1). 性能测试

  CloudWatch 监控记录一下吧 类似于之前做的assignment



  (2). 优化任务分配

    2.1 实现负载均衡策略
在 `Coordinator` 中改进任务分配逻辑：
1. **追踪 Worker 负载**：
   为每个 Worker 记录正在处理的任务数量。
2. **优先分配给空闲 Worker**：
   分配任务时，优先选择任务数量最少的 Worker。

示例代码修改：
```go
func assignTaskToWorker(workerID string, task Task) {
    if workers[workerID].TaskCount < MaxTaskPerWorker {
        workers[workerID].TaskCount++
        task.Assigned = true
        fmt.Printf("分配任务 %v 给 Worker %v\n", task.ID, workerID)
    }
}
```

#### **2.2 动态任务重新分配**
如果某个 Worker 超时未完成任务，`Coordinator` 将该任务重新分配：
```go
func reassignTimeoutTasks() {
    for _, task := range tasks {
        if task.Assigned && time.Since(task.StartTime) > Timeout {
            task.Assigned = false
            fmt.Printf("任务 %v 超时，重新分配\n", task.ID)
        }
    }
}
```
  （3） 优化中间文件存储
    之后的都做不了


