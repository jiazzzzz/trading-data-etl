#!/usr/bin/python3

from bs4 import BeautifulSoup
import datetime
import requests

url = "http://quotes.money.163.com/data/ipo/shengou.html"
resp = requests.get(url)

table_data = [[cell.text for cell in row("td")] for row in BeautifulSoup(resp.text,features="lxml")("tr")]
today = datetime.datetime.now().strftime("%Y-%m-%d")
#print(today)

for item in table_data:
    if item[4] == today and not (item[2].startswith('688') or item[2].startswith('300')):
        print(item)

