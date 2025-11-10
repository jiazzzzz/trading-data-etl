# SQLite database handler
import sqlite3
from sqlalchemy import create_engine
from logger import Logger
import os

class DbSqlite():
    def __init__(self, db_path='jia-stk.db'):
        self.logger = Logger("db_sqlite")
        self.db_path = db_path
        self.logger.info(f"Using SQLite database: {self.db_path}")

    def create_engine(self):
        """Create SQLAlchemy engine for SQLite"""
        db_url = f"sqlite:///{self.db_path}"
        engine = create_engine(db_url, echo=False)
        return engine
    
    def connect(self):
        """Create a direct SQLite connection"""
        conn = sqlite3.connect(self.db_path)
        return conn

    def exec_and_fetch(self, sql_str):
        """Execute SQL and fetch results"""
        conn = self.connect()
        cursor = conn.cursor()
        try:
            cursor.execute(sql_str)
            ret = cursor.fetchall()
            return ret
        finally:
            conn.close()
    
    def exec(self, sql_str):
        """Execute SQL without fetching results"""
        conn = self.connect()
        cursor = conn.cursor()
        try:
            cursor.execute(sql_str)
            conn.commit()
            return cursor
        finally:
            conn.close()
    
    def drop_table(self, table_name):
        """Drop table if exists"""
        self.logger.info(f"Dropping table {table_name}")
        sql_str = f"DROP TABLE IF EXISTS {table_name}"
        self.exec(sql_str)

    def get_db_tables(self):
        """Get all table names in database"""
        sql = "SELECT name FROM sqlite_master WHERE type='table'"
        tbs = self.exec_and_fetch(sql)
        ret = [tb[0] for tb in tbs]
        return ret
    
    def get_db_daily_tables(self):
        """Get all daily stock tables"""
        ret = []
        for tb in self.get_db_tables():
            if 'stock_daily' in tb:
                ret.append(tb)
        return ret

    def get_stock_count(self):
        """Get total number of stocks in stock_list table"""
        try:
            sql = "SELECT COUNT(*) FROM stock_list"
            ret = self.exec_and_fetch(sql)
            return ret[0][0] if ret else 0
        except Exception as e:
            self.logger.warn(f"Failed to get stock count: {str(e)}")
            return 0

    def get_stock_symbol_from_pinyin(self, pinyin):
        """Get stock symbols by pinyin search"""
        ret = []
        sql = f"SELECT symbol FROM stock_list WHERE pinyin LIKE '%{pinyin}%'"
        for item in self.exec_and_fetch(sql):
            ret.append(item[0])
        return ret
    
    def get_stock_detail_from_name(self, stock_name):
        """Get stock details by name"""
        sql = f"SELECT symbol, industry, area FROM stock_list WHERE name='{stock_name}'"
        tmp = self.exec_and_fetch(sql)
        if len(tmp) == 0:
            return []
        else:
            return list(tmp[0])

if __name__ == '__main__':
    db = DbSqlite('test.db')
    tables = db.get_db_tables()
    print(f"Tables: {tables}")
