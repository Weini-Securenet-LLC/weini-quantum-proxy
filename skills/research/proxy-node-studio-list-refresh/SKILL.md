---
name: proxy-node-studio-list-refresh
description: 使用 weini-proxy 抓取免费代理节点，下载当前平台可用的 sing-box，运行 cmd/nodevalidate 做真实协议可用性验证，并生成/同步 proxy-node-studio 两个前端的 list.json。
---

# proxy-node-studio-list-refresh

适用场景：
- 需要刷新 `proxy-node-studio` / `proxy-node-studio-wails` 的 `web/list.json`
- 需要把第一轮 crawl 结果再经过 `cmd/nodevalidate` 做真实协议可用性筛选
- 需要在 Linux 上自动选择正确的 `sing-box` 运行时，避免 `exec format error`

## 前提
- 仓库工作目录：`/root/.hermes/hermes-agent`
- 先加载 `weini-proxy` skill
- 运行 Python 前优先激活项目 venv：
  ```bash
  source venv/bin/activate
  ```

## 标准步骤

### 1. 运行 weini-proxy 抓取与 TCP 初筛
```bash
source venv/bin/activate && \
PYTHONUNBUFFERED=1 python3 /root/.hermes/skills/research/weini-proxy/scripts/ss_crawler.py \
  --output /tmp/weini_proxy.xlsx \
  --csv-output /tmp/weini_nodes.csv \
  --json-output /tmp/weini_summary.json \
  --delta-output /tmp/weini_delta.json
```

期望产物：
- `/tmp/weini_proxy.xlsx`
- `/tmp/weini_nodes.csv`
- `/tmp/weini_summary.json`
- `/tmp/weini_delta.json`

### 2. 下载当前系统匹配的 sing-box 运行时
```bash
source venv/bin/activate && python3 scripts/fetch_sing_box.py --os linux --arch amd64 --output-dir dist/runtime
```

期望输出：
- `dist/runtime/sing-box`

### 3. 运行真实协议可用性验证
```bash
go run ./cmd/nodevalidate \
  -csv /tmp/weini_nodes.csv \
  -out /tmp/usable_nodes.json \
  -workdir /tmp/nodevalidate-work \
  -workers 8 \
  -per-protocol 120
```

关键要求：
- **不要传 `-binary`**，让 `nodevalidate` 自动按当前系统选择正确的 sing-box
- 必须等待进程真正退出后再读取 `/tmp/usable_nodes.json`
- 不要因为文件“已经出现”就提前消费，结果文件在验证过程中可能被反复覆盖

### 4. 生成 list.json
从：
- `/tmp/weini_summary.json`
- `/tmp/usable_nodes.json`

生成目标 JSON，字段要求：
- `generated_from = "weini-proxy skill + protocol usability validation"`
- `source_csv = "/tmp/weini_nodes.csv"`
- `validation_sources = ["/tmp/usable_nodes.json"]`
- `supported_protocols = ["ss","vmess","vless","trojan"]`
- `total_nodes = usable_nodes 的数量`
- `protocol_counts = 按 usable_nodes 里的 URI scheme 重新统计`
- `notes` 至少包含：
  - `crawl_total_nodes`
  - `crawl_reachable_nodes`
  - `crawl_reachable_rate`
  - `usable_validation_candidates`
  - `validated_count`
  - `validation_method`
- `nodes = usable_nodes 数组`

推荐 Python 生成逻辑：
```python
import json
from collections import Counter
from pathlib import Path

summary = json.loads(Path('/tmp/weini_summary.json').read_text())
usable = json.loads(Path('/tmp/usable_nodes.json').read_text())
usable_nodes = usable.get('usable_nodes', [])

supported_protocols = ['ss', 'vmess', 'vless', 'trojan']
counter = Counter()
for uri in usable_nodes:
    proto = uri.split('://', 1)[0].lower() if '://' in uri else 'unknown'
    counter[proto] += 1

# 这个工作流的 list.json 更稳妥的做法是始终写出全部支持协议键，
# 再按 usable_nodes 实际结果填值；缺失协议补 0。
# 这样即使某协议本轮 usable 为 0，前端/下游仍能拿到稳定 schema。
protocol_counts = {proto: counter.get(proto, 0) for proto in supported_protocols}

list_data = {
    'generated_from': 'weini-proxy skill + protocol usability validation',
    'source_csv': '/tmp/weini_nodes.csv',
    'validation_sources': ['/tmp/usable_nodes.json'],
    'supported_protocols': supported_protocols,
    'total_nodes': len(usable_nodes),
    'protocol_counts': protocol_counts,
    'notes': {
        'crawl_total_nodes': summary.get('total_nodes', 0),
        'crawl_reachable_nodes': summary.get('reachable_nodes', 0),
        'crawl_reachable_rate': summary.get('reachable_rate', 0),
        'usable_validation_candidates': usable.get('candidate_count', 0),
        'validated_count': usable.get('validated_count', 0),
        'validation_method': 'go run ./cmd/nodevalidate -csv /tmp/weini_nodes.csv -out /tmp/usable_nodes.json -workdir /tmp/nodevalidate-work -workers 8 -per-protocol 120'
    },
    'nodes': usable_nodes,
}
```

