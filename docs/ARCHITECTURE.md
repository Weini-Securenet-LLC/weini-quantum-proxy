# 架构文档

## 系统概览

维尼量子节点是一个多层架构的代理解决方案，包含节点发现、验证、管理和运行时系统。

## 核心组件

### 1. 桌面应用层 (Desktop Application Layer)

使用Wails框架构建的跨平台桌面应用。

```
┌─────────────────────────────────────────┐
│          Wails Desktop App              │
├─────────────────────────────────────────┤
│  Frontend (HTML/CSS/JS)                 │
│  - 用户界面                              │
│  - 节点列表展示                          │
│  - 流量监控显示                          │
│  - 配置管理                              │
├─────────────────────────────────────────┤
│  Backend (Go)                           │
│  - Wails桥接 (internal/wailsapp)        │
│  - 业务逻辑处理                          │
│  - 系统调用封装                          │
└─────────────────────────────────────────┘
```

**技术栈**:
- 前端: 原生HTML/CSS/JavaScript
- 后端: Go 1.21+
- 框架: Wails v2

**代码位置**:
- `cmd/proxy-node-studio-wails/`
- `internal/wailsapp/`

---

### 2. 节点发现系统 (Node Discovery System)

负责从多个源聚合代理节点信息。

```
┌─────────────────────────────────────────┐
│       Node Discovery Engine             │
├─────────────────────────────────────────┤
│  GitHub Subscription Aggregator         │
│  ├─ Fetch subscription URLs             │
│  ├─ Parse Clash YAML                    │
│  ├─ Extract ss/vmess/vless/trojan       │
│  └─ Recursive provider discovery        │
├─────────────────────────────────────────┤
│  Node Parser (internal/proxynode)       │
│  ├─ URI解析                              │
│  ├─ 协议识别                             │
│  └─ 标准化处理                           │
└─────────────────────────────────────────┘
```

**功能**:
1. 从GitHub聚合免费代理订阅
2. 解析多种格式 (Clash YAML, 原始URI等)
3. 递归发现provider引用
4. 支持的协议: SS, VMess, VLESS, Trojan

**代码位置**:
- `skills/research/weini-proxy/`
- `internal/proxynode/`

---

### 3. 节点验证系统 (Node Validation System)

对发现的节点进行多层验证。

```
┌─────────────────────────────────────────┐
│      Node Validation Engine             │
├─────────────────────────────────────────┤
│  Layer 1: TCP Reachability              │
│  ├─ 端口可达性检测                       │
│  ├─ 超时控制                             │
│  └─ 并发验证                             │
├─────────────────────────────────────────┤
│  Layer 2: Protocol Validation           │
│  ├─ 使用sing-box测试                     │
│  ├─ 真实协议握手                         │
│  ├─ HTTP请求测试                         │
│  └─ 延迟测量                             │
├─────────────────────────────────────────┤
│  Layer 3: Geo/ASN Enrichment            │
│  ├─ IP地理位置查询                       │
│  ├─ ASN信息获取                          │
│  └─ 元数据丰富化                         │
└─────────────────────────────────────────┘
```

**验证流程**:
1. TCP可达性 → 快速筛选
2. 协议验证 → 真实可用性
3. 地理信息 → 数据丰富化

**代码位置**:
- `cmd/nodevalidate/`
- `skills/research/weini-proxy/scripts/ss_crawler.py`

---

### 4. 代理运行时 (Proxy Runtime)

基于sing-box的代理执行引擎。

```
┌─────────────────────────────────────────┐
│         Proxy Runtime System            │
├─────────────────────────────────────────┤
│  Config Generator                       │
│  ├─ sing-box配置生成                     │
│  ├─ 节点信息转换                         │
│  └─ 路由规则配置                         │
├─────────────────────────────────────────┤
│  Process Manager                        │
│  ├─ sing-box启动/停止                    │
│  ├─ 进程监控                             │
│  ├─ 健康检查                             │
│  └─ 自动重启                             │
├─────────────────────────────────────────┤
│  System Proxy Controller                │
│  ├─ Windows代理设置                      │
│  ├─ macOS代理设置                        │
│  ├─ Linux代理设置                        │
│  └─ 自动恢复机制                         │
├─────────────────────────────────────────┤
│  Traffic Monitor                        │
│  ├─ 实时流量统计                         │
│  ├─ 连接日志                             │
│  └─ 性能监控                             │
└─────────────────────────────────────────┘
```

**代码位置**:
- `internal/globalproxy/`

---

### 5. 民主导航系统 (Democratic Navigation System)

