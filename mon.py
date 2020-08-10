#!/usr/bin/python3
import sys
sys.path.append('./lib')
from lib.stock_info import StockInfo

stock_list = ['sz002152','sz000788','sz000815', 'sh600251', 'sh000001', 'sz399006']

t = StockInfo()
for s in stock_list:
    print(t.get_live_status_pretty(s))

