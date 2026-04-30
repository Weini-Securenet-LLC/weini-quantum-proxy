# AI Agent Skill 开发指南

## 🤖 什么是AI Agent Skill？

AI Agent Skill是一套**可被AI工具调用的自动化脚本**，让AI编程助手（如Cursor、GitHub Copilot等）能够：

- 🔍 自动发现和聚合代理节点
- ✅ 验证节点的真实可用性
- 📊 分析节点质量和性能
- 🔄 定期更新节点列表
- 📤 分享节点给他人

**核心理念**: 
> 授人以鱼不如授人以渔。开源AI工具，让每个人都能用AI实现通信自由。

---

## 🎯 为什么要开源Skill？

### 1. 赋能个人
让不懂技术的人也能通过AI助手获取节点：
```
用户: "帮我找一些可用的代理节点"
AI: *运行weini-proxy skill* → 返回节点列表
```

### 2. 帮助他人
技术志愿者可以用AI自动化流程，为社区提供节点：
```
开发者: "每天自动更新节点并分享"
AI: *定时运行skill* → 自动发布更新
```

### 3. 知识传播
分享**方法**比分享节点更有价值：
- ✅ 教会别人如何用AI获取节点
- ✅ 持续可用，不依赖单一来源
- ✅ 社区共建，去中心化

### 4. 可持续发展
- 单个节点会失效
- 单个服务器会关闭
- 但开源的方法永远有效

---

## 📦 现有的Skill

### 1. weini-proxy (节点抓取)

**位置**: `skills/research/weini-proxy/`

**功能**:
- 从GitHub聚合免费代理订阅
- 解析多种格式（Clash YAML、原始URI等）
- 支持协议：SS、VMess、VLESS、Trojan
- TCP可达性验证
- Geo/ASN信息丰富化
- 导出多种格式（xlsx、csv、json）

**使用方式**:

```bash
# 基础用法
python3 skills/research/weini-proxy/scripts/ss_crawler.py \
  --output nodes.xlsx

# 完整用法
python3 skills/research/weini-proxy/scripts/ss_crawler.py \
  --output /tmp/nodes.xlsx \
  --csv-output /tmp/nodes.csv \
  --json-output /tmp/summary.json \
  --delta-output /tmp/delta.json
```

**AI Agent使用**:

在Cursor或其他AI工具中：
```
1. 打开Skill文档: skills/research/weini-proxy/SKILL.md
2. 要求AI: "根据这个skill帮我获取节点"
3. AI会自动运行脚本并解析结果
```

---

### 2. proxy-node-studio-list-refresh (节点验证)

**位置**: `skills/research/proxy-node-studio-list-refresh/`

**功能**:
- 调用weini-proxy获取节点
- 下载sing-box进行真实协议验证
- 生成前端可用的list.json
- 自动同步到应用目录

**使用方式**:

```bash
# 完整流程
1. 运行weini-proxy获取节点
2. 使用nodevalidate验证节点
3. 生成list.json
```

**AI Agent使用**:

```
用户: "帮我更新可用的节点列表"
AI: 
  1. 运行weini-proxy抓取节点
  2. 运行nodevalidate验证
  3. 生成list.json
  4. 更新到应用
```

---

## 🛠️ 如何使用Skill

### 方式一：手动运行

```bash
cd skills/research/weini-proxy
python3 scripts/ss_crawler.py --output nodes.xlsx
```

### 方式二：AI助手运行（推荐）

**在Cursor中**:

1. 打开skill目录
2. 阅读`SKILL.md`文档
3. 向AI提问：
   ```
   "根据这个skill帮我获取最新的可用节点，
   并保存到output目录"
   ```
4. AI会自动：
   - 理解skill功能
   - 运行相关脚本
   - 解析并展示结果

**在GitHub Copilot中**:

1. 创建新的Python脚本
2. 在注释中说明需求：
   ```python
   # 使用weini-proxy skill获取节点
   # 验证可用性
   # 导出为JSON格式
   ```
3. Copilot会自动生成调用代码

### 方式三：集成到自己的项目

```python
import subprocess
import json

# 调用weini-proxy
result = subprocess.run([
    'python3',
    'skills/research/weini-proxy/scripts/ss_crawler.py',
    '--json-output', 'nodes.json'
], capture_output=True)

# 解析结果
with open('nodes.json', 'r') as f:
    nodes = json.load(f)

# 使用节点数据
for node in nodes:
    print(f"Node: {node['server']}:{node['port']}")
```

---

## 📝 开发新的Skill

### Skill结构

```
skills/research/your-skill-name/
├── SKILL.md              # Skill说明文档（必需）
├── scripts/              # 脚本目录
│   └── main_script.py   # 主脚本
├── references/           # 参考资料
│   └── config.json      # 配置文件
└── README.md            # 详细说明（可选）
```

### SKILL.md模板

```markdown
# Your Skill Name

## 功能描述
简要说明这个skill的功能和用途

## 使用场景
- 场景1
- 场景2

## 使用方法

### 基础用法
\`\`\`bash
python3 scripts/main_script.py
\`\`\`

### 高级用法
\`\`\`bash
python3 scripts/main_script.py --option value
\`\`\`

## 输出格式
说明输出的数据格式

## 依赖
- Python 3.9+
- requests
- ...

## AI Agent指引
告诉AI如何使用这个skill
```

