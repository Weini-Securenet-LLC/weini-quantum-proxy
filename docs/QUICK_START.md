# 🚀 快速开始指南 | Quick Start Guide

[English](#english) | [中文](#中文)

---

## 中文

### 📥 下载安装

#### Windows

1. 访问 [Releases 页面](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/releases)
2. 下载最新版本的 `Weini-Quantum-Proxy-Windows-amd64.exe`
3. 双击运行即可（无需安装）

#### macOS

1. 访问 [Releases 页面](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/releases)
2. 下载 `Weini-Quantum-Proxy-Darwin-universal.dmg`
3. 打开 DMG 文件，拖拽到 Applications 文件夹
4. 首次运行需要在"系统偏好设置 > 安全性与隐私"中允许

#### Linux

```bash
# 下载最新版本
wget https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/releases/latest/download/Weini-Quantum-Proxy-Linux-amd64

# 添加执行权限
chmod +x Weini-Quantum-Proxy-Linux-amd64

# 运行
./Weini-Quantum-Proxy-Linux-amd64
```

---

### ⚡ 5分钟上手

#### 1. 启动应用

首次启动会自动创建配置文件。

#### 2. 添加节点

**方法一：订阅链接**
```
设置 → 订阅管理 → 添加订阅
输入订阅链接 → 保存 → 更新订阅
```

**方法二：手动导入**
```
节点列表 → 添加节点
选择协议类型（SS/SSR/Trojan/V2Ray）
填写服务器信息 → 保存
```

**方法三：AI 自动发现**
```
工具 → AI 节点发现
选择来源 → 开始爬取
验证节点 → 一键导入
```

#### 3. 连接节点

```
节点列表 → 选择节点 → 点击"连接"
状态栏显示"已连接" → 开始使用
```

#### 4. 系统代理设置

**自动配置（推荐）**
```
设置 → 系统代理 → 启用"自动配置系统代理"
```

**手动配置**
- HTTP 代理：`127.0.0.1:1080`
- SOCKS5 代理：`127.0.0.1:1080`

---

### 📱 移动设备配置

#### Android

查看详细教程：[Android 配置指南](MOBILE_ANDROID.md)

**快速步骤：**
1. 安装代理客户端（推荐 v2rayNG、ShadowRocket）
2. 导出节点配置（二维码或剪贴板）
3. 在手机上扫描或导入
4. 连接使用

#### iOS

查看详细教程：[iOS 配置指南](MOBILE_IOS.md)

**快速步骤：**
1. 安装 Shadowrocket 或 Quantumult X
2. 导出节点配置
3. 在 iOS 设备上导入
4. 允许 VPN 配置
5. 连接使用

---

### 🛠️ 常见问题

#### 1. 无法连接节点？

**检查清单：**
- ✅ 节点是否在线（查看延迟测试）
- ✅ 系统代理是否正确设置
- ✅ 防火墙是否允许应用通过
- ✅ 本地端口是否被占用

**解决方法：**
```bash
# 测试节点延迟
节点列表 → 右键节点 → 测试延迟

# 切换端口
设置 → 本地代理 → 修改端口号（如改为 1081）

# 查看日志
帮助 → 查看日志 → 检查错误信息
```

#### 2. 速度很慢？

**优化建议：**
- 切换到延迟更低的节点
- 启用"智能路由"模式（国内直连，国外代理）
- 关闭其他占用带宽的应用
- 尝试不同的协议（Trojan 通常较快）

#### 3. 部分网站无法访问？

**检查路由规则：**
```
设置 → 路由规则 → 切换模式
- 全局模式：所有流量走代理
- 智能模式：根据规则自动选择
- 直连模式：不使用代理
```

#### 4. 应用崩溃或无法启动？

**排查步骤：**
```bash
# 1. 删除配置文件（会清空所有设置）
rm -rf ~/.config/weini-quantum-proxy

# 2. 重新启动应用
# 3. 如果仍然崩溃，提交 Issue 并附上日志
```

---

### 🔧 高级功能

#### 自动切换节点

```
设置 → 自动切换
启用"失败自动切换"
设置超时时间（如 3 秒）
```

#### 订阅自动更新

```
设置 → 订阅管理
启用"自动更新订阅"
设置更新周期（如每天）
```

#### 规则分流

```
设置 → 路由规则 → 自定义规则
添加域名或IP规则
选择走代理或直连
```

---

### 📚 更多文档

- [架构设计](ARCHITECTURE.md)
- [Android 使用指南](MOBILE_ANDROID.md)
- [iOS 使用指南](MOBILE_IOS.md)
- [AI Skills 开发](SKILL_DEVELOPMENT.md)
- [贡献指南](../CONTRIBUTING.md)

---

### 💬 获取帮助

- 📖 [完整文档](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy)
- 🐛 [报告问题](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/issues)
- 💡 [功能建议](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/discussions)
- 📧 Email: info@weinisecure.net

---

## English

### 📥 Download & Install

#### Windows

1. Visit [Releases page](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/releases)
2. Download `Weini-Quantum-Proxy-Windows-amd64.exe`
3. Double-click to run (no installation required)

#### macOS

1. Visit [Releases page](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/releases)
2. Download `Weini-Quantum-Proxy-Darwin-universal.dmg`
3. Open DMG and drag to Applications folder
4. First run: Allow in System Preferences > Security & Privacy

#### Linux

```bash
# Download latest version
wget https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/releases/latest/download/Weini-Quantum-Proxy-Linux-amd64

# Add execute permission
chmod +x Weini-Quantum-Proxy-Linux-amd64

# Run
./Weini-Quantum-Proxy-Linux-amd64
```

---

### ⚡ 5-Minute Quick Start

#### 1. Launch Application

Configuration files will be created automatically on first launch.

#### 2. Add Nodes

**Method 1: Subscription**
```
Settings → Subscription → Add Subscription
Enter subscription URL → Save → Update
```

**Method 2: Manual Import**
```
Node List → Add Node
Select protocol (SS/SSR/Trojan/V2Ray)
Fill in server details → Save
```

**Method 3: AI Auto-Discovery**
```
Tools → AI Node Discovery
Select source → Start crawling
Validate nodes → Import
```

#### 3. Connect to Node

```
Node List → Select node → Click "Connect"
Status shows "Connected" → Ready to use
```

#### 4. System Proxy Settings

**Auto Configuration (Recommended)**
```
Settings → System Proxy → Enable "Auto Configure"
```

**Manual Configuration**
- HTTP Proxy: `127.0.0.1:1080`
- SOCKS5 Proxy: `127.0.0.1:1080`

---

### 📱 Mobile Device Setup

#### Android

See detailed guide: [Android Configuration](MOBILE_ANDROID.md)

**Quick Steps:**
1. Install proxy client (v2rayNG or ShadowRocket recommended)
2. Export node config (QR code or clipboard)
3. Scan or import on phone
4. Connect and use

#### iOS

See detailed guide: [iOS Configuration](MOBILE_IOS.md)

**Quick Steps:**
1. Install Shadowrocket or Quantumult X
2. Export node configuration
3. Import on iOS device
4. Allow VPN configuration
5. Connect and use

---

### 🛠️ Troubleshooting

#### 1. Cannot Connect?

**Checklist:**
- ✅ Is node online? (Check latency test)
- ✅ Is system proxy configured correctly?
- ✅ Is firewall allowing the app?
- ✅ Is local port occupied?

**Solutions:**
```bash
# Test node latency
Node List → Right-click node → Test Latency

# Change port
Settings → Local Proxy → Change port (e.g., to 1081)

# Check logs
Help → View Logs → Check error messages
```

#### 2. Slow Speed?

**Optimization:**
- Switch to lower latency node
- Enable "Smart Routing" (Direct for domestic, Proxy for international)
- Close bandwidth-heavy applications
- Try different protocols (Trojan is usually faster)

#### 3. Some Sites Not Accessible?

**Check Routing Rules:**
```
Settings → Routing Rules → Switch Mode
- Global: All traffic through proxy
- Smart: Auto-select based on rules
- Direct: No proxy
```

#### 4. App Crashes?

**Troubleshooting:**
```bash
# 1. Delete config (will clear all settings)
rm -rf ~/.config/weini-quantum-proxy

# 2. Restart app
# 3. If still crashes, submit Issue with logs
```

---

### 🔧 Advanced Features

#### Auto-Switch Nodes

```
Settings → Auto Switch
Enable "Auto-switch on failure"
Set timeout (e.g., 3 seconds)
```

#### Auto-Update Subscription

```
Settings → Subscription
Enable "Auto-update subscription"
Set update interval (e.g., daily)
```

#### Rule-Based Routing

```
Settings → Routing Rules → Custom Rules
Add domain or IP rules
Select proxy or direct
```

---

### 📚 More Documentation

- [Architecture Design](ARCHITECTURE.md)
- [Android Guide](MOBILE_ANDROID.md)
- [iOS Guide](MOBILE_IOS.md)
- [AI Skills Development](SKILL_DEVELOPMENT.md)
- [Contributing](../CONTRIBUTING.md)

---

### 💬 Get Help

- 📖 [Full Documentation](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy)
- 🐛 [Report Issues](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/issues)
- 💡 [Feature Requests](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/discussions)
- 📧 Email: info@weinisecure.net

---

**© 2026 Weini Securenet LLC | MIT License**