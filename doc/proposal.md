# CSC4160 Final Project: Cloud-Enabled Distributed MapReduce System Implementation

> **Team Members:**
> - Danyang Chen (123090018)
> - Boshi Xu (122040075)

## Topic and Background

In this project, we will implement a cloud-enabled distributed MapReduce system, inspired by the MapReduce framework introduced by Google. MapReduce is a programming model used for processing and generating large datasets by dividing tasks into smaller subtasks, which are then executed in parallel across multiple nodes. The original concept of MapReduce was detailed in a paper by Jeffrey Dean and Sanjay Ghemawat, presented at **OSDI '04: 6th Symposium on Operating Systems Design and Implementation, USENIX Association**, which highlighted its efficiency in handling large-scale data processing by utilizing a master-worker architecture [1].

This project focuses on building a cloud-enabled distributed system, leveraging modern cloud computing technologies to achieve scalability, fault tolerance, and high availability. Cloud platforms like **AWS** provide the essential infrastructure needed to run large-scale MapReduce operations, utilizing cloud services such as **Amazon EC2**, **Amazon S3**, and **AWS Auto Scaling**. This allows for seamless scaling of resources based on workload, ensuring both performance and cost efficiency.

## Project Description

### Two Core Components

1. **Coordinator**:
   - Responsible for handing out tasks to workers, keeping track of task progress, and redistributing tasks if a worker fails or takes too long.
   - Starts an RPC (Remote Procedure Call) server to manage communication with workers.
   - Manages distributed task execution across cloud instances, monitors workers for failure using health checks, and reallocates tasks to ensure completion. This follows cloud best practices of **resiliency** and **fault tolerance** by incorporating retry mechanisms and automatic re-execution in the case of failures.

2. **Workers**:
   - Handle application-specific **Map** and **Reduce** functions, which involve reading input data from cloud storage, processing it, and writing output back to cloud storage.
   - Each worker executes in a loop to:
     1. Request a task from the coordinator using RPC.
     2. Read input data from **Amazon S3** or a similar cloud storage service.
     3. Perform the **Map** or **Reduce** operation.
     4. Write the resulting output to cloud storage.
     5. Request the next available task from the coordinator.

The **coordinator** will monitor the progress of tasks and reassign tasks as necessary, ensuring tasks that are not completed within a given time window are redistributed to healthy workers. By running the system on **Amazon EC2**, we can simulate real-world distributed environments where data and compute nodes are spread across different virtual machines, with the **Auto Scaling** feature dynamically adjusting the number of worker instances based on the computational load.

### Cloud Integration:

1. **Deploying Coordinator and Worker Instances on EC2**:
   - The system will deploy the coordinator and multiple worker processes across **Amazon EC2** instances, allowing the MapReduce framework to be run in a truly distributed environment. EC2 provides the necessary compute power, with the flexibility to scale up or down based on the current needs of the MapReduce job. Each EC2 instance will handle a portion of the MapReduce tasks, making it easy to simulate a large cluster environment in the cloud.

2. **Auto Scaling**:
   - One of the key benefits of running the MapReduce system in the cloud is the use of **AWS Auto Scaling** to automatically adjust the number of worker instances based on demand. When the workload increases, Auto Scaling can launch additional worker instances to distribute the computational load, and when the workload decreases, it can terminate instances to save costs. This ensures **elastic scalability**, allowing the system to adapt to changing workloads and optimize resource usage, which is one of the major advantages of cloud computing.

3. **Distributed Storage with Amazon S3**:
   - In cloud environments, distributed storage systems like **Amazon S3** offer scalable, reliable, and highly available storage for large datasets. The input data for Map tasks, intermediate data generated during processing, and final output data from Reduce tasks will be stored in **S3 buckets**. S3’s capabilities in handling large datasets, versioning, and access control make it ideal for MapReduce operations.

4. **Network Communication**:
   - Workers will communicate with the coordinator via **RPC over cloud networks**, and the cloud environment provides the necessary infrastructure for secure, high-speed communication. **AWS VPC (Virtual Private Cloud)** and security groups will ensure isolated and secure communication between instances, with the necessary firewall rules and network policies in place.


5. **(Optional) Fault Tolerance**:
   - Fault tolerance in the cloud environment can be ensured through **EC2 instance health checks** and **Auto Recovery**. If a worker instance becomes unhealthy, AWS automatically detects the issue, and replacement instances are launched. This prevents job failures from affecting the overall system performance and ensures continuous operation without manual intervention. The system's ability to handle worker failures gracefully is a critical aspect of running robust cloud-based distributed applications.

6. **(Optional) Data Security and Access Control**:
   - Leveraging **IAM (Identity and Access Management)** roles, the system can control access to sensitive data stored in **Amazon S3**. Each worker instance will have permission only to access the relevant data, ensuring a secure and controlled data pipeline. Encryption of data both at rest (in S3) and in transit (across EC2 instances) can be implemented using **AWS KMS (Key Management Service)** to ensure the security of the system.

### Architecture
![alt text](pictures/Architecture.png)

## Expected Outcome

By the end of this project, we expect to have a cloud-enabled, functional distributed MapReduce system capable of running on both a single machine and a multi-node environment with AWS EC2. The implementation will demonstrate how cloud computing enables the distribution of data processing across multiple nodes, efficiently handle failure scenarios, and showcase the **elastic scalability** provided by cloud infrastructure.

This project will help us gain a deeper understanding of cloud computing concepts such as **distributed storage**, **auto-scaling**, and **resource monitoring**. The integration with AWS services like **EC2**, **S3**, **Auto Scaling**, and **CloudWatch** will highlight the power and flexibility that cloud computing offers for large-scale data processing.

---

**References**

1. Dean, J., & Ghemawat, S. (2004). MapReduce: Simplified Data Processing on Large Clusters. *OSDI '04: 6th Symposium on Operating Systems Design and Implementation*, USENIX Association.
2. *MIT Graduate Course 6.5840: Distributed Systems (Spring 2024) - Lab: MapReduce*. Retrieved from [https://pdos.csail.mit.edu/6.824/labs/lab-mr.html](https://pdos.csail.mit.edu/6.824/labs/lab-mr.html).



