## **展示视频大纲**

### **1. 开场介绍**
- **项目名称**：基于云的分布式 MapReduce 系统  
- **主要功能**：
  - **Mapreduce 实现**：基于Go实现了Google论文中的Mapreduce框架
  - **云环境集成**：将 MapReduce 部署到 AWS EC2，并尽量使用 S3 存储数据。
  - **自动扩展与容错**：实现自动恢复机制和任务重分配。
  - **性能测试**：通过 CloudWatch 可视化系统性能。
台词：
Hello everyone, and welcome to the demonstration of our CSC4160 Cloud Computing project: A Cloud-Based Distributed MapReduce System.
This demo aims to run a wordcount task on our Cloud-Based Distributed MapReduce System.
We will show three parts:
1. Cloud Environment Integration: Deploying MapReduce on AWS EC2 and utilizing S3 for data storage.
2. Wordcount execution : Proof that we have implement Distributed MapReduce framework.
3. Fault Tolerance: Ensuring the system can recover from failures and redistribute tasks efficiently.
Now, let's dive into the demonstration."


---

### **2. 云环境集成**Cloud environment integration
- **演示内容**：
  1. 展示 Python 脚本（`./py/download.py`）从开放的 S3 桶下载输入数据的过程。   
  2. 在 EC2 中启动 MapReduce 的 `Coordinator` 和多个 `Worker` 进程。
  3. 展示pg-all文件，说我们将会用这个文件MapReduce以统计词汇出现的频率。We will use this file with MapReduce to count the frequency of word occurrences.
  4. 执行 MapReduce 任务：
     - 启动 `Coordinator`，to process this pg-all file。
     - 启动 `Worker`，完成 Map 和 Reduce 任务。
     - Map出来的文件在mr-xx上，展示2个分布式文件，告诉他这是map之后的文件。
  5. 展示输出文件 `mr-out-*` 的生成（这是reduce的结果，），验证任务结果。
     对于map和reduce的，因为他们是分布式 distributed 的，所以分布 distribut 在各个mr-out-

---

### **3. 容错** fault tolerance
- **演示内容**：
  **Worker 故障场景**：
     - 观察 `worker` 日志，日志中的 skipping 标识即意味着遇到故障、任务重新分配。By observing the `worker` logs, the presence of the "skipping" indicator shows that a failure has occurred and the task has been reassigned.
     - 如果在运行大文件的时候中断（包括键盘Ctrl+C中断）也可以正常恢复。Our MapReduce system also supports recovery in case of interruptions while processing large files, including manual interruptions of workers.
     - It help us achieve our requirement of System Robustness, where the system can continue to complete all tasks even in the event of Worker failures or instance crashes.

---

### **4 结尾总结**
- **系统特点**：
  1. 分布式架构支持高效的任务处理。
  2. 自动恢复和任务重分配保证系统容错性。
  3. 使用云环境提升系统的扩展性和可靠性。
- **进步空间**：
  - 进一步使用 S3 提升中间文件处理效率，但由于权限问题暂未实现。

```
"In conclusion, our system demonstrates the following key features:

Distributed architecture that enables efficient task processing across multiple workers.
Fault tolerance ensuring the system remains robust even in the event of failures.
Cloud integration that enhances scalability and reliability by leveraging AWS infrastructure.

That's all for our demonstration, thank you for watching!"
```
