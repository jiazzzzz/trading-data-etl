# stock_concept — 概念成分股查询工具

按概念/题材名称查询成分股，支持多层数据源回退。

## 用法

```
concept.exe -name CONCEPT_NAME [-db DB_PATH]
concept.exe -list -db DB_PATH
```

## 目标

知道概念名（如"白酒""OLED"），找全成分股列表。

## 数据源优先级（自动回退）

```
搜索概念名 → 东方财富 suggest API
  ↓
[1] concept_index.json（最快）       ← update.exe -update-concepts 生成
  ↓ 未命中
[2] stock_concepts 表               ← 同上，数据库版
  ↓ 未命中
[3] concept_stock_relation 表       ← concept_relation.exe 爬取
  ↓ 未命中
[4] 本地缓存 (concept_cache.json)    ← 上次扫描结果
  ↓ 未命中
[5] igu888.com 逐股扫描（最慢, ~5min）← 全 A 股遍历匹配
```

### 链路特点

- [1]~[4] 都需要 `-db` 参数指向 SQLite
- [5] 是纯网络扫描，不需要数据库
- 当概念在东方财富索引中不存在（如"路桥交通"），[3] 的表仍然能找到

### 概念关联描述

如果数据库中有 `concept_stock_relation` 表，每只股票的**概念关联说明**会显示在 `关联:` 行（只显示前 80 字）。

## 架构

单文件 `main.go`，约 714 行：

```
main()
  ├─ searchConcept()       → 东方财富 suggest API → ConceptBoard
  ├─ readConceptIndex()    → JSON 索引精确/模糊匹配
  ├─ queryDB()             → stock_concepts 表 SQL 查询
  ├─ queryRelationDB()     → concept_stock_relation 表回退
  │   └─ findRelationConceptName()  → 模糊名称匹配
  ├─ checkCache()          → 本地 JSON 缓存
  ├─ scanConcept()         → igu888.com 逐股扫描（并发 workers）
  │   └─ stockHasConcept() → 正则匹配概念板块标签
  ├─ enrichDescriptions()  → 补充业务描述
  └─ printResult()         → 格式化输出
```

### 输出增强

- **含 `-db` 时**: 显示概念说明 + 每只股票的业务关联描述
- **无 `-db` 时**: 纯代码/名称列表

### 数据源

| 用途 | API |
|------|-----|
| 概念搜索 | `searchadapter.eastmoney.com/api/suggest/get?type=14` |
| 股票概念扫描 | `igu888.com/ticai/{code}.html` |
| 全 A 股列表 | 新浪行情 API（分页） |
