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
def dump_stock_list(db,table_name):    
    db.drop_table("stock_list")
    db_engine = db.create_engine()
    stock_info = StockInfo()
    stock_list = stock_info.get_stock_list()
    stock_list[['pinyin']] = stock_list.apply(insert_pinyin, axis=1)
    print(stock_list)
    stock_list.to_sql(table_name, db_engine, index=False, if_exists='append', chunksize=5000)

def insert_pinyin(row):
    out = {}
    com = Common()
    out['pinyin'] = com.get_py_from_name(row['name'])
    return pd.Series(out)

#Dump daily data from sina to database, table is by date...
def dump_daily_data(db):    
    stock_info = StockInfo()
    last_trading_date = stock_info.get_last_trading_date()
    table_name = "stock_daily_%s"%(last_trading_date)
    total_stock_count = db.get_stock_count()
    estimated_pages = int(1+int(total_stock_count)/100)

    db_engine = db.create_engine()
    print("Total estimated pages are %s"%(estimated_pages))
    for i in range(estimated_pages):  #50*100 = total 5000 stocks, if increased, need to change 500 to a bigger value
        print("Retriving page %s"%(i+1))
        info = stock_info.get_daily_info_by_page(i+1)
        data = pd.read_json(info)
        data.to_sql(table_name, db_engine, index=False, if_exists='append', chunksize=5000)
        time.sleep(1)

if __name__ == '__main__':
    com = Common()
    db_ip = com.read_conf('settings.conf', 'db', 'ip')
    db_user = com.read_conf('settings.conf', 'db', 'user')
    db_passwd = com.read_conf('settings.conf', 'db', 'passwd')
    db = Db(db_ip, db_user, db_passwd)
    
    stock_info = StockInfo()

    #Dump stock list
    dump_stock_list(db, "stock_list")


    #Dump daily trading info
    dump_daily_data(db)
    '''
    last_trading_date = stock_info.get_last_trading_date()
    table_name = "stock_daily_%s"%(last_trading_date)
    total_stock_count = db.get_stock_count()
    db.drop_table(table_name)
    dump_daily_data(db_engine, table_name, total_stock_count)
    '''
