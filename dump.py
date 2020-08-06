import sys
sys.path.append('./lib')
from lib.stock_info import StockInfo
from lib.db import Db
import pandas as pd

def dump():
    db_ip = "10.64.165.81"
    db = Db()
    db_engine = db.connect("mysql+pymysql://root:mariadb@%s:3306/jia-stk"%db_ip)
    stock_info = StockInfo()
    for i in range(5):
        info = stock_info.get_daily_info_by_page(i)
        data = pd.read_json(info)
        data.to_sql('stock_daily',db_engine, index=False, if_exists='append', chunksize=5000)

if __name__ == '__main__':
    dump()
