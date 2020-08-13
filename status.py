#This program is to get the market status, maybe can be ran after 15:00 every day
import sys
sys.path.append('./lib')
from lib.stock_info import StockInfo
from lib.db import Db
from lib.common import Common
import pandas as pd
import datetime
import time
from prettytable import PrettyTable
import prettytable as pt
import textwrap


def format_f10(comment, max_line_length):
    #accumulated line length
    ACC_length = 0
    words = comment.split(" ")
    formatted_comment = ""
    for word in words:
        #if ACC_length + len(word) and a space is <= max_line_length 
        if ACC_length + (len(word) + 1) <= max_line_length:
            #append the word and a space
            formatted_comment = formatted_comment + word + " "
            #length = length + length of word + length of space
            ACC_length = ACC_length + len(word) + 1
        else:
            #append a line break, then the word and a space
            formatted_comment = formatted_comment + "\n" + word + " "
            #reset counter of length to the length of a word and a space
            ACC_length = len(word) + 1
    return formatted_comment

#Dump stock list from tushare to database
def get_market_status(db,target_type):
    stock_info = StockInfo()
    last_trading_date = stock_info.get_last_trading_date()
    table_name = 'stock_daily_%s'%(last_trading_date)
    if target_type == 'high':
        sql = "select name,changepercent,trade,turnoverratio from %s where changepercent>9.8"%(table_name)
    elif target_type == 'low':
        sql = "select name,changepercent,trade,turnoverratio from %s where changepercent<-6.9"%(table_name)
    ret = db.exec_and_fetch(sql)    
    print("There are total %s stocks meet requirement"%(len(ret)))
    #Add to prettytable
    table_high = PrettyTable(["名称", "涨幅", "当前价格", "换手", "代码", "行业", "地区", "F10"])    
    for item in ret:
        tmp_arr = list(item)
        stock_name = tmp_arr[0]
        if stock_name.startswith('N'):
            print("Skip stock %s"%(stock_name))
            continue
        other_fields = db.get_stock_detail_from_name(stock_name)
        #print(other_fields)
        if other_fields==[]:
            print("Stock %s has no other fields"%(stock_name))
            continue
        stock_id = other_fields[0]
        stock_f10 = "\n".join(textwrap.wrap(stock_info.get_f10(stock_id)[1],60)+['--------------------------------------------------------------------------------------------------------'])
        #print(stock_f10)
        tmp_arr = tmp_arr+other_fields+[stock_f10]
        table_high.add_row(tmp_arr)
    print(table_high)
    '''
    print("==================================")
    ret = db.exec_and_fetch(top_low_sql)
    print("There are total %s stocks where change percent<-6.9"%(len(ret)))
    print(ret)   
    '''

if __name__ == '__main__':
    com = Common()
    db_ip = com.read_conf('settings.conf', 'db', 'ip')
    db_user = com.read_conf('settings.conf', 'db', 'user')
    db_passwd = com.read_conf('settings.conf', 'db', 'passwd')
    db = Db(db_ip, db_user, db_passwd)
    stock_info = StockInfo()
    last_trading_date = stock_info.get_last_trading_date()

    get_market_status(db, 'high')

    print("===============================================================")

    get_market_status(db, 'low')
