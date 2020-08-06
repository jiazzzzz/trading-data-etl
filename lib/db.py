from sqlalchemy import create_engine
import pymysql
from logger import Logger

class Db():
    def __init__(self):
        self.logger = Logger("db")

    def connect(self, db_info):
        engine = create_engine(db_info, max_overflow=0, pool_size=5)
        return engine


