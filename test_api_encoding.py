import requests
import json

# Test the filter API
url = "http://localhost:8080/api/filter/stocks?days=7&min_change=7&max_change=100&max_mktcap=1000&limit=100"
response = requests.get(url)

print("Response encoding:", response.encoding)
print("Response status:", response.status_code)

data = response.json()
stocks = data.get('filtered_stocks', [])

print(f"\nTotal stocks: {len(stocks)}")
print("\nChecking specific stocks:")

for stock in stocks:
    code = stock.get('stock_code', '')
    name = stock.get('stock_name', '')
    if code in ['000514', '002136', '000007']:
        print(f"  {code}: '{name}' (length: {len(name)}, bytes: {len(name.encode('utf-8'))})")
        # Print each character
        for i, char in enumerate(name):
            print(f"    [{i}] {char} (U+{ord(char):04X})")
