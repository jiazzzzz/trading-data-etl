#This file is to get information from outside, not db
import requests
from logger import Logger
import re
import tushare as ts
import json
from bs4 import BeautifulSoup
from pandas import DataFrame
import pandas as pd

class StockInfo():
    def __init__(self):
        self.logger = Logger("StockInfo")
    
    def get_last_trading_date(self):        
        self.logger.info("Getting last trading date live...")
        url = "http://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData?\
            symbol=sh000001&scale=5&ma=no&datalen=5"
        resp = requests.get(url)
        date = json.loads(resp.text)   
        ret = date[0]['day'].split(' ')[0].replace('-','')
        self.logger.info("Last trading date is %s"%(ret))
        return ret
	
    #Below functions are getting data from TDX folder
    def get_stock_list_from_tdx(self, tdx_folder='./tdx'):
        """
        Get stock list from TDX folder files.
        TDX files are named like: SH#600000.txt, SZ#000001.txt, BJ#920000.txt
        """
        import os
        import glob
        
        self.logger.info(f"Getting stock list from TDX folder: {tdx_folder}")
        
        stock_data = []
        
        # Get all txt files in tdx folder
        pattern = os.path.join(tdx_folder, '*.txt')
        files = glob.glob(pattern)
        
        self.logger.info(f"Found {len(files)} TDX files")
        
        for file_path in files:
            try:
                # Parse filename: SH#600000.txt -> exchange=SH, code=600000
                filename = os.path.basename(file_path)
                if '#' not in filename:
                    continue
                    
                parts = filename.replace('.txt', '').split('#')
                if len(parts) != 2:
                    continue
                    
                exchange = parts[0].upper()  # SH, SZ, BJ
                symbol = parts[1]  # 600000, 000001, etc.
                
                # Read first line to get stock name
                with open(file_path, 'r', encoding='gbk', errors='ignore') as f:
                    first_line = f.readline().strip()
                    # First line format: "600000 浦发银行 日线 前复权" or "002181 粤 传 媒 日线 不复权"
                    # Stock name may have spaces between characters
                    line_parts = first_line.split()
                    if len(line_parts) >= 2:
                        # Find where "日线" appears to know where name ends
                        if '日线' in line_parts:
                            day_index = line_parts.index('日线')
                            # Name is everything between code and "日线"
                            name = ''.join(line_parts[1:day_index])
                        else:
                            # Fallback: take everything after code, remove spaces
                            name = ''.join(line_parts[1:])
                    else:
                        name = symbol
                
                # Create ts_code format: 600000.SH, 000001.SZ, 920000.BJ
                ts_code = f"{symbol}.{exchange}"
                
                stock_data.append({
                    'ts_code': ts_code,
                    'symbol': symbol,
                    'name': name,
                    'area': '',  # Not available in TDX files
                    'industry': '',  # Not available in TDX files
                    'list_date': ''  # Not available in TDX files
                })
                
            except Exception as e:
                self.logger.err(f"Error processing file {file_path}: {str(e)}")
                continue
        
        self.logger.info(f"Successfully parsed {len(stock_data)} stocks from TDX files")
        
        # Convert to DataFrame
        df = DataFrame(stock_data)
        return df
    
    #Below functions are getting data from tushare (deprecated, use get_stock_list_from_tdx instead)
    def get_stock_list(self):
        """
        DEPRECATED: Use get_stock_list_from_tdx() instead.
        This method requires tushare API which may not be available.
        """
        self.logger.warn("get_stock_list() is deprecated. Using get_stock_list_from_tdx() instead.")
        return self.get_stock_list_from_tdx()
        
    def get_stock_list_from_tushare(self):
        """
        Get stock list from tushare API (requires API key).
        """
        tushare_api_key = '85f02fe6125d57c0cb8b467121ce63620a59e2ccc419b3a3e0f7dad7'
        pro = ts.pro_api(tushare_api_key)
        stock_list = pro.stock_basic(exchange='', list_status='L', fields='ts_code,symbol,name,area,industry,list_date')
        return stock_list
    
    #Below functions are getting data from xueqiu
    def get_market_status_from_xueqiu(self,direction,page_number,page_size):
        #direction = asc 跌幅榜， direction = desc 涨幅榜
        url = "https://xueqiu.com/stock/cata/stocklist.json?page=%s&size=%s&order=%s&orderby=percent&type=11%%2C12"%(page_number,page_size,direction)
        #url = "https://xueqiu.com/stock/cata/stocklist.json?page=%s&size=%s&order=%s&orderby=percent&type=11%%2C12&_=1541985912951"%(page_number,page_size,direction)
        #self.logger.info(url)
        return self.get_xueqiu_info(url)
    
    def get_xueqiu_info(self,url):
        cookie_url = "https://xueqiu.com"
        headers = {'User-Agent':'Mozilla/5.0 (Windows NT 6.3; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36'}
        r = requests.get(cookie_url,headers=headers)
        cookies = r.cookies
        r1 = requests.get(url,headers=headers,cookies=cookies)
        #self.logger.info(r1.text)
        stock_list = eval(r1.text)['stocks']
        return DataFrame(stock_list)

    #Below functions are getting data from 163
    def get_f10(self, stock_id): #only available to normal stock id like 000002
        '''    
        Will return 3 values: 主营， 公司业务， 公司历史
        '''
        url = "http://quotes.money.163.com/f10/gszl_"+stock_id+".html"
        resp = requests.get(url)
        soup = BeautifulSoup(resp.text,features="lxml")
        ret = []
        for item in soup.find_all(colspan='3'): #find all useful informations
            r = re.findall('>(.*?)<', str(item))
            if len(r)>0:
                ret.append(r[0].strip())
        return ret
    
    #Below functions are getting data from sina
    def get_daily_info_by_page(self,page_num,sort_by='changepercent',order='desc'):
        asc = 0
        if order=='asc':
            asc = 1
        url = "http://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/Market_Center.getHQNodeData?\
                 page=%s&num=100&sort=%s&asc=%s&node=hs_a&symbol=&_s_r_a=init"%(page_num,sort_by,asc)
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
        #Auto add sz/sh to stock_id if it doesn't exist
        if stock_id.startswith('60') or stock_id.startswith('688'):
            stock_id = "sh%s"%(stock_id)
        elif stock_id.startswith('300') or stock_id.startswith('002') or stock_id.startswith('000'):
            stock_id = "sz%s"%(stock_id)
        url = "http://hq.sinajs.cn/list=%s"%(stock_id)
        r = requests.get(url)
        if r.status_code != 200:
            return ret
        re_info = re.compile(r'="(.*)"')
        ret = re_info.findall(r.text)[0]
        return ret
    
    def get_live_status_pretty(self,stock_id):
        info = self.get_live_status(stock_id).split(',')
        #print(info)
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
    #v = t.get_last_trading_date()
    #v = t.get_f10('603077')
    #v = t.get_stock_list()
    #v = t.get_daily_info_by_page(1,sort_by='turnoverratio')
    v = t.get_daily_info_by_page(1)
    df = pd.read_json(v)
    #p1 = df[df.changepercent<5]
    df = df[df.turnoverratio<10]
    df = df[df.turnoverratio>3]
    with pd.option_context('display.max_rows', None, 'display.max_columns', None):  # more options can be specified also
        print(df)


