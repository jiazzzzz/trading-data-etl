#This program is to get the market status, maybe can be ran after 15:00 every day
import sys
sys.path.append('./lib')
from lib.stock_info import StockInfo
from lib.db import Db
import pandas as pd
import datetime
import time

#Dump stock list from tushare to database
def get_market_status(db_engine,table_name):
    conn = db_engine.connect()
    top_high_sql = "select name,changepercent from %s where changepercent>9.8"%(table_name)
    top_low_sql = "select name,changepercent from %s where changepercent<-6.9"%(table_name)
    ret = conn.execute(top_high_sql).fetchall()
    print("==================================")
    print("There are total %s stocks where change percent>9.8"%(len(ret)))
    print(ret)
    print("==================================")
    ret = conn.execute(top_low_sql).fetchall()
    print("There are total %s stocks where change percent<-6.9"%(len(ret)))
    print(ret)
    #print(ret)
    

#Dump daily data from sina to database, table is by date...
def dump_daily_data(db_engine, table_name):
    stock_info = StockInfo()
    for i in range(500):  #500*10 = total 5000 stocks, if increased, need to change 500 to a bigger value
        info = stock_info.get_daily_info_by_page(i)
        data = pd.read_json(info)
        data.to_sql(table_name, db_engine, index=False, if_exists='append', chunksize=5000)
        time.sleep(1)

if __name__ == '__main__':
    db = Db("10.64.165.81")
    db_engine = db.create_engine()
    today = datetime.datetime.now().strftime("%Y%m%d")
    market_status = get_market_status(db_engine, 'stock_daily_%s'%(today))
