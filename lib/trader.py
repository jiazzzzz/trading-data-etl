import os
import datetime
from logger import Logger

'''
Trading log looks like below:
cash#holding#date-time#trading activity
100000#{'000002':300,'000151':400}#2020-08-14 12:00:00#buy 000002:300
'''

class Trader():
    def __init__(self, trading_log, init_rmb='100000'):
        self.trading_log = trading_log
        self.logger = Logger("StockTrade")
        if not os.path.exists(trading_log):
            self.logger.info("No trading log %s exist, creating..."%(trading_log))
            with open(self.trading_log,'w') as f:
                item = "%s#{}#%s#init trading log\n"%(init_rmb,self._get_date())
                f.write(item)
        else:
            self.logger.info("Trading log %s already exist, continue to record"%(trading_log))

    def _get_date(self):
        return datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
    
    def _writing_trading_log(self,trading_action):
        with open(self.trading_log,'a') as f:
            f.write(trading_action)

    def buy(self,stock_id,stock_count,stock_price):
        '''
        Example: buy('000002',1000, 22.2) means 以22.2的价格买入万科1000股
                 buy('000002',0.3, 22.2) means 以22.2的价格买入万科剩余资金的30%股份
        '''
        cur = self.get_cur_status()
        if stock_count>=100:       #普通买入模式  
            self.logger.info("普通买入模式")
            cash = cur['cash'] - stock_count*stock_price
        else: #资金比买入模式
            self.logger.info("资金比买入模式，买入剩余资金的百分之%s"%(int(stock_count*100)))
            stock_count = int((cur['cash']*stock_count)/stock_price)
            self.logger.info("目前可以买%s共%s股"%(stock_id,stock_count))
            cash = cur['cash'] - stock_count*stock_price
        if cash<0:
            self.logger.info("Cash not enough, skip this buy operation...")
        else:
            self.logger.info("Start to buy stock %s with count %s at price %s"%(stock_id, stock_count, stock_price))
            stocks = cur['stocks']
            self.logger.info("Current stocks are %s"%(stocks))
            if stocks == {}:
                self.logger.info("目前是空仓状态，可以买入")
                stocks[stock_id]=stock_count
                trading_activity = "%s#%s#%s#Buy %s:%s at price %s\n"%(cash,str(stocks),self._get_date(),stock_id,stock_count,stock_price)
                self.logger.info(trading_activity)
                self._writing_trading_log(trading_activity)
            else:
                for (s_id,s_count) in stocks.items():
                    if s_id == stock_id:
                        self.logger.info("Already have stock %s, 加仓%s股"%(stock_id,stock_count))
                        stocks[stock_id] = s_count+stock_count
                        self.logger.info("目前持有股票%s:%s股"%(stock_id,stocks[stock_id]))
                        trading_activity = "%s#%s#%s#Buy %s:%s at price %s\n"%(cash,str(stocks),self._get_date(),stock_id,stock_count,stock_price)
                        self.logger.info(trading_activity)
                        self._writing_trading_log(trading_activity)
                        return
                #When buy a new stock_id
                self.logger.info("买入的股票%s目前没有持仓"%(stock_id))
                stocks[stock_id] = stock_count
                trading_activity = "%s#%s#%s#Buy %s:%s at price %s\n"%(cash,str(stocks),self._get_date(),stock_id,stock_count,stock_price)
                self.logger.info(trading_activity)
                self._writing_trading_log(trading_activity)

    def sell(self,stock_id,stock_count,stock_price):
        cur = self.get_cur_status()
        stocks = cur['stocks']
        if stocks == {}:
            self.logger.info("目前持仓没有股票，不能卖出")
            return
        if not stock_id in stocks.keys():
            self.logger.info("目前想要卖出的股票没有持仓，放弃")
            return
        for (s_id,s_count) in stocks.items():
            if s_id == stock_id:
                self.logger.info("目前股票%s持仓%s"%(s_id,s_count))
                if s_count<stock_count:
                    self.logger.info("股票持仓小于卖出数量，不能卖出")
                    return
                if stock_count<1 and s_count*stock_count>100: #百分比卖出模式，保证卖出数量大于100股
                    stock_count = s_count*stock_count
                stocks[stock_id] = s_count-stock_count
                cash = cur['cash'] + int(stock_count*stock_price)
                trading_activity = "%s#%s#%s#Sell %s:%s at price %s\n"%(cash,str(stocks),self._get_date(),stock_id,stock_count,stock_price)
                self.logger.info(trading_activity)
                self._writing_trading_log(trading_activity)


    def get_cur_status(self):
        ret = {}
        with open(self.trading_log,'r') as f:
            lines = f.readlines()
            item = lines[-1] #get last line
            temp_arr = item.split('#')
            self.logger.info(temp_arr[1])
            ret['cash'] = float(temp_arr[0])
            ret['stocks'] = eval(temp_arr[1])
        return ret

if __name__ == '__main__':
    t = Trader('test.log')
    v = t.get_cur_status()
    print(v)
    #t.buy('000415',1000,15)
    #t.buy('000002',0.3,15)
    #t.sell("000415",0.3,20)
    t.sell('000002',0.3,10)