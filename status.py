#This program is to get the market status, maybe can be ran after 15:00 every day
import sys
sys.path.append('./lib')
from lib.stock_info import StockInfo
from lib.db import Db
from lib.common import Common
import pandas as pd
import datetime
import time

output_template = '''
<!DOCTYPE html> 
<html> 
<head> 
<meta charset="utf-8" /> 
<title>选股输出</title> 
</head> 
 
<body> 
PLACEHOLDER_OUTPUT
</body> 
</html> 
'''
#Dump stock list from tushare to database
def get_market_status(db,table_name):
    #conn = db_engine.connect()
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

def get_status(db, table_name):
    sql = "SELECT symbol, name, changepercent, open, high, low, per, turnoverratio FROM stock_daily_20200807\
         WHERE turnoverratio>5 AND changepercent>5"
    db_engine = db.create_engine()
    ret = pd.read_sql(sql, db_engine)
    #print(ret)
    return ret 

def insert_image(row):
    out = {}
    image_url = "http://image.sinajs.cn/newchart/daily/n/%s.gif"%(row['symbol'])
    out['rx'] = "<img src=\"%s\" width=\"521\" height=\"287\" />"%(image_url)
    return pd.Series(out)


if __name__ == '__main__':
    com = Common()
    db_ip = com.read_conf('settings.conf', 'db', 'ip')
    db_user = com.read_conf('settings.conf', 'db', 'user')
    db_passwd = com.read_conf('settings.conf', 'db', 'passwd')
    db = Db(db_ip, db_user, db_passwd)
    stock_info = StockInfo()
    last_trading_date = stock_info.get_last_trading_date()
    get_market_status(db, 'stock_daily_%s'%(last_trading_date))
    df = get_status(db, 'stock_daily_%s'%(last_trading_date))
    df[['rx']] = df.apply(insert_image, axis=1)
    print(df)
    html_body = df.to_html(escape=False)
    with open("output.html",'w') as f:
        final_output = output_template.replace('PLACEHOLDER_OUTPUT', html_body)
        f.write(final_output)
