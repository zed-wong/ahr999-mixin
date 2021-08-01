# Ahr999-mixin
English | [中文](README-ZH.md)

A mixin bot that offers an up-to-date ahr999 index for the subscribed users, and a web page that provides historical data chart of the index.

## Living example
 website: [http://ahr999mixin.tk](http://ahr999mixin.tk)

 mixin bot: 7000103262


## Quick start

### mixin bot
  1. Register a bot at [https://developers.mixin.one/](https://developers.mixin.one/)

  2. Generate new client secrets

  3. Run `git clone github.com/who3m1/ahr999-mixin`
  
  4. Fill main.go with your credential
  ```
        ClientID   = ""        
        SessionID  = ""
        PrivateKey = ""
        PinToken   = ""
        Pin        = ""
  ```
  5. Run `go run main.go`, then access your bot in mixin messenger.

### website
  1. Rent a Ubuntu server (vultr.com or digitalocean.com)

  2. Install nginx on your server

  3. Run `git clone github.com/who3m1/ahr999-mixin` on your server.(not needed if you have done this before.)

  4. Copy index.html and main.py to /var/www/html/

  5. Run `python3 main.py` in /var/www/html

  6. Access your website through your server's ip, then the chart should be there.
  
  7. Run `crontab -e`, append `0 0 * * * python3 /var/www/html/main.py` to the end of file to make it executed everyday.

## File explanation

 - main.go 
   - Execute every 24 hours to notify the index.
   - Notify with mixin when the index hit the line.
   - Handles the bots message module, writes subbed userid to database.
 - main.py 
   - Execute every day to keep data.json up to date.
 - index.html
   - The webpage of ahr999 charts.
 - data.json 
   - A JSON file that contains timestamp and ahr999 index.

 - data.db 
   - Stores subscribed users.
