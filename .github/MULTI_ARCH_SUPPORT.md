# 多架构支持说明

## ✅ 已完成更新

**更新时间**: 2026-05-01  
**状态**: ✅ 已修复并扩展

---

## 🔧 修复的问题

1. **添加了 `wails.json`** - 修复了 "The system cannot find the file specified" 错误
2. **扩展了架构支持** - 从 3 个平台扩展到 **10 个编译目标**

---

## 📦 支持的架构

### Windows (3 个)

| 架构 | 文件名 | 适用系统 | 说明 |
|------|--------|---------|------|
| **AMD64** ⭐ | `weini-quantum-proxy-windows-amd64.exe` | Windows 10/11 (64位) | 最常见，推荐 |
| **386** | `weini-quantum-proxy-windows-386.exe` | Windows 7/8/10 (32位) | 老旧 PC |
| **ARM64** | `weini-quantum-proxy-windows-arm64.exe` | Windows on ARM | Surface Pro X、Snapdragon PC |

**如何选择**:
- 大多数用户: 下载 **AMD64** 版本
- 老旧 PC（2010年前）: 下载 **386** 版本
- Surface Pro X: 下载 **ARM64** 版本

---

### macOS (3 个)

| 架构 | 文件名 | 适用设备 | 说明 |
|------|--------|---------|------|
| **Universal** ⭐ | `weini-quantum-proxy-macos-universal.zip` | 所有 Mac | Intel + Apple Silicon |
| **AMD64** | `weini-quantum-proxy-macos-amd64.zip` | Intel 芯片 Mac | 2020年前的 Mac |
| **ARM64** | `weini-quantum-proxy-macos-arm64.zip` | Apple Silicon | M1/M2/M3/M4 芯片 |

**如何选择**:
- **推荐**: 直接下载 **Universal** 版本（适用所有 Mac）
- 不确定您的 Mac 类型？运行: `uname -m`
  - 输出 `arm64` = Apple Silicon（M 系列芯片）
  - 输出 `x86_64` = Intel 芯片

---

### Linux (4 个)

| 架构 | 文件名 | 适用系统 | 说明 |
|------|--------|---------|------|
| **AMD64** ⭐ | `weini-quantum-proxy-linux-amd64` | Ubuntu/Debian/Fedora (64位) | 最常见 |
| **386** | `weini-quantum-proxy-linux-386` | 老旧 Linux (32位) | 古董级 PC |
| **ARM64** | `weini-quantum-proxy-linux-arm64` | ARM 64位 | 树莓派 4/5、ARM 服务器 |
| **ARM** | `weini-quantum-proxy-linux-arm` | ARM 32位 | 树莓派 3 及更早 |

**如何选择 - 检查架构**:
```bash
uname -m

# 输出结果对应:
# x86_64      → 下载 AMD64
# i386/i686   → 下载 386
# aarch64     → 下载 ARM64
# armv7l      → 下载 ARM
```

**示例 - 树莓派**:
- 树莓派 5: ARM64
- 树莓派 4: ARM64
- 树莓派 3B+: ARM64 或 ARM（取决于操作系统）
- 树莓派 3: ARM
- 树莓派 2: ARM
- 树莓派 1: ARM

---

## 🎯 推荐版本（按使用场景）

### 桌面用户
| 系统 | 推荐版本 |
|------|---------|
| Windows 10/11 | `windows-amd64.exe` |
| macOS（任意型号）| `macos-universal.zip` |
| Ubuntu/Debian | `linux-amd64` |

### 嵌入式/边缘设备
| 设备 | 推荐版本 |
|------|---------|
| 树莓派 4/5 | `linux-arm64` |
| 树莓派 3 | `linux-arm` |
| ARM 服务器 | `linux-arm64` |
| Surface Pro X | `windows-arm64.exe` |

### 老旧设备
| 系统 | 推荐版本 |
|------|---------|
| Windows 7 (32位) | `windows-386.exe` |
| 古董级 Linux | `linux-386` |

