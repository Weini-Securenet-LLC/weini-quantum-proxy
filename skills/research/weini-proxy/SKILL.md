---
name: weini-proxy
description: 从 GitHub 聚合免费代理订阅，解析 ss/vmess/vless/trojan，兼容 Clash YAML，递归发现 provider URL，做 TCP 可达性验证与 Geo/ASN 丰富化，并导出 xlsx/csv/json/历史差异/告警报告。
version: 5.0
---

# weini-proxy v5

`weini-proxy` 是一个面向 agent 的免费代理节点聚合 skill。

它会从 GitHub 抓取公开订阅内容，解析多协议代理 URI，兼容 Clash YAML 代理列表，递归发现 provider URL，做基础 TCP 连通性验证与 Geo/ASN 丰富化，并生成结构化报告、历史快照、变化 delta 与国家/ASN 告警结果。

适合作为：
- 免费节点情报收集器
- 第一轮存活筛选器
- 自动化巡检器
- 供其他 agent 消费的标准化节点输入层
- 免费代理源变化监控器
- 国家/ASN 变化预警器

## v5 相比 v4 的升级

### 新增能力 1：provider 递归抓取策略
v4 只是在文本里发现 provider URL 后继续抓取；v5 进一步把它升级成**有深度控制的递归抓取**。

现在你可以控制：
- `--max-provider-depth`
- `--max-discovered-urls`

每个仓库都会统计：
- 抓了多少普通 URL
- 抓了多少 provider URL
- 最深递归到第几层

### 新增能力 2：更细粒度去重指纹
v4 的去重主要基于：
- protocol
- host
- port
- credential
- network
- tls

v5 会进一步把下列字段纳入指纹：
- `method`
- `sni`
- `path`
- `host_header`
- `alpn`

这可以减少不同链路参数却被误判成同一节点的问题。

### 新增能力 3：国家 / ASN 变化告警
v5 除了输出普通 delta 之外，还会额外计算：
- 国家分布的 top gains / top losses
- ASN 分布的 top gains / top losses

也就是说，除了知道“节点总数变了多少”，你还可以知道：
- 哪些国家节点明显变多/变少
- 哪些 ASN 明显变多/变少

### 新增能力 4：告警 sheet
XLSX 新增：
- **变化告警**

用于查看：
- `country_gain`
- `country_loss`
- `asn_gain`
- `asn_loss`

## 文件结构

### Skill 文档
```text
/root/.hermes/skills/research/weini-proxy/SKILL.md
```

### 主脚本
```text
/root/.hermes/skills/research/weini-proxy/scripts/ss_crawler.py
```

### 默认源配置
```text
/root/.hermes/skills/research/weini-proxy/references/default_sources.json
```

### 默认历史目录
```text
/root/.hermes/skills/research/weini-proxy/data/history
```

## 支持协议
- `ss://`
- `vmess://`
- `vless://`
- `trojan://`

## 快速使用

### 默认运行
```bash
PYTHONUNBUFFERED=1 python3 /root/.hermes/skills/research/weini-proxy/scripts/ss_crawler.py
```

### 快速冒烟测试
```bash
PYTHONUNBUFFERED=1 python3 /root/.hermes/skills/research/weini-proxy/scripts/ss_crawler.py \
  --skip-verify \
  --skip-enrich \
  --max-nodes 100
```

### 控制 provider 递归深度
```bash
PYTHONUNBUFFERED=1 python3 /root/.hermes/skills/research/weini-proxy/scripts/ss_crawler.py \
  --max-provider-depth 3 \
  --max-discovered-urls 30
```

### 禁用 provider 自动发现
```bash
PYTHONUNBUFFERED=1 python3 /root/.hermes/skills/research/weini-proxy/scripts/ss_crawler.py \
  --skip-provider-discovery
```

### 禁用 Geo / ASN 丰富化
```bash
PYTHONUNBUFFERED=1 python3 /root/.hermes/skills/research/weini-proxy/scripts/ss_crawler.py \
  --skip-enrich
```

