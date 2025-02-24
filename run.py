import sys
import os
import time
import datetime
import json
import re
import base64
import shutil
import psutil
import logging


from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.firefox.options import Options as FirefoxOptions
from selenium.webdriver.firefox.service import Service
from selenium.webdriver.firefox.firefox_profile import FirefoxProfile
from selenium.webdriver.edge.options import Options as EdgeOptions
from selenium.webdriver.edge.service import Service

from selenium.webdriver.common.by import By
from selenium.webdriver.common.action_chains import ActionChains
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait
import time
from post import Post, Reply
from config import WebConfig


def print_progress(iteration, total, prefix, suffix, decimals, length, fill='█'):
    terminal_size = shutil.get_terminal_size()
    max_length = terminal_size.columns
    out = '\r'
    p = True
    for i in range(len(iteration)):
        percent = (
            "{0:." + str(decimals[i]) + "f}").format(100 * (iteration[i] / float(total[i])))
        filled_length = int(length[i] * iteration[i] // total[i])
        bar = fill * filled_length + '-' * (length[i] - filled_length)
        out += f'{prefix[i]} |{bar}| {iteration[i]}/{total[i]} {percent}% {suffix[i]}'
        p = p and iteration[i] >= total[i]

    sys.stdout.write(out[:max_length-2])
    sys.stdout.flush()
    if p:
        print()


def scroll_element(driver, element):
    element_location = element.location_once_scrolled_into_view
    element_height = element.size['height']
    scroll_y = element_location['y'] + element_height
    driver.execute_script("""
        window.scrollTo({
                          top: arguments[0],
                          behavior: 'smooth'
                          });
    """, scroll_y)


def save_html(driver, file_name='index.html'):
    html = driver.page_source
    with open(file_name, 'w', encoding='utf-8') as file:
        file.write(html)


def convert_posts_to_json(posts, file_name='output.json'):
    # print('Saving into json...')
    output = []
    for post in posts:
        output.append({
            'id': post.id,
            'likenum': post.likenum,
            'badge': post.badge,
            'content': post.content,
            'time': str(post.time),
            'quote': post.quote,
            'replies': [
                {
                    'id': reply.id,
                    'name': reply.name,
                    'content': reply.content,
                    'time': str(reply.time),
                    'quote': reply.quote
                }
                for reply in post.replies
            ],
            'tip': post.tip
        })
    current_dir = os.path.dirname(os.path.abspath(__file__))
    if not os.path.exists(os.path.join(current_dir, 'data')):
        os.makedirs(os.path.join(current_dir, 'data'))
    file_path = os.path.join(current_dir, 'data', file_name)
    json.dump(output, open(file_path, 'w', encoding='utf-8'),
              ensure_ascii=False, indent=2)


def get_image(box_content, image_name):
    img_element = box_content.find_element(
        By.XPATH, ".//p[@class='img']/a/img[starts-with(@src, 'blob:')]")
    result = driver.execute_async_script("""
        var img = arguments[0];
        var callback = arguments[1];
        var xhr = new XMLHttpRequest();
        xhr.open('GET', img.src, true);
        xhr.responseType = 'blob';
        xhr.onload = function(e) {
            if (this.status == 200) {
                var reader = new FileReader();
                reader.onloadend = function() {
                    callback(reader.result);
                }
                reader.readAsDataURL(this.response);
            } else {
                callback(null);
            }
        };
        xhr.send();
    """, img_element)
    if result and 'data:image' in result:
        current_dir = os.path.dirname(os.path.abspath(__file__))
        if not os.path.exists(os.path.join(current_dir, 'data', 'download')):
            os.makedirs(os.path.join(current_dir, 'data', 'download'))
        image_path = os.path.join(current_dir, 'data', 'download', image_name)
        image_data = result.split(',')[1]
        with open(image_path, 'wb') as f:
            f.write(base64.b64decode(image_data))
        # print("download a image")
    else:
        pass
        # print("cannot download")


def extract_post(post_tree, crawled_pids):
    try:
        pid = post_tree.find_element(
            By.XPATH, ".//div[@class='flow-item']//code[@class='box-id --box-id-copy-content']").get_attribute('textContent').strip()

        if pid in crawled_pids:
            return None
        else:
            crawled_pids.add(pid)
        try:
            pquote = post_tree.find_element(
                By.XPATH, ".//div[@class='flow-item  flow-item-quote']/div[@class='box']/div[@class='box-header']//code[@class='box-id --box-id-copy-content']").get_attribute('textContent').strip()
        except:
            pquote = None
        try:
            plikenum = post_tree.find_element(
                By.XPATH, ".//div[@class='flow-item']//span[@class='box-header-badge likenum']").get_attribute('textContent').strip()
        except:
            plikenum = 0
        try:
            pbadge = post_tree.find_element(
                By.XPATH, ".//div[@class='flow-item']//span[@class='box-header-badge']").get_attribute('textContent').strip()
        except:
            pbadge = 0
        pcontent_fold_body = post_tree.find_element(
            By.XPATH, ".//div[@class='flow-item']//div[@class='box-content']//div[@class='content-fold-body']")
        pcontent = pcontent_fold_body.get_attribute('textContent')
        try:
            get_image(pcontent_fold_body, f'image_{pid}.png')
            pimage = True
        except:
            pimage = False
        ptime = post_tree.find_element(
            By.XPATH, ".//div[@class='flow-item']//div[@class='box-header']").get_attribute('textContent').strip()
        ptime = re.search(r'\d{2}-\d{2}\s\d{2}:\d{2}', ptime).group()

        new_post = Post(pid, plikenum, pbadge, pcontent, ptime, pquote, pimage)
    except Exception as e:
        html_code = driver.execute_script(
            "return arguments[0].outerHTML;", post_tree)
        logger.error(f'cannot extract post: {e} \n html code:\n{html_code}')
        # print(e)
        return None

    post = post_tree.find_element(By.XPATH, ".//div[@class='flow-item']")
    post = post.find_element(By.XPATH, "..")
    try:
        pbox_tip = post.find_element(
            By.XPATH, ".//div[@class=box box-tip]").text
        new_post.tip = pbox_tip
    except:
        new_post.tip = None
    for reply_tree in post.find_elements(By.XPATH, ".//div[@class='flow-reply box dialog-hole-reply']"):
        rid = reply_tree.find_element(
            By.XPATH, ".//code[@class='box-id']").get_attribute('textContent').strip()

        rtime_ = reply_tree.find_element(
            By.XPATH, "./div[@class='box-header']")
        rtime = rtime_.get_attribute('textContent').strip()
        if not rtime:
            html_code = driver.execute_script(
                "return arguments[0].outerHTML;", rtime_)
            print(html_code)
        rtime = re.search(r'\d{2}-\d{2}\s\d{2}:\d{2}', rtime).group()
        # rtime = datetime.datetime.strptime(rtime,'%Y-%m-%dT%H:%M:%S')
        rbox = reply_tree.find_element(By.XPATH, "./div[@class='box-content']")
        try:
            rquote = rbox.find_element(
                By.XPATH, "./div[contains(@class, 'quote')]").get_attribute('textContent').strip()
        except:
            rquote = None
        rcontents = rbox.find_elements(By.XPATH, "./span")
        name = rcontents[1].get_attribute('textContent')
        if rquote:
            quote_name = rcontents[-2].get_attribute('textContent')
        else:
            quote_name = None
        rcontent = rcontents[-1].get_attribute('textContent')[2:]
        if rquote:
            new_post.add_reply(rid, name, rcontent, rtime,
                               (quote_name, rquote))
        else:
            new_post.add_reply(rid, name, rcontent, rtime, None)

    return new_post


def get_posts(driver, crawled_pids):
    posts = []
    post_trees = driver.find_elements(
        By.XPATH, "//div[@class='flow-chunk']/div")
    '''
    driver.execute_script("""
                          arguments[0].scrollIntoView({
                          behavior: 'smooth',
                          block: 'start'
                          });
                          """, post_trees[-1])
    '''
    for post_tree in post_trees:
        # print('a new post')
        new_post = extract_post(post_tree, crawled_pids)
        if new_post != None:
            posts.append(new_post)
    scroll_element(driver, post_trees[-1])
    for i in range(len(post_trees) - 3):
        post_tree = post_trees[i]
        try:
            driver.execute_script("""
                        arguments[0].parentNode.removeChild(arguments[0]);
                    """, post_tree)
        except:
            logger.warning('cannot delete node')

    return posts


if __name__ == '__main__':
    webconfig = WebConfig()
    logger = logging.getLogger()
    logger.setLevel(logging.INFO)
    log_name = f'{datetime.datetime.now(datetime.UTC).strftime("UTC%Y-%m-%d %H%M%S")}.log'
    file_handler = logging.FileHandler(log_name, encoding='utf-8')
    formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s')
    file_handler.setFormatter(formatter)
    logger.addHandler(file_handler)
    mode = webconfig.mode
    browser = webconfig.browser
    profiles_path = webconfig.profiles_path
    crawl_size = webconfig.crawl_size
    part = webconfig.part

    logger.info(
        f'browser={browser}, profiles_path={profiles_path}, crawl_size={crawl_size}, part={part}')

    if browser == 'Firefox':
        options = FirefoxOptions()
        firefox_profile = FirefoxProfile(profiles_path)
        options.profile = firefox_profile
        # options.add_argument("--headless")
        driver = webdriver.Firefox(options=options)
    elif browser == 'Edge':
        options = EdgeOptions()
        options.add_argument(rf'user-data-dir={profiles_path}')
        # options.add_argument("--headless=new")
        driver = webdriver.Edge(options=options)
    driver.execute_script("""
        Object.defineProperty(Navigator, 'webdriver', {get: () => undefined});
        Object.defineProperty(navigator, 'webdriver', {get: () => false});               
    """)
    driver.get('https://treehole.pku.edu.cn')
    time.sleep(1)
    current_url = driver.current_url
    if current_url == 'https://treehole.pku.edu.cn/web/verification':
        content = driver.find_element(By.TAG_NAME, 'body').text
        if '短信' in content:
            logger.error('need message verification')
        elif '手机令牌' in content:
            logger.error('need mobile token')
        else:
            logger.error('unknown error')
        print('Fail to log')
        driver.close()
    elif current_url.startswith('https://treehole.pku.edu.cn'):
        logger.info('Log in successfully')
        print('Log in successfully')
        time.sleep(5)
        # save_html(driver)

        posts = []
        crawled_pids = set([])
        i = 1
        total_length = 0
        memory_warn = False
        timeout_warn = False
        while (total_length < crawl_size):
            memory = psutil.virtual_memory()
            if not memory_warn and memory.available <= 524288000:
                memory_warn = True
                logger.warning(
                    f'lack of available memory: {memory.available / (1024**2):.2f} MB')
            elif memory.available <= 262144000:
                logger.error(
                    f'serious lack of available memory: {memory.available / (1024**2):.2f} MB')
                break
            new_posts = get_posts(driver, crawled_pids)
            if not timeout_warn:
                if len(new_posts) == 0:
                    t_start = time.time()
                    timeout_warn = True
            else:
                if len(new_posts) == 0:
                    t_end = time.time()
                    if t_end - t_start >= 20:
                        logger.error(f'time out: {(t_end - t_start):.2f}s')
                        break
                else:
                    timeout_warn = False

            posts += new_posts
            if len(posts) >= part:
                start = posts[0].id
                end = posts[part - 1].id
                now = datetime.datetime.now(datetime.UTC)
                now = now.strftime("UTC%Y-%m-%d %H%M%S")
                json_name = f'tree_hole_{part}_{start}-{end}({now}).json'
                convert_posts_to_json(
                    posts[:part], file_name=json_name)
                total_length += part
                logger.info(f'part {i} done: {total_length}/{crawl_size}')
                i += 1
                posts = posts[part:]
            print_progress((len(posts), total_length), (part, crawl_size),
                           (f'第{i}部分：', '总进度：'), ('  ', ''), (1, 1), (10, 10))
            time.sleep(0.1)
        logger.info('crawling done')
        print('Crawling done')
        input("Press any key to finish...")
        driver.close()

    else:
        logger.warning('fail to log, try again')
        print('Fail to log')
        # input('Press any key to quit...')
        driver.close()
