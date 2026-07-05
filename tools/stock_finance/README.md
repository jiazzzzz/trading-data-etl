# stock_finance — 财务数据 + 风险预警工具

实时从东方财富 API 拉取财务指标和风险公告，无需本地数据库。

## 用法

```
finance.exe -code STOCK_CODE [-top N] [-risk]
```

## 目标

个股排雷：查看财务健康度 + 监管风险。

## 输出

### 财务数据表（-top N 控制期数，默认最近4期）

| 指标 | 说明 |
|------|------|
| EPS | 每股收益 |
| BPS | 每股净资产 |
| ROE | 净资产收益率 |
| 营收/净利润 | 带同比增速 |
| 毛利率/净利率 | |
| 负债率/流动比率 | |
| 经营现金流 | 每股经营现金流 |

### 风险预警

自动扫描最近公告标题，匹配以下风险关键词：

| 级别 | 关键词 |
|------|--------|
| **CRITICAL** | 立案、强制退市、终止上市、行政处罚、财务造假、虚假记载 |
| **WARNING** | 公开谴责、监管警示、资金占用、违规担保、退市风险 |
| **NOTICE** | 风险提示、停牌、整改、调查 |

默认显示摘要（各等级数量）；`-risk` 显示每条公告详情。

```
finance.exe -code 600519          # 财务摘要 + 风险条数
finance.exe -code 600811 -risk    # 财务 + 完整风险明细
```

## 架构

单文件 `main.go`：

```
东方财富 datacenter-web API ──→ fetchFinance()  ──→ 财务数据表
东方财富 np-anotice API     ──→ fetchRiskInfo() ──→ 风险关键词匹配 → 摘要/详情
```

### 数据源

| 用途 | API |
|------|-----|
| 财务数据 | `datacenter-web.eastmoney.com/api/data/v1/get?reportName=RPT_F10_FINANCE_MAINFINADATA` |
| 风险公告 | `np-anotice-stock.eastmoney.com/api/security/ann` |

### 特点

- 无数据库依赖，纯 HTTP 请求
- 支持多种 code 格式：`600519` / `sh600519` / `000001.SZ`
- 风险检测基于标题关键词匹配，非结构化分析
