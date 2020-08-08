#This file is to get information from outside, not db
import requests
from logger import Logger
import re
import tushare as ts

class StockInfo():
    def __init__(self):
        self.logger = Logger("StockInfo")
	
    def get_stock_list(self):
        tushare_api_key = '85f02fe6125d57c0cb8b467121ce63620a59e2ccc419b3a3e0f7dad7'
        pro = ts.pro_api(tushare_api_key)
        stock_list = pro.stock_basic(exchange='', list_status='L', fields='ts_code,symbol,name,area,industry,list_date')
        return stock_list

    
    def get_daily_info_by_page(self,page_num):
        url = "http://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/Market_Center.getHQNodeData?\
                 page=%s&num=100&sort=changepercent&asc=0&node=hs_a&symbol=&_s_r_a=init"%(page_num)
        resp =  requests.get(url)
        #self.logger.info(resp.text)
        if resp.text == 'null':
            return ''
        elif '?xml' in resp.text:
            self.logger.info(resp.text)
        else:
            return resp.text

    def get_live_status(self,stock_id):
        ret = ""
        url = "http://hq.sinajs.cn/list=%s"%(stock_id)        
        r = requests.get(url)
        if r.status_code != 200:
            return ret
        re_info = re.compile(r'="(.*)"')
        ret = re_info.findall(r.text)[0]
        return ret
    
    def get_live_status_pretty(self,stock_id):
        info = self.get_live_status(stock_id).split(',')
        cur_price = float(info[3])
        last_day_price = float(info[2])
        open_price = float(info[1])
        stock_name = info[0]
        if len(stock_name)<4:            
            stock_name = "%s%s"%('  ',stock_name)        
        aoi = round((cur_price-last_day_price)*100/last_day_price,2)        
        volume = round(float(info[8])/1000000,2)
        rmb = round(float(info[9])/100000000,2)
        ret = "%s(%s) | %8s%% | %8s | %8s(万手) | %8s(亿) "%(stock_name,stock_id,aoi,info[3],volume,rmb)
        return ret

if __name__ == '__main__':
    t = StockInfo()
    resp = t.get_daily_info_by_page(1)
    print(resp)


