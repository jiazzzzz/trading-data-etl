from sqlalchemy import create_engine
import pymysql
from logger import Logger

class Db():
    def __init__(self,db_ip):
        self.logger = Logger("db")
        self.db_ip = db_ip

    def create_engine(self):
        db_info = "mysql+pymysql://root:mariadb@%s:3306/jia-stk"%(self.db_ip)
        engine = create_engine(db_info, max_overflow=0, pool_size=5)
        return engine

    def exec(self, sql_str):
        pass

