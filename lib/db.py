#This file will handle all db related operations
from sqlalchemy import create_engine
import pymysql
from logger import Logger

class Db():
    def __init__(self, db_ip, user, passwd):
        self.logger = Logger("db")
        self.db_ip = db_ip
        self.db_user = user
        self.db_passwd = passwd

    def create_engine(self):
        db_info = "mysql+pymysql://%s:%s@%s:3306/jia-stk"%(self.db_user, self.db_passwd, self.db_ip)
        engine = create_engine(db_info, max_overflow=0, pool_size=5)
        return engine
    
    def connect(self):
        db_engine = self.create_engine()
        conn = db_engine.connect()
        return conn

    def exec(self, sql_str):
        conn = self.connect()
        ret = conn.execute(sql_str).fetchall()
        return ret

    #Below are common high-level wrap functions to get db items, can be called from outside
    def get_stock_count(self):
        sql = "select * from stock_list"
        ret = self.exec(sql)
        return len(ret)

if __name__ == '__main__':
    t = Db('127.0.0.1', 'root', 'su')
    v = t.get_stock_count()
    print(v)


