import sys
sys.path.append('./lib')
from lib.stock_info import StockInfo
from lib.db import Db
from lib.common import Common
import pandas as pd
import datetime
import time




if __name__ == '__main__':
    com = Common()
    db_ip = com.read_conf('settings.conf', 'db', 'ip')
    db_user = com.read_conf('settings.conf', 'db', 'user')
    db_passwd = com.read_conf('settings.conf', 'db', 'passwd')
    db = Db(db_ip, db_user, db_passwd)

    table_list = db.get_db_daily_tables()
    print(table_list)
    candidate = []
    while 1:
        to_do = input("static>")
        if to_do.startswith('clear'):
            candidate = []
        elif to_do.startswith('60') or to_do.startswith('688') or to_do.startswith('00') or to_do.startswith('300'):
            print("下一步为你选择版块~")
        else:
            pass

    