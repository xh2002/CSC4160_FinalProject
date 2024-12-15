# CSC4160 Final Project
Final Project of CSC4160 at CUHK(SZ): Cloud-Based Distributed MapReduce System

## **Introduction**  
This project implements a cloud-based distributed MapReduce system inspired by Google’s original MapReduce framework. It demonstrates scalability, fault tolerance, and integration with cloud services for large-scale data processing.


## **Setup Instructions**  

### **1. Prerequisites**  
- Go (1.20 or later) installed on your system.  
- Access to a cloud environment such as AWS EC2 for multi-node execution (optional).  
- Source code cloned from the [Github Repository](https://github.com/xh2002/CSC4160_FinalProject).  

### **2. Build locally and Run WordCount**  

1. **Navigate to the `src/main` directory**:  
   ```bash  
   cd src/main  
   ```  

2. **Build the WordCount plugin**:  
   ```bash  
   go build -buildmode=plugin ../mrapps/wc.go  
   ```  

3. **Clean up previous outputs** (if any):  
   ```bash  
   rm mr-out*  
   ```  

4. **Run the WordCount task in sequential mode**:  
   ```bash  
   go run mrsequential.go wc.so pg*.txt  
   ```  

5. **View the output**:  
   ```bash  
   more mr-out-0  
   ```  