### 开发checklist

- [ ] 创建SKILL.md文档
- [ ] 脚本支持命令行参数
- [ ] 输出格式标准化（JSON/CSV/XLSX）
- [ ] 错误处理完善
- [ ] 添加使用示例
- [ ] 编写测试用例
- [ ] 添加AI Agent使用说明

---

## 🌟 Skill开发最佳实践

### 1. 输出标准化

**推荐使用JSON格式**:
```python
import json

output = {
    "success": True,
    "data": [...],
    "metadata": {
        "count": 100,
        "timestamp": "2026-04-30T12:00:00Z"
    }
}

with open('output.json', 'w') as f:
    json.dump(output, f, indent=2)
```

### 2. 错误处理

```python
try:
    # 主要逻辑
    result = process_data()
except Exception as e:
    print(json.dumps({
        "success": False,
        "error": str(e)
    }))
    sys.exit(1)
```

### 3. 进度显示

```python
from tqdm import tqdm

for item in tqdm(items, desc="Processing"):
    process(item)
```

### 4. 配置文件

```python
import json

with open('config.json', 'r') as f:
    config = json.load(f)

# 使用配置
api_key = config.get('api_key')
```

### 5. 日志记录

```python
import logging

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)

logging.info("Starting process...")
```

---

## 💡 Skill创意

我们欢迎以下类型的Skill贡献：

### 节点相关
- 🔍 更多订阅源聚合
- ✅ 更智能的节点验证
- 📊 节点质量分析和排序
- 🗺️ 节点地理位置可视化
- 📈 节点历史数据追踪

### 自动化工具
- 🔄 定时自动更新
- 📤 自动分享到社交平台
- 💬 Telegram Bot集成
- 📧 邮件订阅推送

### 数据分析
- 📊 节点可用率统计
- 🌍 全球节点分布分析
- ⚡ 速度和延迟测试
- 🔒 安全性评估

### 用户工具
- 🎨 生成分享二维码
- 📱 移动端配置生成
- 🔗 订阅链接转换
- 📝 配置文件生成器

---

## 🤝 贡献Skill

### 提交流程

1. **Fork仓库**
   ```bash
   git clone https://github.com/your-username/weini-quantum-proxy.git
   ```

2. **创建skill分支**
   ```bash
   git checkout -b skill/your-skill-name
   ```

3. **开发skill**
   - 创建skill目录结构
   - 编写SKILL.md
   - 实现功能脚本
   - 添加测试

4. **测试**
   ```bash
   # 手动测试
   python3 skills/research/your-skill/scripts/main.py
   
   # AI Agent测试
   # 在Cursor中让AI运行你的skill
   ```

5. **提交PR**
   ```bash
   git add skills/research/your-skill/
   git commit -m "feat: add [skill-name] skill for [purpose]"
   git push origin skill/your-skill-name
   ```

6. **PR描述**
   ```markdown
   ## Skill名称
   Your Skill Name
   
   ## 功能
   简要描述功能
   
   ## 使用场景
   说明适用场景
   
   ## 测试
   - [ ] 手动测试通过
   - [ ] AI Agent测试通过
   - [ ] 文档完整
   ```

---

## 🔄 Skill维护

### 定期更新

- 📅 每月检查依赖版本
- 🐛 修复发现的bug
- 📝 更新文档
- ✨ 添加新功能

### 社区反馈

- 💬 及时回复issue
- 🙋 帮助用户解决问题
- 📊 收集改进建议
- 🎉 采纳有价值的PR

---

## 📚 学习资源

### Python自动化
- [Python官方文档](https://docs.python.org/3/)
- [Requests库](https://requests.readthedocs.io/)
- [BeautifulSoup](https://www.crummy.com/software/BeautifulSoup/bs4/doc/)

### AI Agent开发
- [Cursor Documentation](https://cursor.sh/docs)
- [GitHub Copilot Guide](https://github.com/features/copilot)
- [LangChain](https://python.langchain.com/)

### 代理协议
- [Shadowsocks](https://shadowsocks.org/)
- [V2Ray](https://www.v2ray.com/)
- [sing-box](https://sing-box.sagernet.org/)

---

## 🆘 获取帮助

开发Skill遇到问题？

1. 📖 查看现有Skill的实现
2. 💬 在[Discussions](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/discussions)提问
3. 🐛 在[Issues](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/issues)报告bug
4. 💌 加入[Telegram群组](https://t.me/weini_quantum)交流

---

## 🎉 让AI为自由服务

通过开源AI Skill，我们可以：

- 🤖 让AI助手帮助每个人获取节点
- 🌐 去中心化节点获取方式
- 📚 传播知识而不仅是节点
- 🤝 构建可持续的自由互联网生态

**一起用AI打破信息封锁！**

---

**相关文档**:
- [贡献指南](../community/CONTRIBUTING.md)
- [项目架构](ARCHITECTURE.md)
- [用户指南](USER_GUIDE.md)
