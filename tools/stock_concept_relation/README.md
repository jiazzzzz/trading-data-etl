# stock_concept_relation — 概念关联度爬虫

从 ddx.gubit.cn 爬取全量概念-股票关联数据，包含每只股票在特定概念下的业务描述。

## 用法

```
concept_relation.exe -update -db DB_PATH [-limit N]
```

## 目标

为 `concept.exe` 和 `concept_screener.ps1` 提供丰富的信息：
- 概念说明（concept_desc）
- 个股在该概念下的关联描述（description）

例如"聚石化学"在"磷化工"概念下显示： *"公司主营业务是磷化工产品的研发、生产和销售..."*

## 架构

单文件 `main.go`，约 376 行。

### 爬取流程

```
[1] 获取侧边栏索引
    GET ddx.gubit.cn/gainian/px.php?zf=1
    → 正则解析 <li><a href="./{slug}/">{name}</a></li>
    → 得到 ~778 个概念 slug → name 映射

[2] 并发爬取详情页
    3 并发 goroutine + 100ms 间隔
    GET ddx.gubit.cn/gainian/{slug}/
    → 正则解析:
      - 左侧表格: <tr><td class="gpxh">N</td><td><a>{code}</a></td><td>{name}</td>
      - 右侧 info-box: <td class="info-box"><div class="info"><p>StockName: description</p>
      - 概念说明: <div class="gpintro">...</div>
    → ~84s 完成全部 778 个页面

[3] 写库
    INSERT OR REPLACE INTO concept_stock_relation
    → ~12,000+ 条记录
```

### 数据库表

```sql
CREATE TABLE concept_stock_relation (
    concept_name TEXT NOT NULL,   -- 概念名称
    stock_code   TEXT NOT NULL,   -- 股票代码
    stock_name   TEXT,            -- 股票名称
    description  TEXT,            -- 该概念下的业务关联描述
    concept_desc TEXT,            -- 概念总体说明
    updated_date TEXT,            -- 更新时间
    source       TEXT DEFAULT 'ddx',
    PRIMARY KEY (concept_name, stock_code)
);
```

### 数据源

| 用途 | URL |
|------|-----|
| 概念列表（侧边栏） | `ddx.gubit.cn/gainian/px.php?zf=1` |
| 概念详情页 | `ddx.gubit.cn/gainian/{slug}/` |

### 编码

- ddx.gubit.cn 页面为 GBK 编码 → `simplifiedchinese.GBK.NewDecoder()` 解码

### 进度控制

- `-limit N`：只爬前 N 个概念（用于测试）
- 每 50 个概念输出一次进度条
