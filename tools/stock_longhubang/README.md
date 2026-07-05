# stock_longhubang — 龙虎榜机构买入查询

实时查询龙虎榜中机构席位的买卖数据，默认过滤大盘股和 ST 股。

## 用法

```
longhubang.exe [-date YYYY-MM-DD] [-top N] [-maxcap N] [-mincap N] [-all] [-skiprisk]
```

## 目标

快速发现机构资金持续加仓的中小盘标的，过滤掉大盘股（机构互倒意义不大）。

## 默认过滤逻辑

| 条件 | 默认值 | 说明 |
|------|--------|------|
| 机构净买入 > 0 | 开启 | 只看机构买入的，排除纯粹卖出 |
| 排除 ST/*ST/退市 | 开启 | `-skiprisk` 跳过 |
| 流通市值上限 | 200亿 | `-maxcap 0` 不限制 |
| 排序 | 净买额降序 | |
| 数量限制 | 前30只 | `-top N` 调整 |

## 输出

```
序号  代码      名称      涨跌幅    净买额    机构买  机构卖  流通市值
1    688669  聚石化学  +6.57%  3.35亿    2      2     136亿
```

## 架构

单文件 `main.go`，约 295 行。

```
main()
  ├─ guessLatestTradeDay()  → 自动计算最近交易日（周末/盘前回退）
  ├─ fetchLHB()             → 东方财富 API 分页拉取
  │   └─ RPT_ORGANIZATION_TRADE_DETAILS report
  ├─ 过滤: isST / net buy / market cap
  ├─ sort by NetBuyAmt DESC
  └─ printResult()
```

### 数据源

| 用途 | API |
|------|-----|
| 机构交易明细 | `datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_ORGANIZATION_TRADE_DETAILS` |

### API 字段

| 字段 | 说明 |
|------|------|
| SECURITY_CODE | 股票代码 |
| SECURITY_NAME_ABBR | 股票名称 |
| NET_BUY_AMT | 机构净买入额 |
| BUY_TIMES / SELL_TIMES | 买入/卖出机构家数 |
| FREECAP | 流通市值（亿） |
| CHANGE_RATE | 涨跌幅 |
| EXPLANATION | 上榜原因 |

### 特点

- 无数据库依赖
- 自动分页（pageSize=100）
- 交易日自动判断（周末退回周五，盘前退回前一日）
