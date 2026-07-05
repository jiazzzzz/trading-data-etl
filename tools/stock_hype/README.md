# stock_hype — 炒作点分析工具

整合行情、概念题材、新闻动态、资本运作事件，快速判断个股炒作逻辑。

## 用法

```
hype.exe -code STOCK_CODE [-region REGION]
```

## 目标

在推荐或分析个股时，快速了解"这只股票炒什么"。

## 输出

### 基础信息

- 价格 / 涨跌幅
- 总市值
- 市盈率
- 所属板块（地区）

### 概念题材

从 igu888.com 抓取"所属板块"标签，例如：`OLED` `白酒` `专精特新` 等

### 近期动态

近期新闻标题（最多 8 条），格式：`[2026-07-03] 新闻内容`

### 资本运作（潜在催化点）

近期公告事件，如：增发、关联交易、资产重组、股权激励等（最多 5 条）

## 架构

单文件 `main.go`：

```
腾讯行情 API  ──→ fetchFromTencent() ──┐
东方财富 push2  ──→ fetchFromPush2()  ──┤→ BasicInfo
                                         │
igu888.com/ticai/  ──→ fetchConcepts() ──┤→ 概念标签
igu888.com/stocknews/ ─→ fetchNews() ────┤→ 新闻 + 事件
```

### 数据源

| 用途 | API |
|------|-----|
| 实时行情 | `qt.gtimg.cn/q=sh600519`（腾讯财经） |
| 行情备用 | `push2.eastmoney.com/api/qt/stock/get` |
| 概念题材 | `igu888.com/ticai/{code}.html` |
| 新闻/事件 | `igu888.com/stocknews/{code}.html` |

### 编码处理

- 腾讯 API 返回 GBK 编码 → 使用 `golang.org/x/text/encoding/simplifiedchinese` 解码
- 爱股网页面可能为 GB2312 → 自动检测 UTF-8/GBK

### 特点

- 无数据库依赖
- `-region` 参数可手动指定板块名（当自动探测失败时备用）
