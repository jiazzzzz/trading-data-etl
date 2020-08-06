import requests
from logger import Logger

class StockInfo():
    def __init__(self):
        self.logger = Logger("StockInfo")
    
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

if __name__ == '__main__':
    t = StockInfo()
    resp = t.get_daily_info_by_page(1)
    print(resp)


