import os
import requests

# 替换为 S3 存储桶的 URL 根路径
base_url = "https://lab1-mapreduce-bucket.s3.amazonaws.com/"
# 要下载的文件列表
file_list = [
    "pg-being_ernest.txt",
    "pg-dorian_gray.txt",
    "pg-frankenstein.txt",
    "pg-grimm.txt",
    "pg-huckleberry_finn.txt",
    "pg-metamorphosis.txt",
    "pg-sherlock_holmes.txt",
    "pg-tom_sawyer.txt"
]

# 本地保存目录
save_dir = "~/CSC4160_FinalProject/src/main"
save_dir = os.path.expanduser(save_dir)  # 转换为绝对路径
os.makedirs(save_dir, exist_ok=True)  # 如果目录不存在，则创建

# 下载文件
for file_name in file_list:
    file_url = base_url + file_name
    local_file_path = os.path.join(save_dir, file_name)

    print(f"Downloading {file_url} to {local_file_path}...")
    try:
        response = requests.get(file_url, stream=True)
        response.raise_for_status()  # 如果请求出错则抛出异常

        with open(local_file_path, "wb") as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)
        print(f"Downloaded {file_name} successfully!")
    except requests.exceptions.RequestException as e:
        print(f"Failed to download {file_name}: {e}")

print("All downloads complete!")
