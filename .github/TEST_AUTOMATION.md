# 快速测试 GitHub Actions 自动编译

## ✅ 已完成的配置

已成功配置 GitHub Actions 自动编译系统，支持：
- ✅ Windows AMD64
- ✅ macOS Universal (Intel + Apple Silicon)
- ✅ Linux AMD64 (Ubuntu)

---

## 🧪 测试步骤

### 方法 1：创建测试版本（推荐）

```bash
# 1. 确保所有更改已提交
cd /home/robot/dark/new_weini/weini-securenet-release
git status

# 2. 创建测试 tag
git tag -a v0.9.1-test -m "Test automated build system"

# 3. 推送 tag 到 GitHub
git push origin v0.9.1-test

# 4. 查看 GitHub Actions 运行状态
# 访问: https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/actions
```

### 方法 2：通过 GitHub 网页界面测试

1. **访问 Actions 页面**
   ```
   https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/actions
   ```

2. **手动触发 Workflow**
   - 点击左侧 "Build and Release"
   - 点击右侧 "Run workflow" 按钮
   - 选择 branch: `main`
   - 点击绿色 "Run workflow" 按钮

3. **监控进度**
   - 页面会显示 3 个并行任务：
     - Build Windows AMD64
     - Build macOS Universal
     - Build Linux AMD64
   - 每个任务约需 5-10 分钟

4. **查看结果**
   - 编译成功后，会创建一个 Draft Release
   - 访问: https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/releases

---

## 📊 预期结果

### 1. GitHub Actions 页面

应该看到：
```
✅ Build Windows AMD64 - Success
✅ Build macOS Universal - Success  
✅ Build Linux AMD64 - Success
✅ Release - Success (仅 tag 推送时)
```

### 2. Release 页面

应该生成 3 个文件：
```
📦 weini-quantum-proxy-windows-amd64.exe
📦 weini-quantum-proxy-macos-universal.zip
📦 weini-quantum-proxy-linux-amd64
```

### 3. 文件大小预估

- Windows .exe: ~15-30 MB
- macOS .zip: ~20-40 MB
- Linux binary: ~15-30 MB

---

## 🐛 如果遇到问题

### 问题 1：Workflow 未触发

**检查**：
```bash
# 确认 tag 已推送
git ls-remote --tags origin

# 查看最近的 tags
git tag -l

# 如果 tag 未推送，重新推送
git push origin v0.9.1-test
```

### 问题 2：编译失败

**常见原因**：
1. Go 模块依赖问题
2. Wails 配置问题
3. 缺少必要的编译工具

**解决方案**：
- 查看具体的错误日志
- 访问 Actions 页面，点击失败的任务查看详细日志
- 在本地尝试编译：`make build-linux`（或其他平台）

### 问题 3：Release 未创建

**原因**：
- 只有推送 `v*` 格式的 tag 才会创建 Release
- 手动触发不会创建 Release（仅编译）

**解决方案**：
```bash
# 创建正确格式的 tag
git tag -a v0.9.1 -m "Release v0.9.1"
git push origin v0.9.1
```

---

## 🎉 测试成功后

### 清理测试 tag（可选）

```bash
# 删除本地 tag
git tag -d v0.9.1-test

# 删除远程 tag
git push origin :refs/tags/v0.9.1-test

# 删除测试 Release（在 GitHub 网页操作）
# 访问: https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/releases
# 找到测试版本，点击 Delete
```

### 创建正式版本

```bash
# 1. 更新 CHANGELOG.md
vim CHANGELOG.md

# 2. 提交更改
git add CHANGELOG.md
git commit -m "准备发布 v1.0.0"
git push

# 3. 创建正式 tag
git tag -a v1.0.0 -m "Release version 1.0.0 - First Stable Release"
git push origin v1.0.0

# 4. GitHub Actions 自动编译和发布
# 访问: https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/releases/latest
```

---

## 📝 版本发布检查清单

在发布新版本前，确保：

- [ ] 所有代码已提交并推送
- [ ] `CHANGELOG.md` 已更新
- [ ] 版本号遵循语义化版本规范（v1.0.0）
- [ ] 本地测试通过：`make test`
- [ ] 本地编译成功：`make build-linux`（或当前平台）
- [ ] README 和文档已更新
- [ ] 创建 tag：`git tag -a vX.Y.Z -m "Release vX.Y.Z"`
- [ ] 推送 tag：`git push origin vX.Y.Z`
- [ ] 监控 GitHub Actions 执行状态
- [ ] 验证 Release 页面的文件完整性

---

## 🔗 相关链接

- **Actions 页面**: https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/actions
- **Releases 页面**: https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/releases
- **详细文档**: `.github/WORKFLOWS_GUIDE.md`
- **Makefile 命令**: 查看 `Makefile` 文件

---

## 💡 提示

1. **首次测试**建议使用 `v0.9.1-test` 这样的测试标签
2. **正式发布**使用 `v1.0.0`、`v1.0.1` 这样的标准格式
3. **预发布版**可使用 `v1.0.0-beta`、`v1.0.0-rc.1` 格式
4. 每次推送 tag 都会触发完整的编译流程，请谨慎操作

---

**准备就绪！现在可以推送一个测试 tag 来验证自动编译系统。** 🚀