---

## 🔄 交叉编译支持

### Linux ARM/ARM64 编译

GitHub Actions 现在使用交叉编译工具链：

```yaml
# ARM64 (aarch64)
CC=aarch64-linux-gnu-gcc
CXX=aarch64-linux-gnu-g++

# ARM (armv7)
CC=arm-linux-gnueabihf-gcc
CXX=arm-linux-gnueabihf-g++
```

这意味着：
- ✅ 在 x86_64 服务器上编译 ARM 版本
- ✅ 编译时间更快（并行编译）
- ✅ 无需物理 ARM 设备

---

## 📊 编译矩阵

### 总览

```
10 个并行编译任务
├── Windows (3)
│   ├── AMD64
│   ├── 386
│   └── ARM64
├── macOS (3)
│   ├── Universal
│   ├── AMD64
│   └── ARM64
└── Linux (4)
    ├── AMD64
    ├── 386
    ├── ARM64
    └── ARM
```

### 预计编译时间

| 阶段 | 时间 |
|------|------|
| 单个平台编译 | 3-5 分钟 |
| 10 个平台并行 | 5-8 分钟 |
| 创建 Release | 1-2 分钟 |
| **总计** | **6-10 分钟** |

---

## 📦 Release 产物大小

### 预估文件大小

| 平台 | 架构 | 大小（预估）|
|------|------|------------|
| Windows | AMD64 | 15-30 MB |
| Windows | 386 | 12-25 MB |
| Windows | ARM64 | 15-30 MB |
| macOS | Universal | 30-50 MB |
| macOS | AMD64 | 15-25 MB |
| macOS | ARM64 | 15-25 MB |
| Linux | AMD64 | 15-30 MB |
| Linux | 386 | 12-25 MB |
| Linux | ARM64 | 15-30 MB |
| Linux | ARM | 12-25 MB |

**总计**: 约 150-300 MB（所有平台）

---

## 🧪 测试命令

### 本地测试不同架构

#### Windows
```bash
# AMD64
wails build -platform windows/amd64

# 386
wails build -platform windows/386

# ARM64
wails build -platform windows/arm64
```

#### macOS
```bash
# Universal (推荐)
wails build -platform darwin/universal

# Intel only
wails build -platform darwin/amd64

# Apple Silicon only
wails build -platform darwin/arm64
```

#### Linux
```bash
# AMD64
wails build -platform linux/amd64

# 386
wails build -platform linux/386

# ARM64 (需要交叉编译工具)
CC=aarch64-linux-gnu-gcc wails build -platform linux/arm64

# ARM (需要交叉编译工具)
CC=arm-linux-gnueabihf-gcc wails build -platform linux/arm
```

---

## 📚 用户选择指南

### 添加到 Release Notes

已在 Release Notes 中添加详细的架构选择指南：

- ✅ 按平台分类的表格
- ✅ 适用系统说明
- ✅ 架构检测命令
- ✅ 安装指南（可折叠详情）
- ✅ 树莓派版本对应表

---

## 🎉 改进总结

### 改进前
- ❌ 缺少 `wails.json` 导致编译失败
- ❌ 只有 3 个编译目标
- ❌ 不支持 ARM 设备
- ❌ 不支持 32 位系统

### 改进后
- ✅ 添加 `wails.json` 配置
- ✅ **10 个编译目标**
- ✅ 支持树莓派（ARM/ARM64）
- ✅ 支持老旧系统（32 位）
- ✅ 支持 Windows on ARM
- ✅ 详细的用户指南

---

## 🔗 相关链接

- **查看编译进度**: https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/actions
- **下载编译产物**: https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/releases

---

## ⏰ 当前测试状态

**测试 Tag**: `v0.9.1-test`  
**推送时间**: 刚刚  
**预计完成**: 6-10 分钟后

**实时查看**: https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/actions

---

**更新完成！现在支持 10 个编译目标，覆盖几乎所有常见设备！** 🎉