未来规划的去中心化节点发布和评价系统。

```
┌─────────────────────────────────────────┐
│    Democratic Navigation System         │
├─────────────────────────────────────────┤
│  Publishing Platform                    │
│  ├─ 去中心化节点发布                     │
│  ├─ 社区贡献接口                         │
│  └─ 质量标准验证                         │
├─────────────────────────────────────────┤
│  Voting & Rating System                 │
│  ├─ 节点质量投票                         │
│  ├─ 用户评分机制                         │
│  ├─ 信誉系统                             │
│  └─ 防刷票机制                           │
├─────────────────────────────────────────┤
│  Recommendation Engine                  │
│  ├─ AI驱动推荐                           │
│  ├─ 地理位置优化                         │
│  ├─ 用途分类                             │
│  └─ 个性化推荐                           │
└─────────────────────────────────────────┘
```

**状态**: 规划中 (查看 [ROADMAP.md](../ROADMAP.md))

---

## 数据流

### 完整的数据流程

```
1. 节点发现
   GitHub订阅源 → 抓取脚本 → 原始节点列表
   
2. 节点验证
   原始节点 → TCP检测 → 协议验证 → 验证后节点
   
3. 数据丰富化
   验证后节点 → Geo/ASN查询 → 完整节点信息
   
4. 生成list.json
   完整节点信息 → 格式化 → list.json
   
5. 前端展示
   list.json → 前端加载 → 用户选择
   
6. 代理启动
   用户选择 → 生成sing-box配置 → 启动代理 → 设置系统代理
   
7. 运行时监控
   sing-box运行 → 流量日志 → 前端实时显示
```

---

## 关键接口

### Frontend ↔ Backend (Wails Bridge)

```go
// internal/wailsapp/app.go

type App struct {
    ctx context.Context
}

// 获取节点列表
func (a *App) GetNodeList() ([]Node, error)

// 启动全局代理
func (a *App) StartProxy(nodeID string) error

// 停止代理
func (a *App) StopProxy() error

// 获取流量统计
func (a *App) GetTrafficStats() (*TrafficStats, error)
```

### Node Parser

```go
// internal/proxynode/proxynode.go

// 解析节点URI
func ParseNode(uri string) (*Node, error)

// 获取节点列表
func FetchNodeList(url string) ([]*Node, error)
```

### Proxy Runtime

```go
// internal/globalproxy/runtime.go

type Runtime struct {
    singbox *exec.Cmd
    config  *Config
}

// 启动代理
func (r *Runtime) Start() error

// 停止代理
func (r *Runtime) Stop() error

// 获取状态
func (r *Runtime) Status() (*Status, error)
```

---

## 配置管理

### sing-box配置生成

```json
{
  "inbounds": [
    {
      "type": "mixed",
      "listen": "127.0.0.1",
      "listen_port": 7890
    }
  ],
  "outbounds": [
    {
      "type": "shadowsocks",
      "server": "example.com",
      "server_port": 443,
      "method": "aes-256-gcm",
      "password": "password"
    }
  ],
  "route": {
    "rules": [],
    "auto_detect_interface": true
  }
}
```

---

## 错误处理

### 分层错误处理策略

1. **UI层**: 用户友好的错误提示
2. **业务层**: 详细的错误日志
3. **运行时层**: 自动重试和恢复

### 错误类型

- `NodeNotFoundError`: 节点不存在
- `ConnectionError`: 连接失败
- `ProxyStartError`: 代理启动失败
- `SystemProxyError`: 系统代理设置失败

---

## 性能优化

### 并发控制

- 节点验证: Worker pool模式
- TCP检测: 并发+超时控制
- 协议验证: 分批处理

### 缓存策略

- 节点列表缓存: 1小时
- Geo/ASN缓存: 永久
- 配置缓存: 会话级

---

## 安全考虑

1. **代理连接**
   - 仅连接用户选择的节点
   - 支持加密协议
   - 流量不经过我们的服务器

2. **系统权限**
   - 最小权限原则
   - 管理员权限仅用于设置系统代理
   - 退出时清理所有设置

3. **数据隐私**
   - 不收集用户数据
   - 本地存储配置
   - 开源透明

---

## 扩展性

### 插件架构 (规划中)

```
Plugin Interface
├─ Protocol Plugin (新协议支持)
├─ Source Plugin (新订阅源)
├─ Validator Plugin (自定义验证)
└─ UI Plugin (界面扩展)
```

---

## 更多文档

- [用户指南](USER_GUIDE.md)
- [API文档](API.md)
- [开发指南](DEVELOPMENT.md)
