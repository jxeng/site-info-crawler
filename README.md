# site-info-crawler
A tool for batch crawling website's title, description, favicon.

# how to use

```
git clone https://github.com/jxeng/site-info-crawler.git

cd site-info-crawler

go mod tidy

go build

./site-info-crawler.exe
```

raw.json
```
[
  {
    "id": "binance",
    "url": "https://www.binance.com/",
    "title": "",
    "describe": "",
    "favicon": "binance_com.png"
  },
  ...
]
```

filled.json
```
[
  {
    "id": "binance",
    "url": "https://www.binance.com/",
    "title": "交易比特币、以太币和altcoin | 加密货币交易平台 | 币安",
    "description": "Binance cryptocurrency exchange - We operate the worlds biggest bitcoin exchange and altcoin crypto exchange in the world by volume",
    "favicon": "./icons/binance.png"
  },
  ...
]
```