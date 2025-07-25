import requests
import enum
import random
import re
import os
import json
import uuid
from http.cookiejar import Cookie

class TreeHoleWeb(enum.Enum):
    OAUTH_LOGIN = "https://iaaa.pku.edu.cn/iaaa/oauthlogin.do"
    REDIR_URL = "https://treehole.pku.edu.cn/cas_iaaa_login?uuid=fc71db5799cf&plat=web"
    SSO_LOGIN = "http://treehole.pku.edu.cn/cas_iaaa_login"
    UN_READ = "https://treehole.pku.edu.cn/api/mail/un_read"
    SEARCH = "https://treehole.pku.edu.cn/api/pku_hole"
    COMMENT = "https://treehole.pku.edu.cn/api/pku_comment_v3"
    FOLLOW = "https://treehole.pku.edu.cn/api/pku_attention"
    GET_FOLLOW = "https://treehole.pku.edu.cn/api/follow_v2"
    REPORT = "https://treehole.pku.edu.cn/api/pku_comment/report"
    LOGIN_BY_TOKEN = "https://treehole.pku.edu.cn/api/login_iaaa_check_token"
    LOGIN_BY_MESSAGE = "https://treehole.pku.edu.cn/api/jwt_msg_verify"
    SEND_MESSAGE = "https://treehole.pku.edu.cn/api/jwt_send_msg"
    COURSE_TABLE = "https://treehole.pku.edu.cn/api/getCoursetable_v2"
    GRADE = "https://treehole.pku.edu.cn/api/course/score_v2"


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
            'uuid': str(uuid.uuid4()).split("-")[-1],
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
        response = self.session.post(TreeHoleWeb.LOGIN_BY_TOKEN.value, data={'code': token})
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
    
    def get_comment(self, post_id, page=1, limit=15, sort="asc"):
        response = self.session.get(f"https://treehole.pku.edu.cn/api/pku_comment_v3/{post_id}", params={
            "page": page,
            "limit": limit,
            "sort": sort
        })
        response.raise_for_status()
        return response.json()
    
    def get_image(self, post_id, file_name):
        response = self.session.get(f"https://treehole.pku.edu.cn/api/pku_image/{post_id}", stream=True)
        if response.status_code == 200:
            with open(f"{file_name}", "wb") as file:
                for chunk in response.iter_content(1024):
                    file.write(chunk)

    def search(self, keyword=None, page=1, limit=25, label=None):
        response = self.session.get(TreeHoleWeb.SEARCH.value, params={
            "page": page,
            "limit": limit,
            "keyword": keyword,
            "label": label
        })
        return response
    
    def follow(self, post_id):
        response = self.session.post(TreeHoleWeb.FOLLOW.value + f"/{post_id}")
        return response
    
    def get_follow(self, page=1, limit=25):
        response = self.session.get(TreeHoleWeb.GET_FOLLOW.value, params={
            "page": page,
            "limit": limit
        })
        return response
    
    def comment(self, post_id, text, comment_id=None):
        response = self.session.post(TreeHoleWeb.COMMENT.value, data={
            "comment_id": comment_id,
            "pid": post_id,
            "text": text
        } if comment_id else {
            "pid": post_id,
            "text": text
        })
        return response
    
    def report(self, tp, xid, other, reason):
        if tp == 'post':
            post_id = xid
            response = self.session.post(TreeHoleWeb.REPORT.value + f"/{post_id}", data={
                "other": other,
                "reason": reason
            })
        elif tp == 'comment':
            comment_id = xid
            response = self.session.post(TreeHoleWeb.REPORT.value, data={
                "cid": comment_id,
                "other": other,
                "reason": reason
            })
        return response
    
    def get_course_table(self):
        response = self.session.get(TreeHoleWeb.COURSE_TABLE.value)
        return response
    
    def get_grade(self):
        response = self.session.get(TreeHoleWeb.GRADE.value)
        return response
    
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

        current_path = os.path.abspath(__file__)
        cookie_path = os.path.join(os.path.dirname(current_path), "cookies.json")
        with open(cookie_path, 'w') as f:
            json.dump(cookies_list, f, indent=4)

    def load_cookies(self):
        current_path = os.path.abspath(__file__)
        cookie_path = os.path.join(os.path.dirname(current_path), "cookies.json")
        try:
            with open(cookie_path, 'r') as f:
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
