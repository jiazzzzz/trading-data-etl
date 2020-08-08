#This program need to be executed everyday to save data to db
import sys
sys.path.append('./lib')
from lib.stock_info import StockInfo
from lib.db import Db
from lib.common import Common
import pandas as pd
import datetime
import time

#Dump stock list from tushare to database
def dump_stock_list(db_engine,table_name):
    stock_info = StockInfo()
    stock_list = stock_info.get_stock_list()
    stock_list.to_sql(table_name, db_engine, index=False, if_exists='append', chunksize=5000)
	

#Dump daily data from sina to database, table is by date...
def dump_daily_data(db_engine, table_name):
    stock_info = StockInfo()
    for i in range(500):  #500*10 = total 5000 stocks, if increased, need to change 500 to a bigger value
        info = stock_info.get_daily_info_by_page(i)
        data = pd.read_json(info)
        data.to_sql(table_name, db_engine, index=False, if_exists='append', chunksize=5000)
        time.sleep(1)

if __name__ == '__main__':
    com = Common()
    db_ip = com.read_conf('settings.conf', 'db', 'ip')
    db_user = com.read_conf('settings.conf', 'db', 'user')
    db_passwd = com.read_conf('settings.conf', 'db', 'passwd')
    db = Db(db_ip, db_user, db_passwd)
    db_engine = db.create_engine()
    today = datetime.datetime.now().strftime("%Y%m%d")
    #dump_stock_list(db_engine, 'stock_list')
    #dump_daily_data(db_engine, 'stock_daily_%s'%(today))
    dump_daily_data(db_engine, 'stock_daily_20200807')