### 指定历史目录
```bash
PYTHONUNBUFFERED=1 python3 /root/.hermes/skills/research/weini-proxy/scripts/ss_crawler.py \
  --history-dir /tmp/weini-history
```

### 自定义源文件
```bash
PYTHONUNBUFFERED=1 python3 /root/.hermes/skills/research/weini-proxy/scripts/ss_crawler.py \
  --source-file /path/to/my_sources.json
```

### 完整验证模式
```bash
PYTHONUNBUFFERED=1 python3 /root/.hermes/skills/research/weini-proxy/scripts/ss_crawler.py \
  --timeout 5 \
  --workers 100 \
  --output weini_proxy.xlsx
```

## 命令行参数

| 参数 | 作用 |
|---|---|
| `--skip-verify` | 跳过 TCP 验证，只聚合解析 |
| `--skip-enrich` | 跳过 Geo / ASN 丰富化 |
| `--skip-provider-discovery` | 跳过 provider URL 自动发现 |
| `--timeout 5` | 设置连接超时秒数 |
| `--workers 100` | 设置并发线程数 |
| `--output out.xlsx` | 指定 xlsx 输出路径 |
| `--json-output out.json` | 指定 JSON 摘要路径 |
| `--csv-output out.csv` | 指定 CSV 明细路径 |
| `--delta-output out.delta.json` | 指定 delta 报告路径 |
| `--source-file file.json` | 指定源配置文件 |
| `--protocols ss,vmess` | 指定协议过滤范围 |
| `--max-nodes 500` | 限制采集节点数 |
| `--history-dir /path/dir` | 指定历史快照目录 |
| `--no-history` | 禁用历史快照与对比 |
| `--max-discovered-urls 20` | 每个仓库最多追加发现多少 provider URL |
| `--max-provider-depth 2` | provider URL 递归发现最大深度 |

## 输出内容

v5 默认会生成 4 类结果文件：

1. **XLSX**：主报告
2. **CSV**：全部节点明细
3. **JSON**：摘要统计
4. **DELTA JSON**：和上一份历史快照的差异

如未禁用历史，还会额外写入历史目录：
- `summary_YYYYMMDD_HHMMSS.json`
- `nodes_YYYYMMDD_HHMMSS.json`

## XLSX 工作表
1. **可达节点**：仅 TCP 可达节点，含 IP / 国家 / ASN / provider 深度
2. **全部节点**：全部节点，含失败原因与地理信息
3. **统计**：总体统计 + 协议统计 + 来源仓库统计 + 历史变化摘要
4. **地理分布**：国家分布 + ASN 分布
5. **变化告警**：国家/ASN 的 top gains / top losses
6. **原始URI**：可达节点的原始 URI

## JSON 摘要内容
主 `.json` 中包含：
- 生成时间
- 总节点数
- 可达节点数
- 可达率
- 已丰富化节点数
- 协议级统计
- 来源仓库统计
- 国家分布统计
- ASN 分布统计
- ISP 分布统计
- provider 深度统计
- 最快前 20 个节点
- 历史差异摘要
- 国家/ASN 告警变化

### JSON 示例结构
```json
{
  "generated_at_utc": "2026-04-22 11:00:00 UTC",
  "total_nodes": 1200,
  "reachable_nodes": 260,
  "reachable_rate": 21.7,
  "enriched_nodes": 1180,
  "protocol_stats": [],
  "source_stats": [],
  "country_stats": [],
  "asn_stats": [],
  "isp_stats": [],
  "provider_depth_stats": [],
  "alert_delta": {
    "country": {
      "top_gains": [],
      "top_losses": []
    },
    "asn": {
      "top_gains": [],
      "top_losses": []
    }
  }
}
```

## 默认源配置格式

`default_sources.json` 中每个源对象格式如下：

