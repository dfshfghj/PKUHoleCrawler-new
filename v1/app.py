from client import Client
import getpass
from concurrent.futures import ThreadPoolExecutor
import os
import datetime
import json

class App:
    def __init__(self):
        self.client = Client()
        self.executor = ThreadPoolExecutor(max_workers=20)
        self.current_dir = os.path.dirname(os.path.abspath(__file__))
        if not os.path.exists(os.path.join(self.current_dir, 'data', 'download')):
            os.makedirs(os.path.join(self.current_dir, 'data', 'download'))

        response = self.client.un_read()
        while response.status_code != 200:
            print(f"{response.status_code}: 需要登录")
            
            username = input('username: ')
            password = getpass.getpass('password: ')
            token = self.client.oauth_login(username, password)["token"]
            self.client.sso_login(token)
            response = self.client.un_read()

        while not response.json()["success"]:
            if response.json()["message"] == "请手机短信验证":
                tmp = input("发送验证码(Y/n)：")
                if tmp == 'Y':
                    self.client.send_message()
                    code = input("短信验证码：")
                    self.client.login_by_message(code)
            elif response.json()["message"] == "请进行令牌验证":
                token = input("手机令牌：")
                self.client.login_by_token(token)
            response = self.client.un_read()

    def read(self, post_id):
        post = self.client.get_post(post_id)
        if post["success"]:
            post = post["data"]

            reply = post["reply"]
            likenum = post["likenum"]
            text = post["text"]
            print(f"{post_id}  reply:{reply} likenum:{likenum}")
            print(text)
        else:
            print(f"{post_id}: {post["message"]}")

    def get_post(self, post_id):
        post = self.client.get_post(post_id)
        if post["success"]:
            post = post["data"]
            if post["type"] == "image":
                image_type = post["url"].split(".")[-1]
                self.client.get_image(post_id, os.path.join(self.current_dir, 'data', 'download', post_id) + "." + image_type)
            comments = self.client.get_comment(post_id)["data"]

            if comments:
                last_page = comments["last_page"]
                for page in range(2, last_page + 1):
                    part_comments = self.client.get_comment(post_id, page)["data"]
                    comments["data"] += part_comments["data"]
                comments = comments["data"]
            else:
                comments = []
            return post, comments
        else:
            return {'pid': post_id, 'text': '您查看的树洞不存在', 'type': 'text'}, []

    def get_posts(self, posts):
        posts_data = []
        futures = [self.executor.submit(lambda post_id=post_id: self.get_post(post_id)) for post_id in posts]
        for future in futures:
            post, comments = future.result()
            posts_data.append({"post": post, "comments": comments})
        data_name = os.path.join(self.current_dir, 'data', datetime.datetime.now().strftime('%Y-%m-%d-%H-%M-%S')) + ".json"
        with open(data_name, 'w') as file:
            json.dump(posts_data, file, indent=4)

if __name__ == "__main__":
    app = App()
    while True:
        post_id = input("post id: ")
        app.read(post_id)





