# API Usage Examples

## Quick Test

Start the server:
```bash
cd server
go run main.go
```

## Using curl

### 1. Get API Info
```bash
curl http://localhost:8080/
```

### 2. List All Tables
```bash
curl http://localhost:8080/api/tables
```

Response:
```json
{
  "count": 2,
  "tables": [
    {"name": "stock_daily_20251107", "type": "table"},
    {"name": "stock_list", "type": "table"}
  ]
}
```

### 3. Get Stock List
```bash
curl http://localhost:8080/api/stocks?limit=5
```

Response:
```json
{
  "total": 5000,
  "count": 5,
  "stocks": [
    {
      "ts_code": "000001.SZ",
      "symbol": "000001",
      "name": "平安银行",
      "area": "深圳",
      "industry": "银行",
      "list_date": "19910403",
      "pinyin": "payx"
    }
  ]
}
```

### 4. Search Stocks
```bash
# Search by name
curl http://localhost:8080/api/search?q=科技

# Search by pinyin
curl http://localhost:8080/api/search?q=payx

# Search by symbol
curl http://localhost:8080/api/search?q=000001
```

### 5. Get Stock by Symbol
```bash
curl http://localhost:8080/api/stocks/000001
```

### 6. Get Daily Data
```bash
curl http://localhost:8080/api/daily/20251107?limit=10
```

### 7. Get Top Gainers
```bash
curl http://localhost:8080/api/daily/20251107/top?limit=10
```

Response:
```json
{
  "date": "20251107",
  "count": 10,
  "top_gainers": [
    {
      "symbol": "sh688286",
      "name": "N敏芯",
      "trade": 231.5,
      "pricechange": 168.83,
      "changepercent": 269.395,
      "volume": 9637479
    }
  ]
}
```

### 8. Custom SQL Query
```bash
# Count stocks
curl "http://localhost:8080/api/query?sql=SELECT%20COUNT(*)%20as%20total%20FROM%20stock_list"

# Stocks in Shanghai
curl "http://localhost:8080/api/query?sql=SELECT%20*%20FROM%20stock_list%20WHERE%20area='上海'%20LIMIT%2010"

# Tech industry stocks
curl "http://localhost:8080/api/query?sql=SELECT%20symbol,%20name,%20industry%20FROM%20stock_list%20WHERE%20industry%20LIKE%20'%25科技%25'%20LIMIT%2010"
```

## Using JavaScript (Browser/Node.js)

### Fetch API
```javascript
// Search stocks
async function searchStocks(keyword) {
  const response = await fetch(`http://localhost:8080/api/search?q=${keyword}`);
  const data = await response.json();
  console.log(data);
}

searchStocks('科技');
```

### Get Top Gainers
```javascript
async function getTopGainers(date, limit = 10) {
  const response = await fetch(`http://localhost:8080/api/daily/${date}/top?limit=${limit}`);
  const data = await response.json();
  
  data.top_gainers.forEach(stock => {
    console.log(`${stock.name}: ${stock.changepercent.toFixed(2)}%`);
  });
}

getTopGainers('20251107', 10);
```

### Custom Query
```javascript
async function customQuery(sql) {
  const response = await fetch(`http://localhost:8080/api/query?sql=${encodeURIComponent(sql)}`);
  const data = await response.json();
  console.table(data.results);
}

customQuery("SELECT * FROM stock_list WHERE area='北京' LIMIT 10");
```

## Using Python

### requests library
```python
import requests

BASE_URL = 'http://localhost:8080/api'

# Search stocks
response = requests.get(f'{BASE_URL}/search', params={'q': '科技'})
data = response.json()
print(f"Found {data['count']} stocks")

# Get top gainers
response = requests.get(f'{BASE_URL}/daily/20251107/top', params={'limit': 10})
data = response.json()
for stock in data['top_gainers']:
    print(f"{stock['name']}: {stock['changepercent']:.2f}%")

# Custom query
sql = "SELECT * FROM stock_list WHERE area='上海' LIMIT 10"
response = requests.get(f'{BASE_URL}/query', params={'sql': sql})
data = response.json()
print(data['results'])
```

## Using PowerShell

```powershell
# Get stock list
Invoke-RestMethod -Uri "http://localhost:8080/api/stocks?limit=10" | ConvertTo-Json

# Search stocks
Invoke-RestMethod -Uri "http://localhost:8080/api/search?q=科技" | ConvertTo-Json

# Get top gainers
$response = Invoke-RestMethod -Uri "http://localhost:8080/api/daily/20251107/top?limit=10"
$response.top_gainers | Format-Table name, changepercent, trade
```

## Integration Examples

### React Component
```jsx
import { useState, useEffect } from 'react';

function StockSearch() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);

  const search = async () => {
    const response = await fetch(`http://localhost:8080/api/search?q=${query}`);
    const data = await response.json();
    setResults(data.stocks || []);
  };

  return (
    <div>
      <input 
        value={query} 
        onChange={(e) => setQuery(e.target.value)}
        onKeyPress={(e) => e.key === 'Enter' && search()}
      />
      <button onClick={search}>Search</button>
      
      <ul>
        {results.map(stock => (
          <li key={stock.symbol}>
            {stock.symbol} - {stock.name} ({stock.area})
          </li>
        ))}
      </ul>
    </div>
  );
}
```

### Vue Component
```vue
<template>
  <div>
    <input v-model="query" @keyup.enter="search" />
    <button @click="search">Search</button>
    
    <ul>
      <li v-for="stock in results" :key="stock.symbol">
        {{ stock.symbol }} - {{ stock.name }} ({{ stock.area }})
      </li>
    </ul>
  </div>
</template>

<script>
export default {
  data() {
    return {
      query: '',
      results: []
    };
  },
  methods: {
    async search() {
      const response = await fetch(`http://localhost:8080/api/search?q=${this.query}`);
      const data = await response.json();
      this.results = data.stocks || [];
    }
  }
};
</script>
```

## Error Handling

```javascript
async function safeQuery(endpoint) {
  try {
    const response = await fetch(`http://localhost:8080/api/${endpoint}`);
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const data = await response.json();
    
    if (data.error) {
      console.error('API Error:', data.error);
      return null;
    }
    
    return data;
  } catch (error) {
    console.error('Request failed:', error);
    return null;
  }
}

// Usage
const data = await safeQuery('stocks?limit=10');
if (data) {
  console.log(data.stocks);
}
```

## Rate Limiting (Optional)

If you need to add rate limiting, consider using middleware like:
- `github.com/ulule/limiter/v3`
- `golang.org/x/time/rate`

## HTTPS (Production)

For production, use a reverse proxy like nginx or Caddy:

```nginx
server {
    listen 443 ssl;
    server_name api.yourdomain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```
