# stock_updater 改造计划

## 目标

将 CLI 改为 `./update.exe -db <path> [-from <YYYY-MM-DD>] [-to <YYYY-MM-DD>]` 形式：

- `-db` **必填**：数据库路径
- `-from` 可选：起始日期，默认 30 天前，格式 `YYYY-MM-DD`
- `-to` 可选：截止日期，默认今天，格式 `YYYY-MM-DD`
- `-workers` 可选：并发数，默认 10

## 改造内容

### 1. `main.go` — CLI 重写

| 当前 | 改为 |
|------|------|
| `--db` 可选, 默认 `../jia-stk.db` | `-db` **必填** |
| `--today` / `--backfill --start YYYYMMDD --end YYYYMMDD` | `-from YYYY-MM-DD -to YYYY-MM-DD` |
| 无默认日期 | `-from` 默认 30 天前, `-to` 默认今天 |
| 入参格式 `YYYYMMDD` | 入参格式 `YYYY-MM-DD` |

移除 `--today`、`--backfill`、`--start`、`--end` 模式，改为统一的 range 模式。

### 2. `updater.go` — 新增统一入口

新增 `UpdateRange(startDate, endDate string)` 方法：
1. **拉取股票列表** → 写入 `stock_list`（复用现有逻辑）
2. **遍历日期范围**，每个交易日：
   - 调用 East Money K-line API 获取全市场数据（按日）
   - 写入 `stock_daily_YYYYMMDD` 表
   - 同时写入 `stock_history` 表
3. 如果 `today` 在范围内，额外调用实时 API `fetchDailyData()` 获取实时快照

保持 `UpdateToday()` 和 `Backfill()` 内部可用，`UpdateRange` 作为统一入口编排。

### 3. `db.go` / `api.go` — 新增辅助

- 新增 `parseDate(s string) (string, error)`：`YYYY-MM-DD` → `YYYYMMDD`
- 新增 `computeDateRange(from, to string) ([]string, error)`：生成交易日列表
- K-line 数据已有 `fetchKLine()` 可复用

### 4. `models.go` — 保持不变

## 数据库 Schema（不变）

```
stock_list
├── ts_code, symbol, name, area, industry, list_date, pinyin

stock_daily_YYYYMMDD
├── symbol, code, name, trade, pricechange, changepercent,
    buy, sell, settlement, open, high, low, volume, amount,
    ticktime, per, pb, mktcap, nmc, turnoverratio, dump_time

stock_history
├── id, stock_code, stock_name, exchange, trade_date,
    open, high, low, close, volume, amount, import_time
```

## 使用示例

```bash
# 更新过去一个月的数据
./update.exe -db test.db

# 更新指定范围
./update.exe -db test.db -from 2025-01-01 -to 2026-07-01

# 仅更新今天
./update.exe -db test.db -from 2026-07-04 -to 2026-07-04
```
