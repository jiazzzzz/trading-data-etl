# jia-stk

#### Objective

  This repo is for dump stock info from different site and do some analysis

#### Sites to get info from

```
Tushare: To get the stock list
Sina: To get the daily info
163: To get the history data
```

#### Programs

```
dump.py - To dump latest stock info and daily trading data
sg.py - To get today new stock (Will not be saved to database)
mon.py - To monitor stock live status
status.py - To get market status after information is dumped to database
```

#### DB

```
stock_basic - Basic info retrieved from tushare, updated daily
stock_daily_xxxx - Daily trading info for all stocks, everyday info is saved to one table, and will be kept 7 days

```



#### 