```json
{
  "repo": "owner/repo",
  "branches": ["main", "master"],
  "paths": ["sub/sub_merge.txt", "sub/base64/mix"],
  "include_readme": true
}
```

## 实现原理

### 1. 采集层
- 遍历源配置文件中的 GitHub 仓库
- 拼接 raw URL
- 获取订阅文件与 README 内容
- 同时尝试原文解析和 base64 解码后解析

### 2. provider 递归发现层
对于每个抓到的文本，v5 会继续尝试发现：
- Clash `proxy-providers` 的 `url`
- 其它 `providers` 的 `url`
- README 中疑似订阅 URL / provider URL

并以队列方式递归抓取，同时记录：
- 深度
- 父 URL
- provider chain

### 3. 解析层
v5 会从两种内容中提取节点：

#### A. 裸 URI
- `ss://`
- `vmess://`
- `vless://`
- `trojan://`

#### B. Clash YAML
如果内容可被识别为 Clash YAML，并且含有 `proxies:` 列表，则会尝试把其中的 `ss/vmess/vless/trojan` 代理项转换为 URI 再统一处理。

### 4. 标准化层
尽量统一保留：
- `protocol`
- `name`
- `host`
- `port`
- `method`
- `credential`
- `network`
- `tls`
- `source_detail`
- `raw_uri`
- `source_repo`
- `source_url`
- `provider_depth`
- `provider_chain`
- `resolved_ip`
- `country`
- `asn`
- `isp`
- `fingerprint`

### 5. 去重层
v5 使用更细粒度指纹，综合：
```text
protocol + host + port + credential + network + tls + method + sni + path + host_header + alpn
```

### 6. 验证层
- 使用 `ThreadPoolExecutor` 并发 TCP 验证
- 支持 IPv4 / IPv6 解析
- 记录：
  - `reachable`
  - `latency_ms`
  - `fail_reason`
  - `resolved_ip`

### 7. Geo / ASN 丰富化层
- 对缺失 IP 的节点再做一次解析
- 批量查询 IP 对应的地理与 ASN 信息
- 回填到节点字段中

### 8. 历史层
如果未指定 `--no-history`：
- 先读取历史目录中最近一份 `summary_*.json`
- 与当前结果做集合比较
- 得出新增/移除/新增可达/移除可达
- 再比较国家与 ASN 分布，生成告警增减榜
- 最后把本次 summary 和节点明细落盘

## 其他 agent 如何使用这个 skill

### 标准工作流
1. 加载 `weini-proxy`
2. 读取 skill 文档
3. 运行脚本
4. 优先读取 `.json` 和 `.delta.json`
5. 若需要国家/ASN 视角，则读取 `country_stats` / `asn_stats`
6. 若需要变化告警，则读取 `alert_delta`
7. 如需人工查看，再打开 `.xlsx`

### 推荐调用提示词
```text
加载 weini-proxy skill，使用默认源文件抓取免费代理节点，解析 ss/vmess/vless/trojan，并兼容 Clash YAML，启用 provider URL 递归发现，做 TCP 可达性验证与 Geo/ASN 丰富化，导出 xlsx/csv/json/delta，并总结总节点数、各协议可达率、来源仓库贡献、provider 深度、国家分布、ASN 分布，以及相对上一轮的新增/移除节点数和国家/ASN 告警变化。
```

### 快速样本提示词
```text
加载 weini-proxy skill，用 --skip-verify --skip-enrich --max-nodes 100 跑一个快速样本，检查解析是否正常，并告诉我协议分布、来源仓库贡献、provider 深度以及发现了多少 provider URL。
```

### 巡检任务提示词
```text
加载 weini-proxy skill，运行完整版验证并启用 history，对比上一轮结果，告诉我新增节点、移除节点、新增可达节点、移除可达节点、国家分布变化、ASN 分布变化，以及 top gains/top losses 告警。
```

## 实战建议

