# Translation Guide | 翻译指南

## 🌍 Supported Languages

We aim to make Weini Quantum Proxy accessible to people worldwide, especially those living under internet censorship.

### Current Translations

- ✅ English (en)
- ✅ 简体中文 (zh-CN)
- ✅ فارسی (fa) - Persian/Farsi
- ✅ العربية (ar) - Arabic
- ✅ Русский (ru) - Russian
- 🔄 Türkçe (tr) - Turkish (In Progress)
- 🔄 မြန်မာ (mm) - Burmese/Myanmar (In Progress)

### Priority Languages

Languages for countries with internet censorship:

| Language | Country | Priority | Status |
|----------|---------|----------|--------|
| فارسی | Iran | 🔴 High | ✅ Done |
| العربية | Multiple Arab countries | 🔴 High | ✅ Done |
| Русский | Russia, Belarus | 🔴 High | ✅ Done |
| Türkçe | Turkey | 🟡 Medium | 🔄 In Progress |
| မြန်မာ | Myanmar | 🟡 Medium | 🔄 In Progress |
| Tiếng Việt | Vietnam | 🟢 Low | ⏳ Planned |
| བོད་ཡིག | Tibet | 🟢 Low | ⏳ Planned |

## 🙋 How to Contribute Translations

### Step 1: Check Existing Translations

1. Check the `i18n/` directory
2. See if your language already has a file
3. If yes, review and improve it
4. If no, create a new one

### Step 2: Create Translation

1. Copy `i18n/README_EN.md` as a template
2. Rename to `README_[LANGUAGE_CODE].md` (e.g., `README_TR.md` for Turkish)
3. Translate all content
4. Keep formatting (Markdown, links, etc.)
5. Update language selector at the top

### Step 3: Submit

1. Fork the repository
2. Create a new branch: `git checkout -b translation/[language]`
3. Add your translation
4. Commit: `git commit -m "feat: add [Language] translation"`
5. Create a Pull Request

## 📋 Translation Checklist

- [ ] README title and subtitle
- [ ] Mission statement
- [ ] Features section
- [ ] Quick start guide
- [ ] Architecture description
- [ ] Important notices
- [ ] All links work correctly
- [ ] Formatting is preserved
- [ ] Language selector is updated

## 🎯 Translation Guidelines

### 1. Tone and Style

- **Empowering**: Focus on freedom and rights
- **Clear**: Simple language, avoid jargon
- **Inclusive**: Neutral gender, accessible to all
- **Respectful**: Sensitive to cultural contexts

### 2. Technical Terms

Keep these in English (or add local equivalent):
- Protocol names: SS, VMess, VLESS, Trojan
- Software names: Wails, sing-box, GitHub
- Technical terms: proxy, node, TCP, etc.

Example:
- English: "Multi-Protocol Support: SS / VMess / VLESS / Trojan"
- Persian: "پشتیبانی از پروتکل‌های متعدد: SS / VMess / VLESS / Trojan"

### 3. Sensitive Content

Be careful when translating:
- Names of censoring countries
- Political statements
- Human rights references

**Principle**: Truth and respect, but prioritize user safety.

### 4. Right-to-Left (RTL) Languages

For Arabic, Persian, etc.:

```html
<div align="center" dir="rtl">
# Content here
</div>
```

## 🌟 Translation Quality

### Good Translation ✅

- Accurate meaning
- Natural flow
- Culturally appropriate
- Technically correct

### Avoid ❌

- Machine translation without review
- Direct word-for-word translation
- Missing cultural context
- Broken formatting

## 💬 Get Help

- Ask in [Discussions](https://github.com/Weini-Securenet-LLC/weini-quantum-proxy/discussions)
- Join our [Telegram](https://t.me/weini_quantum)
- Tag maintainers in your PR

## 🙏 Translators

Special thanks to our translators:

- English: [@weinidaohang](https://github.com/weinidaohang)
- 简体中文: [@weinidaohang](https://github.com/weinidaohang)
- فارسی: [Your name here]
- العربية: [Your name here]
- Русский: [Your name here]

Want to see your name here? Submit a translation!

---

**Together, we make freedom accessible to everyone, in every language.**
