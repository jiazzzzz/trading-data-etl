# stock_updater — 股票日线数据更新工具

从新浪/腾讯 API 拉取全 A 股日 K 线数据，写入 SQLite 数据库供其他工具查询。

## 用法

```
update.exe -db PATH [-from YYYY-MM-DD] [-to YYYY-MM-DD] [-workers N]
update.exe -db PATH -update-concepts [-workers N]
```

## 目标

为 `query.exe`（涨跌幅/均线筛选）提供 `stock_history` 数据源。

## 架构

```
main.go     入口，解析 CLI 参数
updater.go  核心逻辑：UpdateToday / UpdateRange / Backfill / UpdateConcepts
api.go      数据抓取：新浪股票列表、新浪日线、新浪概念板块、腾讯 K 线
db.go       SQLite 操作：建表/插入/查询（stock_list, stock_daily_*, stock_history, stock_concepts）
models.go   数据结构定义
```

### 数据流

```
新浪 API ──→ fetchStockList()        ──→ stock_list 表
新浪 API ──→ fetchDailyData()        ──→ stock_daily_YYYYMMDD + stock_history
腾讯 API  ──→ fetchKLine()           ──→ stock_history（回填缺失日期）
新浪 API  ──→ fetchSinaConceptBoards ──→ stock_concepts 表 + concept_index.json
```

### 数据源

| 用途 | API |
|------|-----|
| A 股股票列表 | `vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/Market_Center.getHQNodeData` |
| 日线实时行情 | 同上（分页 200 只/页） |
| 历史 K 线 | `ifzq.gtimg.cn/appstock/app/fqkline/get`（腾讯财经） |
| 概念板块树 | `vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/Market_Center.getHQNodes` |

### 数据库表

- `stock_list` — 股票代码/名称/拼音映射
- `stock_daily_YYYYMMDD` — 每日快照（分表）
- `stock_history` — 连续日线（code + date 唯一）
- `stock_concepts` — 股票-概念映射（`symbol + concept` 唯一）

### 输出产物

- `concept_index.json` — 概念→成分股 索引（由 `concept.exe` 读取）
