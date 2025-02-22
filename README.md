# PKUHoleCrawler-new
（一个简易的）北大树洞爬虫，基于Selenium动态爬取网页内容。

基于[luciusssss/PKUHoleCrawler: 北大树洞爬虫 (github.com)](https://github.com/luciusssss/PKUHoleCrawler)的改进，适用于新版本树洞[北大树洞 (pku.edu.cn)](https://treehole.pku.edu.cn)。目前支持Edge浏览器与Firefox浏览器。

## 配置

安装selenium

```
pip3 install selenium
```

为了实现自动登录，需要拷贝浏览器的用户数据，Edge的用户数据默认位于`C:\Users\YourName\AppData\Local\Microsoft\Edge\User Data`(Windows)或`/home/YourName/.config/microsoft-edge/User Data`，Firefox的用户数据默认位于`C:\Users\YourName\AppData\Roaming\Mozilla\Firefox\Profiles\`(Windows)或`/home/你YourName/.mozilla/firefox/`(Linux)下的某个随机名称文件夹（例如`32fy5laa.default-release`），需要在原浏览器上保留登录状态（即访问https://treehole.pku.edu.cn不会跳转至登录界面）。

浏览器对应的webdriver按需安装。

## 运行

使用`config.py`修改运行参数：

```
config.py [-h] [--crawl_size CRAWL_SIZE] [--part PART] [--browser BROWSER] [--profiles_path PROFILES_PATH]
```

然后运行`run.py`：

```
python3 run.py
```

爬取文本结果将按照`part`条树洞一组存储在`tree_hole_{part}_{start}-{end}({utc-time}).json`中；

图片将存储于`download`文件夹下，格式为`image_{pid}.png`。

命令行界面：

```
Log in successfully
第1部分： |██████████| 500/500 100.0%   总进度： |██████████| 2000/2000 100.0% 
Crawling done
Press any key to finish...
```



## 示例

爬取的部分文件存放在`/data`下。

