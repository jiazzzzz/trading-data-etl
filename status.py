#This program is to get the market status, maybe can be ran after 15:00 every day
import sys
sys.path.append('./lib')
from lib.stock_info import StockInfo
from lib.db import Db
from lib.common import Common
import pandas as pd
import datetime
import time

#Dump stock list from tushare to database
def get_market_status(db,table_name):
    top_high_sql = "select name,changepercent from %s where changepercent>9.8"%(table_name)
    top_low_sql = "select name,changepercent from %s where changepercent<-6.9"%(table_name)
    ret = db.exec_and_fetch(top_high_sql)
    print("==================================")
    print("There are total %s stocks where change percent>9.8"%(len(ret)))
    print(ret)
    print("==================================")
    ret = db.exec_and_fetch(top_low_sql)
    print("There are total %s stocks where change percent<-6.9"%(len(ret)))
    print(ret)   

if __name__ == '__main__':
    com = Common()
    db_ip = com.read_conf('settings.conf', 'db', 'ip')
    db_user = com.read_conf('settings.conf', 'db', 'user')
    db_passwd = com.read_conf('settings.conf', 'db', 'passwd')
    db = Db(db_ip, db_user, db_passwd)
    stock_info = StockInfo()
    last_trading_date = stock_info.get_last_trading_date()

    get_market_status(db, 'stock_daily_%s'%(last_trading_date))
