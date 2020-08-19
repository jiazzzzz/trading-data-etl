#!/usr/bin/python3
#CLI interactive program to do live query
import sys
sys.path.append('./lib')
from lib.stock_info import StockInfo
from lib.common import Common
from lib.db import Db
from lib.trader import Trader
from tabulate import tabulate


t = StockInfo()
stock_list = ['sh603683','sz000998','sz002152','sz000788','sz000815', 'sh600251', 'sh000001', 'sz399006']
com = Common()
db_ip = com.read_conf('settings.conf', 'db', 'ip')
db_user = com.read_conf('settings.conf', 'db', 'user')
db_passwd = com.read_conf('settings.conf', 'db', 'passwd')
db = Db(db_ip, db_user, db_passwd)

sim_trader = Trader('jia-sim.log')

while 1:
    to_do = input("stk>")
    if 'zx'==to_do:
        for s in stock_list:
            print(t.get_live_status_pretty(s))
    elif 'top'==to_do:
        df = t.get_market_status_from_xueqiu('desc',1,100)
        print(df.to_markdown())
        #print(tabulate(df, headers='keys', tablefmt='psql'))
    elif 'bot'==to_do:
        df = t.get_market_status_from_xueqiu('asc',1,100)
        print(df.to_markdown())
        #print(tabulate(df, headers='keys', tablefmt='psql'))
    elif to_do.startswith('60') or to_do.startswith('688') or to_do.startswith('300') or to_do.startswith('002') or to_do.startswith('000'):
        print(t.get_live_status_pretty(to_do))
    elif 'quit' in to_do or 'exit' in to_do or 'q'==to_do:
        break
    elif to_do.startswith('f10'):
        stock_id = to_do.split(' ')[1]
        v = t.get_f10(stock_id)
        print(v)
    elif to_do.startswith('buy') or to_do.startswith('BUY'): #buy stock 
        arr = to_do.split(' ')
        stock_id = arr[1]
        stock_count = arr[2]
        cur_price = float(t.get_live_status(stock_id).split(',')[3])
        sim_trader.buy(stock_id,stock_count,cur_price)
    elif to_do.startswith('sell') or to_do.startswith('SELL'): #sell stock
        arr = to_do.split(' ')
        stock_id = arr[1]
        stock_count = arr[2]
        cur_price = float(t.get_live_status(stock_id).split(',')[3])
        sim_trader.sell(stock_id,stock_count,cur_price)
    elif to_do.startswith('holding'): #获取现在资金和持股状态
        status = sim_trader.get_cur_status()
        print("当前资金: %s"%(status['cash']))
        print("当前持股: %s"%(status['stocks']))
        total_value = 0
        for (s_id,s_count) in status['stocks'].items():
            total_value = total_value + s_count * float(t.get_live_status(s_id).split(',')[3])
        print("当前市值: %s"%(total_value))
        print("当前盈亏：%s%%"%((total_value+float(status['cash'])-100000)/1000))
    else:
        q = db.get_stock_symbol_from_pinyin(to_do)
        for s in q:
            print("Get stock %s"%(s))
            print(t.get_live_status_pretty(s))

