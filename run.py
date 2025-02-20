import sys
import time
import datetime
from getpass import getpass
import json
import re


from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.firefox.options import Options as FirefoxOptions
from selenium.webdriver.firefox.service import Service
from selenium.webdriver.firefox.firefox_profile import FirefoxProfile
from selenium.webdriver.edge.options import Options as EdgeOptions
from selenium.webdriver.edge.service import Service

from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.support.ui import WebDriverWait
import time
from post import Post, Reply
from config import WebConfig

def save_html(driver, file_name='index.html'):
    html= driver.page_source
    with open(file_name, 'w', encoding='utf-8') as file:
        file.write(html)

def convert_posts_to_json(posts, file_name='output.json'):
    print('Saving into json...')
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
            ]
        })
    json.dump(output, open(file_name, 'w', encoding='utf-8'), ensure_ascii=False, indent=2)

    
def extract_post(post_tree, crawled_pids):
    try:
        pquote = post_tree.find_element(By.XPATH, ".//div[@class='flow-item  flow-item-quote']/div[@class='box']/div[@class='box-header']//code[@class='box-id --box-id-copy-content']").get_attribute('textContent').strip()
    except:
        pquote = None
    try:
        pid = post_tree.find_element(By.XPATH, ".//div[@class='flow-item']//code[@class='box-id --box-id-copy-content']").get_attribute('textContent').strip()

        if pid in crawled_pids:
            return None
        else:
            crawled_pids.add(pid)
        try:
            plikenum = post_tree.find_element(By.XPATH, ".//div[@class='flow-item']//span[@class='box-header-badge likenum']").get_attribute('textContent').strip()
        except:
            plikenum = 0
        try:
            pbadge = post_tree.find_element(By.XPATH, ".//div[@class='flow-item']//span[@class='box-header-badge']").get_attribute('textContent').strip()
        except:
            pbadge = 0
        pcontent = post_tree.find_element(By.XPATH, ".//div[@class='flow-item']//div[@class='box-content']").get_attribute('textContent')
        ptime = post_tree.find_element(By.XPATH, ".//div[@class='flow-item']//div[@class='box-header']").get_attribute('textContent').strip()
        ptime = re.search(r'\d{2}-\d{2}\s\d{2}:\d{2}', ptime).group()
        #ptime = datetime.datetime.strptime(ptime,'%Y-%m-%dT%H:%M:%S')
        
        new_post = Post(pid, plikenum, pbadge, pcontent, ptime, pquote)
    except Exception as e:
        print(e)
        return 'ERROR'
    post = post_tree.find_element(By.XPATH, ".//div[@class='flow-item']")
    post = post.find_element(By.XPATH, "..")
    for reply_tree in post.find_elements(By.XPATH, ".//div[@class='flow-reply box dialog-hole-reply']"):
        rid = reply_tree.find_element(By.XPATH, ".//code[@class='box-id']").get_attribute('textContent').strip()
        
        rtime_ = reply_tree.find_element(By.XPATH, "./div[@class='box-header']")
        rtime = rtime_.get_attribute('textContent').strip()
        #print(rid, rtime, len(rtime), type(rtime))
        if not rtime:
            html_code = driver.execute_script("return arguments[0].outerHTML;", rtime_)
            print(html_code)
        rtime = re.search(r'\d{2}-\d{2}\s\d{2}:\d{2}', rtime).group()
        #rtime = datetime.datetime.strptime(rtime,'%Y-%m-%dT%H:%M:%S')
        rbox = reply_tree.find_element(By.XPATH, "./div[@class='box-content']")
        try:
            rquote = rbox.find_element(By.XPATH, "./div[contains(@class, 'quote')]").get_attribute('textContent').strip()
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
            new_post.add_reply(rid, name, rcontent, rtime, (quote_name, rquote))
            new_post.add_reply(rid, name, rcontent, rtime, None)
        else:
            new_post.add_reply(rid, name, rcontent, rtime, None)

    return new_post

def get_posts(driver, crawled_pids):
    posts = []
    post_trees = driver.find_elements(By.XPATH, "//div[@class='flow-chunk']/div")
    #post_trees = driver.find_elements(By.XPATH, "//div[count(@*)=1 and @data-v-0582b940]")
    print(len(post_trees))
    for post_tree in post_trees:
        #print('a new post')
        #print(post_tree.get_attribute('textContent'))
        new_post = extract_post(post_tree, crawled_pids)
        if new_post != None:
            if new_post != 'ERROR':
                posts.append(new_post)
            else:
                html_code = driver.execute_script("return arguments[0].outerHTML;", post_tree)
                print(html_code)
                input('confirm...')
    for i in range(len(post_trees) - 3):
        #input('next pop:')
        #time.sleep(0.2)
        #print('pop')
        post_tree = post_trees[i]
        try:
            driver.execute_script("""
                        arguments[0].parentNode.removeChild(arguments[0]);
                    """, post_tree)
        except:
            pass
    return posts


if __name__ == '__main__':
    webconfig = WebConfig()
    browser = webconfig.browser
    profiles_path = webconfig.profiles_path
    crawl_size = webconfig.crawl_size
    if browser == 'Firefox':
        options = FirefoxOptions()
        firefox_profile = FirefoxProfile(profiles_path)
        options.profile = firefox_profile
        #options.add_argument("--headless")
        driver = webdriver.Firefox(options=options)
    elif browser == 'Edge':
        options = EdgeOptions()
        options.add_argument(rf'user-data-dir={profiles_path}')
        #options.add_argument("--headless=new")
        driver = webdriver.Edge(options=options)

    driver.get('https://treehole.pku.edu.cn')
    driver.set_window_size(1980, 100)
    current_url = driver.current_url
    if current_url.startswith('https://treehole.pku.edu.cn'):
        print('Log in successfully')
        try:
            time.sleep(3)
            save_html(driver)

            posts = []
            crawled_pids = set([])
            #print(crawl_size)
            i = 1
            total_length = 0
            while(total_length < crawl_size):
                #post_cnt = len(driver.find_elements(By.XPATH, "//div[count(@*)=1 and @data-v-0582b940]"))
                new_posts = get_posts(driver, crawled_pids)
                #print('done')
                #print(new_posts)
                posts += new_posts
                print(len(posts), len(new_posts))
                #print('scroll')
                if len(posts) >= 50:
                    convert_posts_to_json(posts[:50], file_name=f'tree_hole_50_{i}.json')
                    i += 1
                    total_length += 50
                    posts = posts[50:]
                driver.execute_script("arguments[0].scrollIntoView(true);", driver.find_elements(By.XPATH, "//div[count(@*)=1 and @data-v-0582b940]")[-1])
                #time.sleep(0.5)
            #posts = posts[:crawl_size]
            
            print('Crawling done')
            #output_json_name = input('output json name:')
            #convert_posts_to_json(posts, file_name=output_json_name)
            
        except Exception as err:
            print(err)

        input("Press any key to finish...")
        driver.close()

    else:
        print('Fail to log')
        input('Press any key to quit...')
        driver.close()
