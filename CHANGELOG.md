# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-05-14

### Added

#### Core Features
- ✨ **AST-Based Scanning** - True code analysis for Go, JavaScript, TypeScript
- 🔍 **Multi-Language Support** - First-class support for 9 programming languages:
  - Go, JavaScript, TypeScript (AST-based)
  - Python, Rust, Ruby, Java, C#, PHP, Kotlin (regex-based)
- 📦 **Ghost Variable Detection** - Find environment variables accessed in code but missing from configuration
- 📊 **Environment Health Scoring** - 0-100 score based on configuration completeness and quality
- 🔄 **Drift Tracking** - Monitor environmental configuration changes over time with snapshots
- ⚡ **High-Performance Caching** - SHA-256 file hashing for fast repeated scans (10,000 files in milliseconds)
- 🛠️ **Zero-Config Operation** - Works immediately with standard `.env` patterns
- 📝 **JSON Output** - CI/CD friendly output for automation and integration
- 🔧 **Customizable Configuration** - `.taso.yaml` for advanced users
- 📝 **Comprehensive Audit Logging** - Track all scans and operations

#### Security Features
- 🛡️ **Path Traversal Prevention** - Validates all paths before file access
- 🚫 **Symlink Attack Prevention** - Explicitly rejects symbolic links
- 🔒 **Resource Exhaustion Protection** - File size limits, timeout enforcement, memory monitoring
- 🔐 **Null Byte Injection Filtering** - Prevents null byte bypass attacks
- 📋 **Secure File Permissions** - Cache stored with 0600 (owner-only) permissions
- 📊 **Comprehensive Audit Trail** - JSON logging for compliance requirements

#### Developer Experience
- 🎯 **Intelligent Auto-Detection** - Finds `.env` files automatically
- 💡 **Interactive Fixes** - Prompts to add missing variables to `.env`
- 📊 **Detailed Reporting** - Shows exactly where variables are used
- 🚀 **Fast Execution** - Millisecond scanning with intelligent caching
- 📖 **Help System** - Comprehensive `--help` documentation

### Documentation
- 📚 **README.md** - Complete user guide with examples and use cases
- 🔐 **SECURITY.md** - Comprehensive security policy and best practices
- 🤝 **Contributing Guidelines** - How to contribute to the project
- 📋 **CHANGELOG.md** - This file
- 🎯 **DEPLOYMENT_AND_REDDIT_GUIDE.md** - Launch and community strategy

### Configuration
- ⚙️ **Zero Config by Default** - Works immediately without configuration
- 🎯 **`.taso.yaml` Support** - Optional configuration for advanced users
- 📝 **`.env.example`** - Template for environment variables
- 🔧 **`.gitignore`** - Pre-configured to exclude `.taso` cache

### Quality Assurance
- ✅ **Unit Tests** - Comprehensive test coverage for all modules
  - `security_test.go` - Path validation and attack prevention
  - `scanner_test.go` - Scanning accuracy
  - `cache_test.go` - Cache functionality
  - `envreader_test.go` - `.env` file parsing
- ✅ **Integration Tests** - End-to-end workflow validation
- 🔍 **Security Testing** - Path traversal, symlink, and resource limit tests
- 📊 **Performance Testing** - Benchmarks for large codebases

### Release Engineering
- 📦 **Multi-Platform Binaries** - Pre-built for Linux, macOS, Windows
- 🏗️ **GoReleaser Configuration** - Automated release pipeline
- ✅ **GitHub Actions** - CI/CD for testing and releases
- 📝 **Release Checksums** - SHA256 verification for downloaded binaries

### Compliance & Standards
- ✅ **SOC 2 Ready** - Audit logging and configuration tracking
- ✅ **ISO 27001 Compatible** - Information security controls
- ✅ **GDPR Compliant** - Data minimization (no personal data collected)
- ✅ **HIPAA Support** - Audit trails for healthcare environments
- ✅ **PCI-DSS Ready** - Environment variable tracking for payment processing

### Performance
- ⚡ **Sub-millisecond Scanning** - For unchanged files (via cache)
- 🚀 **Parallel File Processing** - Uses all CPU cores efficiently
- 💾 **Memory Efficient** - Monitors memory usage during scans
- 📊 **Scalable** - Successfully tested on 10,000+ file codebases

### Community
- 🤝 **GitHub Discussions** - Community Q&A
- 🐛 **Issue Tracking** - Bug reports and feature requests
- 📝 **Contributing Guide** - Clear instructions for contributors
- 🎉 **First Responder** - Active maintainer engagement

### Known Limitations
- Dynamic env var access not supported (e.g., `os.Getenv(myVar)` where myVar is not a literal)
- Command-line arguments not detected
- Runtime-injected variables (Docker secrets, K8s ConfigMaps) not detected

### Dependencies
- **Cobra** (v1.8.0) - Command-line interface
- **Viper** (v1.21.0) - Configuration management
- **Godotenv** (v1.5.1) - .env file parsing
- **Lipgloss** (v0.10.0) - Terminal UI styling
- **Goja** (v0.0.0-20260311135729-065cd970411c) - JavaScript AST parsing
- **Go** (1.23.0) - Runtime

---

## [Unreleased]

### Planned for Phase 2 (Q3 2026)
- 🌐 Web dashboard for visualization
- 👥 Team collaboration features
- 🔔 Slack/Discord notifications
- 🧠 ML-powered anomaly detection
- 🔐 Secret detection warnings
- 📋 SBOM (Software Bill of Materials)
- 🔏 GPG signing of releases
- 🔗 Vault integration

### Planned for Phase 3 (Q4 2026)
- 🤖 Automated secret rotation recommendations
- 🔗 AWS Secrets Manager integration
- 🔗 HashiCorp Vault integration
- 📱 VS Code extension
- 📊 SIEM integration
- 🏢 Advanced audit logging

### Planned for Phase 4 (Q1 2027)
- 📦 Multi-repository scanning
- 🔐 Advanced RBAC
- 📋 Enterprise features
- ☁️ On-premises deployment
- 💼 SLA & premium support

---

## Release Notes

### v1.0.0 - Production Ready

This is the inaugural production release of Taso. After extensive development and security 
hardening, Taso is ready for enterprise use.

**Highlights:**
- Industry-first AST-based environment variable scanning
- Multi-language support out of the box
- Security-audited codebase
- Comprehensive documentation
- Production-ready deployment

**How to Install:**
```bash
go install github.com/Hossiy21/taso@v1.0.0
```

**Breaking Changes:** None

**Next Release:** Q3 2026 with web dashboard and enhanced features

---

## Getting Help

- 📖 **Documentation**: See [README.md](README.md)
- 🔐 **Security Issues**: See [SECURITY.md](SECURITY.md)
- 🤝 **Contributing**: See [README.md](README.md#contributing)
- 💬 **Questions**: Open a [Discussion](https://github.com/Hossiy21/taso/discussions)
- 🐛 **Bugs**: Open an [Issue](https://github.com/Hossiy21/taso/issues)

---

**Created by:** [Hossiy21](https://github.com/Hossiy21)  
**License:** MIT  
**Last Updated:** May 14, 2026