### 建议 1：先跑轻量冒烟测试
```bash
PYTHONUNBUFFERED=1 python3 /root/.hermes/skills/research/weini-proxy/scripts/ss_crawler.py --skip-verify --skip-enrich --max-nodes 100
```
确认：
- 源文件有效
- 裸 URI 提取正常
- Clash YAML 解析正常
- provider 递归发现工作正常
- 输出文件正常生成

### 建议 2：限制递归深度
如果担心抓取范围膨胀，可以先设：
```bash
--max-provider-depth 1 --max-discovered-urls 10
```

### 建议 3：让其他 agent 优先消费 `.json` 和 `.delta.json`
因为这些文件最适合自动化读取，不需要先解析 Excel。

## Pitfalls
- **TCP 可达 ≠ 真实代理可用**，仍然只是第一轮粗筛
- provider 递归发现会扩大抓取范围，深度过高时更容易引入失效源、重复内容和抓取成本上升
- Clash YAML 字段在不同项目里差异很大，少数字段组合可能无法完整还原为 URI
- Geo / ASN 丰富化依赖外部 IP 信息服务，偶尔会失败或返回不完整结果
- `vmess` / `vless` / `trojan` 的 TCP 可达，不代表 TLS、WS、Reality、gRPC 参数都正确
- 如果后续要接真实协议可用性验证器（例如项目内的 `cmd/nodevalidate`），不要在 Linux 上误传 Windows 版 `sing-box.exe`，否则会出现 `exec format error`，导致所有候选都验证失败、最终 `list.json` 变成 0 节点
- 推荐让验证器按当前系统自动选择 sing-box 二进制，或显式传入匹配当前 OS 的路径；Linux 通常应使用 `sing-box`，Windows 使用 `sing-box.exe`
- sing-box 发布资产在不同平台格式不同：Windows 常见是 `.zip`，Linux 常见是 `.tar.gz`；自动下载脚本不能只匹配 `.zip`
- 免费源变化极快，仓库路径和格式会随时变化
- 更细粒度去重虽然更准确，但也可能把细微参数差异视为不同节点，导致总数上升
- 若后续接 `cmd/nodevalidate` 一类真实协议验证器，要**等待验证进程真正退出后**再消费输出 JSON；验证器可能在中途反复覆盖 `/tmp/usable_nodes.json` 之类的结果文件，进程尚未结束时读到的 `usable_count` / `usable_nodes` 可能只是暂存态，最终数量还会变化
- 因此自动化流水线不要仅凭“结果文件已出现”或“JSON 可解析”就继续；应以验证进程退出成功作为完成信号，再生成最终 `list.json`
- 后台运行 Python 时建议加 `PYTHONUNBUFFERED=1`

## 依赖
- Python 3.10+
- curl
- openpyxl（缺失时自动安装）
- PyYAML（缺失时自动安装）

## 验证标准
运行完成后，至少确认：
1. 生成了 `.xlsx`、`.csv`、`.json`、`.delta.json`
2. JSON 中 `total_nodes > 0`
3. JSON 中存在 `country_stats` / `asn_stats` / `source_stats` / `alert_delta`
4. XLSX 中有“变化告警”sheet
5. CSV 中存在 `provider_depth / provider_chain / fingerprint / resolved_ip / country / asn`
6. 如未禁用 history，则历史目录中出现新的 `summary_*.json` 与 `nodes_*.json`

## v5 的边界
当前 v5 仍然是：
- 聚合
- provider 递归发现
- 多协议解析
- Clash YAML 转换
- 基础 TCP 验证
- Geo / ASN 丰富化
- 历史增量比较
- 国家/ASN 变化告警

它仍然**不是**完整协议握手验证器。

如果以后做 v6，建议升级方向：
- vmess/vless/trojan 的真实协议握手验证
- 告警阈值配置与自动通知
- 更深入的 provider 图谱分析
- 历史趋势图与日报
- 节点质量评分模型
