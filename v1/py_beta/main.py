import os
import json
import pyotp
import time
from concurrent.futures import ThreadPoolExecutor

from client import Client







client = Client()

def init_client():
    current_dir = os.path.dirname(os.path.abspath(__file__))
    with open(os.path.join(current_dir, "config.json"), encoding="utf-8") as file:
        data = json.load(file)

    username = data["username"] if "username" in data else None
    password = data["password"] if "password" in data else None
    secret_key = data["secret_key"] if "secret_key" in data else None
    if not username or not password or not secret_key:
        raise Exception("请填写配置文件")
    
    response = client.un_read()
    if response.status_code == 200:
        print("use cookies")
        client.save_cookies()
        return
    
    token = client.oauth_login(username, password)["token"]
    client.sso_login(token)
    response = client.un_read()

    if response.json()["success"]:
        print("use password")
        client.save_cookies()
        return
    if response.json()["message"] == "请进行令牌验证":
        token = pyotp.TOTP(secret_key).now()
        client.login_by_token(token)
        response = client.un_read()
        if response.json()["success"]:
            print("use token")
            client.save_cookies()
            return
    raise Exception("登录失败")

def _get_posts(page):
    print(f"page: {page}")
    # 更新API调用以使用新的API端点
    response = client.get_posts_list(page=page, limit=100)
    data = response.json()
    
    if data["code"] != 20000:
        raise Exception(f"API错误: {data['message']}")
    
    posts_data = data["data"]["list"]
    posts = []
    
    for item in posts_data:
        # 映射新API字段到数据结构
        post = {
            "pid": item["pid"],
            "text": item["text"],
            "anonymous": item["anonymous"],
            "type": item.get("type", "text"),
            "image_size_x": 0,  # 新API中没有此字段，设为默认值
            "image_size_y": 0,  # 新API中没有此字段，设为默认值
            "extra": item.get("extra", 0),
            "timestamp": item["timestamp"],
            "reply": item["reply"],
            "likenum": item["likenum"],
            "tag": item.get("tag", ""),
            "status": item.get("status", 0),
            "is_comment": item.get("is_comment", 0),
            "is_protect": item.get("is_protect", 0),
            "is_top": item.get("is_top", 0),
            "label": item.get("label", 0)
        }
        
        posts.append(post)
    
    return posts


def _get_comments(pid):
    """获取指定帖子的评论"""
    response = client.get_comments_by_pid(pid)
    data = response.json()
    
    if data["code"] != 20000:
        raise Exception(f"API错误: {data['message']}")
    
    comments_data = data["data"]["list"]
    comments = []
    
    for item in comments_data:
        comment = {
            "cid": item["cid"],
            "pid": item["pid"],
            "name": item.get("name_tag", ""),
            "text": item["text"],
            "timestamp": item["timestamp"],
            "tag": item.get("tag", None),
            "quote": item.get("comment_id", None)  # 如果comment_id为null，则quote也为null
        }
        
        comments.append(comment)
    
    return comments


def update_posts():
    http_time = 0
    sql_time = 0
    
    for page in range(1, 2):
        t1 = time.time()
        posts = _get_posts(page)
        t2 = time.time()
        http_time += t2 - t1
        t3 = time.time()
        # 替换为打印或保存到文件等非数据库操作
        print(f"获取到 {len(posts)} 条帖子数据")
        sql_time += time.time() - t3

    print(f"http time: {http_time}")
    print(f"sql time: {sql_time}")

def batch_update_posts(batch_size):
    executor = ThreadPoolExecutor(max_workers=batch_size)
    futures = [executor.submit(lambda page=page: _get_posts(page)) for page in range(1, 100)]
    http_time = 0
    sql_time = 0
    
    for future in futures:
        t1 = time.time()
        posts = future.result()
        t2 = time.time()
        http_time += t2 - t1
        t3 = time.time()
        # 替换为打印或保存到文件等非数据库操作
        print(f"获取到 {len(posts)} 条帖子数据")
        sql_time += time.time() - t3

    print(f"http time: {http_time}")
    print(f"sql time: {sql_time}")

def update_comments():
    """更新评论数据"""
    # 由于不再使用数据库，这里需要从其他方式获取帖子ID
    # 临时方案：使用固定帖子ID进行演示
    sample_pids = [8123825]  # 从示例数据中获取的帖子ID
    http_time = 0
    sql_time = 0
    
    for pid in sample_pids:
        t1 = time.time()
        try:
            comments = _get_comments(pid)
            t2 = time.time()
            http_time += t2 - t1
            t3 = time.time()
            # 替换为打印或保存到文件等非数据库操作
            print(f"获取到 {len(comments)} 条评论数据")
            sql_time += time.time() - t3
        except Exception as e:
            print(f"获取帖子 {pid} 的评论时出错: {str(e)}")
            continue

    print(f"http time: {http_time}")
    print(f"sql time: {sql_time}")


if __name__ == "__main__":
    init_client()
    #update_posts()
    batch_update_posts(40)
    update_comments()  # 添加评论更新
