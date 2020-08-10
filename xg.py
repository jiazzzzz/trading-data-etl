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
    pd_list = []
    for tb in table_list:
        if 'stock_daily' in tb:
            print(tb)
            sql = "SELECT symbol, name, changepercent, open, high, low, per, turnoverratio FROM %s\
            WHERE turnoverratio>5 AND changepercent>5"%(tb)
            tmp_pd = pd.read_sql(sql, db_engine)
            print(tmp_pd)
            pd_list.append(tmp_pd)
    ret = pd.concat(pd_list, axis=1, join='inner')
    return ret 

def insert_image(row):
    out = {}
    image_url = "http://image.sinajs.cn/newchart/daily/n/%s.gif"%(row['symbol'])
    #print(image_url)
    out['rx'] = "<img src=\"%s\" width=\"521\" height=\"287\" />"%(image_url)
    return pd.Series(out)


if __name__ == '__main__':
    com = Common()
    db_ip = com.read_conf('settings.conf', 'db', 'ip')
    db_user = com.read_conf('settings.conf', 'db', 'user')
    db_passwd = com.read_conf('settings.conf', 'db', 'passwd')
    db = Db(db_ip, db_user, db_passwd)

    table_list = db.get_db_tables()
    df = get_status(db, table_list)
    print(df)
    #df[['rx']] = df.apply(insert_image, axis=1)
    #print(df)
    '''
    html_body = df.to_html(escape=False)
    with open("output.html",'w') as f:
        final_output = output_template.replace('PLACEHOLDER_OUTPUT', html_body)
        f.write(final_output)
    '''
