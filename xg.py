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


def get_status(db, table_list):
    db_engine = db.create_engine()
    
    init_sql ="SELECT symbol, name, changepercent, turnoverratio FROM %s\
        WHERE turnoverratio>5 AND changepercent>3"%(table_list[0])
    print(init_sql)
    ret = pd.read_sql(init_sql, db_engine)
    
    
    for i in range(len(table_list)-1):    
        sql = "SELECT symbol, name, changepercent,turnoverratio FROM %s\
        WHERE turnoverratio>5 AND changepercent>5"%(table_list[i+1])
        tmp_pd = pd.read_sql(sql, db_engine)
        #print(tmp_pd)
        ret = pd.merge(tmp_pd, ret, how='inner', on=['symbol'])
        #print(ret)
    return ret 

#Add a row to pd, indicating the daily chart of a stock
def insert_image(row):
    out = {}
    image_url = "http://image.sinajs.cn/newchart/daily/n/%s.gif"%(row['symbol'])
    #print(image_url)
    out['daily_line'] = "<img src=\"%s\" width=\"521\" height=\"287\" />"%(image_url)
    return pd.Series(out)

def generate_report(df,output_file):
    html_body = df.to_html(escape=False)
    with open(output_file,'w') as f:
        final_output = output_template.replace('PLACEHOLDER_OUTPUT', html_body)
        f.write(final_output)

def xg1():
    pass

def xg2():
    pass

if __name__ == '__main__':
    com = Common()
    db_ip = com.read_conf('settings.conf', 'db', 'ip')
    db_user = com.read_conf('settings.conf', 'db', 'user')
    db_passwd = com.read_conf('settings.conf', 'db', 'passwd')
    db = Db(db_ip, db_user, db_passwd)

    table_list = db.get_db_daily_tables()
    df = get_status(db, table_list)    
    df[['rx']] = df.apply(insert_image, axis=1)
    print(df)

    generate_report(df)
