import requests
import enum
import random
import re
import json
import getpass
from http.cookiejar import Cookie

class TreeHoleWeb(enum.Enum):
    OAUTH_LOGIN = "https://iaaa.pku.edu.cn/iaaa/oauthlogin.do"
    REDIR_URL = "https://treehole.pku.edu.cn/cas_iaaa_login?uuid=fc71db5799cf&plat=web"
    SSO_LOGIN = "http://treehole.pku.edu.cn/cas_iaaa_login"
    UN_READ = "https://treehole.pku.edu.cn/api/mail/un_read"
    LOGIN_BY_TOKEN = "https://treehole.pku.edu.cn/api/login_iaaa_check_token"
    LOGIN_BY_MESSAGE = "https://treehole.pku.edu.cn/api/jwt_msg_verify"
    SEND_MESSAGE = "https://treehole.pku.edu.cn/api/jwt_send_msg"


class Client:
    def __init__(self):
        self.session = requests.Session()
        self.session.headers.update({
            "user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0"
        })
        self.load_cookies()
        if "pku_token" in self.session.cookies.keys():
            self.authorization = self.session.cookies.values()[self.session.cookies.keys().index("pku_token")]
            self.session.headers.update({"authorization": f"Bearer {self.authorization}"})

    def oauth_login(self, username, password):
        response = self.session.post(TreeHoleWeb.OAUTH_LOGIN.value, data={
            'appid': "PKU Helper",
            'userName': username,
            'password': password,
            'randCode': '',
            'smsCode': '',
            'otpCode': '',
            'redirUrl': TreeHoleWeb.REDIR_URL.value
        })
        response.raise_for_status()
        return response.json()
    
    def sso_login(self, token):
        rand = str(random.random())
        response = self.session.get(TreeHoleWeb.SSO_LOGIN.value, params={
            'uuid': "fc71db5799cf",
            'plat': "web",
            '_rand': rand,
            'token': token
            })
        response.raise_for_status()
        print(response.status_code, response.headers)

        self.authorization = re.search(r'token=(.*)', response.url).group(1)
        self.session.cookies.update({"pku_token": self.authorization})
        self.session.headers.update({"authorization": f"Bearer {self.authorization}"})
        return response
    
    def un_read(self):
        response = self.session.get(TreeHoleWeb.UN_READ.value)

        return response
    
    def login_by_token(self, token):
        response = self.session.post(TreeHoleWeb.LOGIN_BY_TOKEN.value, data={'token': token})
        response.raise_for_status()
        print(response.status_code, response.json())
        return response
    
    def login_by_message(self, code):
        response = self.session.post(TreeHoleWeb.LOGIN_BY_MESSAGE.value, data={'valid_code': code})
        response.raise_for_status()
        print(response.status_code, response.json())
        return response
    
    def send_message(self):
        response = self.session.post(TreeHoleWeb.SEND_MESSAGE.value)
        response.raise_for_status()
        return response
    
    def get_post(self, post_id):
        response = self.session.get(f"https://treehole.pku.edu.cn/api/pku/{post_id}")
        response.raise_for_status()
        return response.json()
    
    def get_comment(self, post_id, page, limit, sort="asc"):
        response = self.session.get(f"https://treehole.pku.edu.cn/api/pku_comment_v3/{post_id}", params={
            "page": page,
            "limit": limit,
            "sort": sort
        })
        response.raise_for_status()
        return response.json()

    def save_cookies(self):
        cookies_list = []
        for cookie in self.session.cookies:
                cookie_dict = {
                    'name': cookie.name,
                    'value': cookie.value,
                    'domain': cookie.domain,
                    'path': cookie.path,
                    'expires': cookie.expires if cookie.expires else None,
                    'secure': cookie.secure,
                    'rest': {'HttpOnly': cookie.has_nonstandard_attr('HttpOnly')}
                }
                cookies_list.append(cookie_dict)

        with open("cookies.json", 'w') as f:
            json.dump(cookies_list, f, indent=4)

    def load_cookies(self):
        try:
            with open("cookies.json", 'r') as f:
                cookies_list = json.load(f)
            self.session.cookies.clear()
            for cookie_dict in cookies_list:
                cookie = Cookie(
                    version=0,
                    name=cookie_dict['name'],
                    value=cookie_dict['value'],
                    port=None,
                    port_specified=False,
                    domain=cookie_dict['domain'],
                    domain_specified=bool(cookie_dict['domain']),
                    domain_initial_dot=cookie_dict['domain'].startswith('.'),
                    path=cookie_dict['path'],
                    path_specified=bool(cookie_dict['path']),
                    secure=cookie_dict['secure'],
                    expires=cookie_dict['expires'],
                    discard=False,
                    comment=None,
                    comment_url=None,
                    rest=cookie_dict['rest']
                )
                self.session.cookies.set_cookie(cookie)

        except Exception as e:
            print(e)
    



    
if __name__ == "__main__":
    client = Client()
    response = client.un_read()
    if response.status_code != 200:
        print(f"{response.status_code}: 需要登录")
        
        username = input('username: ')
        password = getpass.getpass('password: ')
        data = client.oauth_login(username, password)
        token = data["token"]
        print(token)
        client.sso_login(token)
        response = client.un_read()
        
    else:
        if response.json()["message"] == "请手机短信验证":
            client.send_message()
            code = input("短信验证码：")
            client.login_by_message(code)
        elif response.json()["message"] == "请进行令牌验证":
            token = input("手机令牌：")
            client.login_by_token(token)

    while True:
        key = input("key (q to quit): ")
        if key == 'q':
            client.save_cookies()
            break
        else:
            print(client.get_post(key))
