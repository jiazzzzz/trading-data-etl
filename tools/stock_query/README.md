# stock_query — 技术面筛选工具

查询 SQLite 中的 `stock_history` 表，按区间涨跌幅或均线多头排列筛选股票。

## 用法

```
query.exe -db PATH -gain PCT [-days N] [-top N] [-from YYYY-MM-DD] [-to YYYY-MM-DD]
query.exe -db PATH -ma [-ma-periods P1,P2,...] [-to YYYY-MM-DD] [-top N]
```

## 目标

快速定位区间强势股（涨跌幅）或均线多头排列股票（趋势股）。

## 模式

### -gain 涨跌幅模式

指定区间 `-from ~ -to`（或 `-days N`），计算首日收盘到末日收盘的涨跌幅，筛选 >= PCT% 的股票。

```
query.exe -db jia-stk.db -gain 7 -days 5         # 近5日涨幅>=7%
query.exe -db jia-stk.db -gain -5 -from 2026-06-01 -to 2026-07-04  # 区间跌幅>=5%
```

### -ma 均线多头排列模式

计算每个股票指定周期的 SMA，检查是否短周期 > 长周期（如 MA5 > MA10 > MA20 > MA60）。

```
query.exe -db jia-stk.db -ma                                   # 默认 MA5>10>20>60
query.exe -db jia-stk.db -ma -ma-periods 10,20,60,120 -to 2026-07-01
```

## 架构

单文件 `main.go`，内联所有逻辑：

```
main() → 解析参数 → runGainQuery() / runMAQuery()
  ↓
query stock_history → 分组计算 → 排序筛选 → printResults()
```

### 依赖

- 数据库：`stock_history` 表（由 `update.exe` 维护）
- 不依赖任何网络 API
