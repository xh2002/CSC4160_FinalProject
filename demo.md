## **展示视频大纲**

### **1. 开场介绍**
- **项目名称**：基于云的分布式 MapReduce 系统  
- **主要功能**：
  - **Mapreduce 实现**：基于Go实现了Google论文中的Mapreduce框架
  - **云环境集成**：将 MapReduce 部署到 AWS EC2，并尽量使用 S3 存储数据。
  - **自动扩展与容错**：实现自动恢复机制和任务重分配。
  - **性能测试**：通过 CloudWatch 可视化系统性能。
台词：
```
Hello everyone, and welcome to the demonstration of our CSC4160 Cloud Computing project: A Cloud-Based Distributed MapReduce System.
This project aims to showcase the power of distributed and cloud computing by running the MapReduce framework in AWS environment.
We have implemented four key functionalities:
1. Mapreduce Implementation: Implemented the Mapreduce framework in Google's paper based on Go on our devices.
2. Cloud Environment Integration: Deploying MapReduce on AWS EC2 and utilizing S3 for data storage as much as we can.
3. Auto Scaling and Fault Tolerance: Ensuring the system can recover from failures and redistribute tasks efficiently.
4. Performance Testing: Visualizing the system's performance and scalability using AWS CloudWatch metrics.
Now, let's dive into the demonstration."
```
---

### **2. 云环境集成**
- **演示内容**：
  1. 展示 Python 脚本（`./py/download.py`）从开放的 S3 桶下载输入数据的过程。
  2. 在 EC2 中启动 MapReduce 的 `Coordinator` 和多个 `Worker` 进程。
  3. 原始的数据map到
  4. 执行 MapReduce 任务：
     - 启动 `Coordinator`，指定输入文件。
     - 启动 `Worker`，完成 Map 和 Reduce 任务。
  5. 展示输出文件 `mr-out-*` 的生成，验证任务结果。

---

### **3. 自动扩展与容错**
- **演示内容**：
  **Worker 故障场景**：
     - 在任务运行过程中，手动停止某个 `Worker` 进程。
     - 观察 `Coordinator` 日志，展示任务重分配的过程。
     - （日志中的 **skipping** 标识即为故障任务重新分配）。
     - 系统鲁棒性：系统能够在 Worker 故障或实例宕机的情况下，继续完成所有任务。

---

### **4. 性能测试**
- **演示内容**：
  1. 使用 **CloudWatch** 监控 EC2 实例的运行状态：
     - 展示 CPU、内存和网络流量的实时变化。
  2. 在仪表板中展示任务执行过程中的资源使用情况。
  3. 比较不同规模任务（小数据集 vs 大数据集）的完成时间。
- **展示重点**：
  - CloudWatch 图表中的资源利用率变化。
  - 系统处理不同任务规模的性能表现。

---

### **5. 结尾总结**
- **系统特点**：
  1. 分布式架构支持高效的任务处理。
  2. 自动恢复和任务重分配保证系统容错性。
  3. 使用云环境提升系统的扩展性和可靠性。
- **进步空间**：
  - 进一步使用 S3 提升中间文件处理效率，但由于权限问题暂未实现。

```
"In conclusion, our system demonstrates the following key features:

Distributed architecture that enables efficient task processing across multiple workers.
Fault tolerance and auto-recovery, ensuring the system remains robust even in the event of failures.
Cloud integration that enhances scalability and reliability by leveraging AWS infrastructure.
As for future improvements, we aim to further optimize intermediate file handling using S3 for enhanced efficiency. However, due to access restrictions, this could not be fully implemented in the current project.

That's all for our demonstration, thank you for watching!"
```

---

### **展示时长**
- 开场介绍：30秒  
- 云环境集成：1分钟  
- 自动扩展与容错：1分钟  
- 性能测试：1分30秒  
- 结尾总结：30秒  
- **总时长**：约4分钟
