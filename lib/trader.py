import os
import datetime
from logger import Logger

'''
Trading log looks like below:
cash,holding,date-time
100000,{'000002':300;'000151':400},2020-08-14 12:00:00
'''

class Trader():
    def __init__(self, trading_log, init_rmb='100000'):
        self.trading_log = trading_log
        self.logger = Logger("StockTrade")
        if not os.path.exists(trading_log):
            with open(self.trading_log,'w') as f:
                item = "%s,{},%s"%(init_rmb,self._get_date())
                f.write(item)

    def _get_date(self):
        return datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')

    def buy(self,stock_id,stock_count,stock_price):
        cur = self.get_cur_status()        
        cash = cur['cash'] - stock_count*stock_price
        if cash<0:
            self.logger.info("Cash not enough, skip this buy operation...")
        else:
            self.logger.info("Start to buy stock %s with count %s at price %s"%(stock_id, stock_count, stock_price))
            stocks = cur['stocks']
            for (s_id,s_count) in stocks.items():
                print(s_id)
                print(s_count)
            pass

    def sell(self,stock_id,stock_count,stock_price):
        pass

    def get_cur_status(self):
        ret = {}
        with open(self.trading_log,'r') as f:
            lines = f.readlines()
            item = lines[-1] #get last line
            temp_arr = item.split(',')
            ret['cash'] = int(temp_arr[0])
            ret['stocks'] = eval(temp_arr[1])
        return ret

if __name__ == '__main__':
    t = Trader('test.log')
    v = t.get_cur_status()
    print(v)