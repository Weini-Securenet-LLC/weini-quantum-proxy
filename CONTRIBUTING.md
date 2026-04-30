# 🤝 贡献指南 | Contributing Guide

感谢你对 Weini Quantum Proxy 的关注！我们欢迎任何形式的贡献。

Thank you for your interest in Weini Quantum Proxy! We welcome all forms of contributions.

[English](#english) | [中文](#中文)

---

## 中文

### 📋 行为准则

参与本项目即表示你同意遵守我们的 [行为准则](CODE_OF_CONDUCT.md)。请友善待人，尊重他人。

### 🚀 如何贡献

#### 1. 报告 Bug

发现 bug？请帮我们修复它！

**提交 Issue 前请检查：**
- ✅ 搜索已有 issues，避免重复
- ✅ 使用最新版本测试
- ✅ 提供详细的复现步骤

**好的 Bug 报告应包含：**
```markdown
**环境信息**
- 操作系统：Windows 11 / macOS 14 / Ubuntu 22.04
- 应用版本：v0.9.0-beta
- 协议类型：Shadowsocks / Trojan / V2Ray

**复现步骤**
1. 打开应用
2. 添加节点
3. 点击连接
4. 观察到错误

**预期行为**
应该成功连接

**实际行为**
连接失败，显示错误：[错误信息]

**截图**
如果适用，添加截图

**日志**
附上应用日志（帮助 → 查看日志）
```

#### 2. 建议新功能

有好点子？我们很乐意听！

**提交功能建议前：**
- ✅ 搜索 issues 和 discussions
- ✅ 明确描述问题和解决方案
- ✅ 考虑对现有用户的影响

**功能建议模板：**
```markdown
**问题描述**
当前的痛点是什么？

**建议的解决方案**
你希望如何改进？

**替代方案**
是否考虑过其他方案？

**使用场景**
谁会使用这个功能？频率如何？
```

#### 3. 提交代码

想要提交代码？太棒了！

**开发流程：**

1. **Fork 仓库**
   ```bash
   # 点击 GitHub 上的 Fork 按钮
   # 克隆你的 fork
   git clone https://github.com/YOUR_USERNAME/weini-quantum-proxy.git
   cd weini-quantum-proxy
   ```

2. **创建分支**
   ```bash
   # 基于 main 创建新分支
   git checkout -b feature/your-feature-name
   
   # 或修复 bug
   git checkout -b fix/bug-description
   ```

3. **开发与测试**
   ```bash
   # 安装依赖
   go mod download
   
   # 运行开发模式
   wails dev
   
   # 运行测试
   go test ./...
   ```

4. **提交更改**
   ```bash
   # 添加文件
   git add .
   
   # 提交（遵循提交规范）
   git commit -m "feat: 添加新功能描述"
   
   # 推送到你的 fork
   git push origin feature/your-feature-name
   ```

5. **创建 Pull Request**
   - 访问你的 fork 页面
   - 点击 "New Pull Request"
   - 填写 PR 模板
   - 等待审核

**提交信息规范：**

遵循 [Conventional Commits](https://www.conventionalcommits.org/):

```
类型(范围): 简短描述

详细描述（可选）

关联 Issue（可选）
```

**类型：**
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 代码重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建/工具更新

**示例：**
```
feat(node): 添加节点自动测速功能

- 实现并发延迟测试
- 添加超时处理
- 更新 UI 显示

Closes #123
```

#### 4. 改进文档

文档和代码一样重要！

**文档贡献：**
- 📝 修正错别字和语法错误
- 🌐 添加或改进翻译
- 📖 完善使用教程
- 💡 添加最佳实践示例

**文档位置：**
- `README.md` - 项目主页
- `docs/` - 详细文档
- `i18n/` - 多语言翻译

#### 5. 翻译

帮助我们支持更多语言！

**当前支持的语言：**
- 🇨🇳 简体中文
- 🇺🇸 English
- 🇮🇷 فارسی (波斯语)
- 🇸🇦 العربية (阿拉伯语)
- 🇷🇺 Русский (俄语)

**需要翻译的内容：**
- README 文件
- 应用界面文本
- 文档和教程

**翻译指南：**
1. 复制 `i18n/README_EN.md` 为新语言
2. 翻译所有文本
3. 保持格式和链接不变
4. 提交 PR

### 📐 代码风格

**Go 代码：**
```bash
# 格式化代码
go fmt ./...

# 静态检查
go vet ./...

# 使用 golangci-lint
golangci-lint run
```

**前端代码：**
```bash
# 格式化
npm run format

# 代码检查
npm run lint
```

**通用规范：**
- ✅ 使用有意义的变量名
- ✅ 添加必要的注释（解释"为什么"，而非"是什么"）
- ✅ 保持函数简洁（单一职责）
- ✅ 编写测试用例
- ✅ 更新相关文档

### 🧪 测试

**运行测试：**
```bash
# 单元测试
go test ./...

# 覆盖率测试
go test -cover ./...

# 集成测试
go test -tags=integration ./...
```

**测试要求：**
- 新功能必须包含测试
- Bug 修复应添加回归测试
- 保持测试覆盖率 > 70%

### 🔐 安全

**发现安全漏洞？**

请**不要**公开提交 issue！

发送邮件至：**security@weinisecure.net**

我们会在 48 小时内回复，并尽快修复问题。

### 📄 许可证

提交代码即表示你同意：
- 你的贡献将采用 [MIT License](LICENSE)
- 你拥有提交内容的版权或有权授权

### 🎯 优先级

**高优先级：**
- 🐛 严重 bug 修复
- 🔒 安全漏洞
- 📱 移动端兼容性
- 🌐 多语言支持

**中优先级：**
- ✨ 新协议支持
- ⚡ 性能优化
- 📖 文档改进

**低优先级：**
- 🎨 UI 美化
- 🔧 代码重构
- 📊 数据统计

### 💬 交流讨论

- 💭 [GitHub Discussions](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/discussions) - 一般讨论
- 🐛 [GitHub Issues](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/issues) - Bug 和功能请求
- 📧 Email: contact@weinisecure.net

### 🙏 致谢

感谢所有贡献者！你们的贡献让互联网自由变得更容易。

---

## English

### 📋 Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). Please be kind and respectful.

### 🚀 How to Contribute

#### 1. Report Bugs

Found a bug? Help us fix it!

**Before submitting:**
- ✅ Search existing issues
- ✅ Test with latest version
- ✅ Provide detailed reproduction steps

**Good bug report includes:**
```markdown
**Environment**
- OS: Windows 11 / macOS 14 / Ubuntu 22.04
- App version: v0.9.0-beta
- Protocol: Shadowsocks / Trojan / V2Ray

**Steps to Reproduce**
1. Open app
2. Add node
3. Click connect
4. Observe error

**Expected Behavior**
Should connect successfully

**Actual Behavior**
Connection failed with error: [error message]

**Screenshots**
If applicable, add screenshots

**Logs**
Attach application logs (Help → View Logs)
```

#### 2. Suggest Features

Have a great idea? We'd love to hear it!

**Before suggesting:**
- ✅ Search issues and discussions
- ✅ Clearly describe the problem and solution
- ✅ Consider impact on existing users

**Feature request template:**
```markdown
**Problem Description**
What's the current pain point?

**Proposed Solution**
How would you like to improve it?

**Alternatives**
Have you considered other approaches?

**Use Cases**
Who will use this feature? How often?
```

#### 3. Submit Code

Want to contribute code? Awesome!

**Development workflow:**

1. **Fork the repository**
   ```bash
   # Click Fork button on GitHub
   # Clone your fork
   git clone https://github.com/YOUR_USERNAME/weini-quantum-proxy.git
   cd weini-quantum-proxy
   ```

2. **Create a branch**
   ```bash
   # Create feature branch from main
   git checkout -b feature/your-feature-name
   
   # Or bug fix
   git checkout -b fix/bug-description
   ```

3. **Develop and test**
   ```bash
   # Install dependencies
   go mod download
   
   # Run dev mode
   wails dev
   
   # Run tests
   go test ./...
   ```

4. **Commit changes**
   ```bash
   # Stage files
   git add .
   
   # Commit (follow commit conventions)
   git commit -m "feat: add new feature description"
   
   # Push to your fork
   git push origin feature/your-feature-name
   ```

5. **Create Pull Request**
   - Visit your fork page
   - Click "New Pull Request"
   - Fill in PR template
   - Wait for review

**Commit message convention:**

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): short description

Detailed description (optional)

Related issues (optional)
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Code formatting
- `refactor`: Code refactoring
- `perf`: Performance
- `test`: Testing
- `chore`: Build/tools

**Example:**
```
feat(node): add automatic node speed testing

- Implement concurrent latency tests
- Add timeout handling
- Update UI display

Closes #123
```

#### 4. Improve Documentation

Documentation is as important as code!

**Documentation contributions:**
- 📝 Fix typos and grammar
- 🌐 Add or improve translations
- 📖 Enhance tutorials
- 💡 Add best practice examples

**Documentation locations:**
- `README.md` - Project homepage
- `docs/` - Detailed docs
- `i18n/` - Translations

#### 5. Translate

Help us support more languages!

**Currently supported:**
- 🇨🇳 简体中文
- 🇺🇸 English
- 🇮🇷 فارسی (Persian)
- 🇸🇦 العربية (Arabic)
- 🇷🇺 Русский (Russian)

**Content to translate:**
- README files
- App UI text
- Documentation and tutorials

**Translation guide:**
1. Copy `i18n/README_EN.md` for new language
2. Translate all text
3. Keep format and links intact
4. Submit PR

### 📐 Code Style

**Go code:**
```bash
# Format code
go fmt ./...

# Static check
go vet ./...

# Use golangci-lint
golangci-lint run
```

**Frontend code:**
```bash
# Format
npm run format

# Lint
npm run lint
```

**General guidelines:**
- ✅ Use meaningful variable names
- ✅ Add necessary comments (explain "why", not "what")
- ✅ Keep functions concise (single responsibility)
- ✅ Write test cases
- ✅ Update related documentation

### 🧪 Testing

**Run tests:**
```bash
# Unit tests
go test ./...

# Coverage
go test -cover ./...

# Integration tests
go test -tags=integration ./...
```

**Test requirements:**
- New features must include tests
- Bug fixes should add regression tests
- Maintain test coverage > 70%

### 🔐 Security

**Found a security vulnerability?**

Please **DO NOT** create a public issue!

Email: **security@weinisecure.net**

We'll respond within 48 hours and fix ASAP.

### 📄 License

By submitting code, you agree that:
- Your contribution will be licensed under [MIT License](LICENSE)
- You own the copyright or have permission to license it

### 🎯 Priority

**High priority:**
- 🐛 Critical bug fixes
- 🔒 Security vulnerabilities
- 📱 Mobile compatibility
- 🌐 Multi-language support

**Medium priority:**
- ✨ New protocol support
- ⚡ Performance optimization
- 📖 Documentation improvements

**Low priority:**
- 🎨 UI enhancements
- 🔧 Code refactoring
- 📊 Analytics

### 💬 Communication

- 💭 [GitHub Discussions](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/discussions) - General discussions
- 🐛 [GitHub Issues](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/issues) - Bugs and features
- 📧 Email: contact@weinisecure.net

### 🙏 Acknowledgments

Thank you to all contributors! Your contributions make internet freedom easier.

---

**© 2026 Weini Securenet LLC | Built with ❤️ for Internet Freedom**