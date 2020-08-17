import sys
sys.path.append('./lib')
from lib.stock_info import StockInfo
from lib.trader import Trader

t = StockInfo()
v = t.get_daily_info_by_page(1,order='asc')
p = pd.read_json(v)

sim_trader = Trader('jia-test.log')