### 5. 同步写入两个前端目标
- `/root/.hermes/hermes-agent/cmd/proxy-node-studio/web/list.json`
- `/root/.hermes/hermes-agent/cmd/proxy-node-studio-wails/web/list.json`

## 验证

### 结果完整性
确认：
1. `/tmp/weini_summary.json` 中 `total_nodes > 0`
2. `/tmp/usable_nodes.json` 中存在：
   - `candidate_count`
   - `validated_count`
   - `usable_count`
   - `usable_nodes`
3. 两个 `list.json` 存在且内容一致
4. `list.json.total_nodes == len(list.json.nodes)`
5. `list.json.protocol_counts` 来自 `usable_nodes` 实际 URI 统计，而不是直接照抄候选计数

### 读取 delta 的正确方式
`/tmp/weini_delta.json` 的增量字段在 `history_delta` 子对象下，不在顶层。读取时优先使用：
- `history_delta.new_nodes`
- `history_delta.removed_nodes`
- `history_delta.new_reachable_nodes`
- `history_delta.removed_reachable_nodes`

不要误读成顶层的 `added_nodes` / `removed_nodes` 一类字段，否则会得到 `null` 或漏报变化。顶层常见字段是：
- `current_total_nodes`
- `current_reachable_nodes`
- `protocol_stats`
- `country_stats`
- `asn_stats`
- `alert_delta`

### 一致性校验
建议使用：
```bash
sha256sum \
  /root/.hermes/hermes-agent/cmd/proxy-node-studio/web/list.json \
  /root/.hermes/hermes-agent/cmd/proxy-node-studio-wails/web/list.json
```

## 汇报模板
至少汇报：
- crawl total
- crawl reachable
- validated candidates
- validated count
- usable total
- protocol counts

如果 `usable total == 0`：
- 不要静默成功
- 必须总结 top validation error
- 优先查看 `/tmp/usable_nodes.json` 的失败明细；必要时结合 `cmd/nodevalidate` 输出日志归纳失败原因
- **不要直接按完整 error 字符串做 top N**；很多 error 会内嵌整段 sing-box 日志、时间戳和连接 ID，导致同类失败被拆成大量唯一值。应先归一化为错误类别再统计，例如：`timeout`、`http_403`、`tls_not_tls`、`tls_handshake_failure`、`connection_refused`、`network_unreachable`、`other`

## Pitfalls
- `weini-proxy` 的 `reachable` 只是 TCP 可达，不等于真实协议可用
- `nodevalidate` 阶段必须等待进程退出后再读结果文件
- 在 Hermes 里后台跑 `nodevalidate` 时，进程可能长时间**没有任何 stdout/stderr**，直到结束才输出汇总；这不代表卡死。应使用 `process wait/poll` 循环按进程状态等待，不要因为“暂时没日志”就提前判失败或读取中间产物
- Linux 下不要误传 Windows 的 `sing-box.exe` 给验证器，否则容易触发 `exec format error`
- 即使 `usable_nodes.json.protocol_counts` 表示每协议尝试了多少候选，真正写入 `list.json.protocol_counts` 时仍应按 `usable_nodes` 最终结果重新统计
- `protocol_counts` 是否要补齐 0 值，取决于调用方要求：若用户只要求“按 usable_nodes 实际 URI 统计”，就按实际出现的协议输出；若前端/下游明确依赖固定 schema，再补齐 `ss/vmess/vless/trojan` 四个键
- `generated_at_utc` 不是本工作流的必需字段；若用户给了精确 schema，就不要额外添加未要求字段
- 如果 `vmess` 候选很多但最终 `usable` 为 0，要检查 `/tmp/usable_nodes.json.detailed_results` 是否出现类似 `alter_id: json: cannot unmarshal string` 的 `sing-box` 配置解码错误；这类失败说明某些 `vmess` 节点字段类型不兼容当前验证链路，汇报 top validation error 时应把它们归并成同一类，而不是按完整原始日志逐条统计
