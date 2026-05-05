# 贡献指南

感谢你对维尼量子节点项目的关注！我们欢迎各种形式的贡献。

## 如何贡献

### 报告Bug

如果你发现了bug，请：
1. 检查 [Issues](../../issues) 确认问题尚未被报告
2. 使用Bug Report模板创建新issue
3. 提供详细的复现步骤、环境信息和日志

### 建议新功能

如果你有好的想法，请：
1. 先在 [Discussions](../../discussions) 中讨论
2. 使用Feature Request模板创建issue
3. 清晰描述功能的价值和使用场景

### 提交代码

1. **Fork仓库**
   ```bash
   # Fork到你的GitHub账号
   # 然后克隆到本地
   git clone https://github.com/your-username/weini-quantum-proxy.git
   cd weini-quantum-proxy
   ```

2. **创建分支**
   ```bash
   git checkout -b feature/your-feature-name
   # 或
   git checkout -b fix/your-bug-fix
   ```

3. **编写代码**
   - 遵循项目的代码风格
   - 添加必要的测试
   - 更新相关文档

4. **提交变更**
   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   ```
   
   提交信息格式：
   - `feat:` 新功能
   - `fix:` bug修复
   - `docs:` 文档更新
   - `style:` 代码格式（不影响功能）
   - `refactor:` 重构
   - `test:` 测试相关
   - `chore:` 构建/工具链相关

5. **推送并创建PR**
   ```bash
   git push origin feature/your-feature-name
   ```
   然后在GitHub上创建Pull Request

## 开发环境设置

### 前置要求

- Go 1.21+
- Python 3.9+
- Node.js 16+ (用于前端开发)
- Wails v2 (用于桌面应用)

### 安装依赖

```bash
# Go依赖
go mod download

# Python依赖
pip install -r skills/research/weini-proxy/requirements.txt

# Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 本地运行

```bash
# 开发模式运行Wails应用
cd cmd/proxy-node-studio-wails
wails dev

# 运行测试
go test ./...

# 运行节点抓取脚本
python3 skills/research/weini-proxy/scripts/ss_crawler.py
```

### 构建

```bash
# 构建Wails桌面应用
./scripts/build_proxy_node_studio_wails_windows.sh

# 或使用Wails命令
cd cmd/proxy-node-studio-wails
wails build
```

## 代码规范

### Go代码

- 使用 `gofmt` 格式化代码
- 使用 `golangci-lint` 检查代码质量
- 遵循 [Effective Go](https://golang.org/doc/effective_go.html)
- 添加必要的注释，特别是导出的函数和类型

### Python代码

- 遵循 PEP 8 规范
- 使用 `black` 格式化代码
- 使用类型提示（type hints）

### 前端代码

- 使用ES6+语法
- 保持代码简洁易读
- 添加必要的注释

## 测试

- 为新功能编写单元测试
- 确保所有测试通过：`go test ./...`
- 测试覆盖率应保持在合理水平

## 文档

- 更新README.md（如果需要）
- 为新功能添加文档到 `docs/` 目录
- 保持代码注释的准确性

## Pull Request流程

1. 确保你的代码通过所有测试
2. 更新相关文档
3. 清晰描述你的变更
4. 链接相关的issue
5. 等待review和反馈
6. 根据反馈进行修改

## 审查标准

我们会检查：
- 代码质量和风格
- 测试覆盖率
- 文档完整性
- 是否引入破坏性变更
- 性能影响

## 行为准则

- 尊重所有贡献者
- 接受建设性批评
- 专注于项目的最佳利益
- 保持专业和友好

## 获取帮助

如果有任何问题：
- 查看 [文档](docs/)
- 在 [Discussions](../../discussions) 提问
- 创建 [Issue](../../issues)

## 许可证

通过贡献代码，你同意你的贡献将基于MIT许可证发布。

---

感谢你的贡献！🎉
