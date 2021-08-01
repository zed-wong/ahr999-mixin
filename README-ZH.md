# Ahr999-mixin
[English](README.md) | 中文

一个提供实时更新的Ahr999指数的mixin机器人，与一个提供历史ahr999指数的web页面。

## 实例
 网站: [https://ahr999mixin.tk]()
 
 mixin机器人: 7000103262


## 快速开始

### mixin机器人
  1. 在mixin开发者界面注册一个机器人 [https://developers.mixin.one/]()
 
  2. 生成新的client secrets，并保存好。

  3. 在终端执行 `git clone github.com/who3m1/ahr999-mixin`
  
  4. 将刚生成的client secrets填入main.go 
  ```
        ClientID   = ""        
        SessionID  = ""
        PrivateKey = ""
        PinToken   = ""
        Pin        = ""
  ```
  5. 执行`go run main.go`，然后在mixin messenger里访问你的机器人。

### 网站
  1. 租用一个云服务器(vultr.com 或 digitalocean.com)

  2. 在服务器上安装ubuntu 

  3. 在服务器上执行 `git clone github.com/who3m1/ahr999-mixin`.(如果在上一步已经执行则不需要。)

  4. 复制 index.html 和 main.py 到/var/www/html/

  5. 在 /var/www/html 执行 `python3 main.py` 

  6. 通过服务器的ip来访问你的网站, 图表会被展示出来。
  
  7. 执行`crontab -e`, 在末尾添加`0 0 * * * python3 /var/www/html/main.py`，这样它会被每天执行。

## 文件解释

 - main.go 
   - 每24小时播报一次ahr999指数
   - 当指数触底时通知用户
   - 处理机器人的消息模块，写入用户信息到数据库。

 - main.py 
   - 用于更新ahr999指数历史数据

 - index.html
   - ahr999指数的网页

 - data.json 
   - 包含timestamp和ahr999指数的json文件

 - data.db 
   - 存储已订阅的用户
