# Ahr999-mixin
[English](README-en.md) | 中文

一个提供实时更新的Ahr999指数的mixin机器人，与一个提供历史ahr999指数的web页面。

## 实例
 网站: [https://zed-wong.github.io/ahr999-mixin/](https://zed-wong.github.io/ahr999-mixin/)
 
 mixin机器人: 7000103262
 

## 计算ahr999指数

[main.go - getahr999()](https://github.com/zed-wong/ahr999-mixin/blob/main/main.go#L349)

[main.py](https://github.com/zed-wong/ahr999-mixin/blob/main/main.py#L40)

## 快速开始

### mixin机器人
  1. 在mixin开发者界面注册一个机器人 [https://developers.mixin.one/](https://developers.mixin.one/)
 
  2. 生成新的client secrets，并保存好。

  3. 在终端执行 `git clone github.com/zed-wong/ahr999-mixin`
  
  4. 将刚生成的client secrets填入main.go 
  ```
        ClientID   = ""        
        SessionID  = ""
        PrivateKey = ""
        PinToken   = ""
        Pin        = ""
  ```
  5. 执行`go run main.go`，然后在mixin messenger里访问你的机器人。


## 文件

 - main.go 
   - 每24小时播报一次ahr999指数
   - 处理机器人的消息模块，写入用户信息到数据库。

 - main.py 
   - 用于更新ahr999指数历史数据

 - index.html
   - ahr999指数的网页

 - data.json 
   - 包含timestamp和ahr999指数的json文件

## 其他指标

- 恐慌指数 https://alternative.me/crypto/fear-and-greed-index/
- 彩虹图 https://www.blockchaincenter.net/en/bitcoin-rainbow-chart/
- S2F https://studio.glassnode.com/metrics?a=BTC&category=Market%20Indicators&chartStyle=line&m=indicators.StockToFlowRatio&mAvg=7&mMedian=0&mScl=log&pScl=log&zoom=all
- Pi Cycle https://www.lookintobitcoin.com/charts/pi-cycle-top-indicator/
- MVRV Z https://www.lookintobitcoin.com/charts/mvrv-zscore/
- Bitcoin Price Models https://charts.woobull.com/bitcoin-price-models/
- Mayer index: https://stats.buybitcoinworldwide.com/mayermultiple/
- 欢迎提交PR添加更多指标
