# 🛸 Taso

**The industry-standard for Environment Variable Drift Detection.**

Find what's silently wrong with your environment — before production does. Taso bridges the gap between your source code and your configuration.

[![Go Report Card](https://goreportcard.com/badge/github.com/Hossiy21/taso)](https://goreportcard.com/report/github.com/Hossiy21/taso)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

<!-- ## 🎬 Demo -->
<!-- ![Taso Demo](./taso-demo.gif) -->

---

## ⚡ Why Taso?

Most tools only compare `.env` files to other `.env` files. **Taso is different.**

Taso analyzes your **actual source code** using **AST (Abstract Syntax Tree)** and optimized scanning to find every environment variable your app *actually* tries to access. It then cross-checks these against your configuration.

If you call `os.Getenv("STRIPE_SECRET")` in your code, but forgot to add it to your `.env` or production secrets, **Taso catches it instantly.**

---

## 🚀 Key Features

*   **AST-Based Accuracy**: True code analysis for Go, JavaScript, and TypeScript (no more regex false positives).
*   **High-Performance Caching**: Uses SHA-256 file hashing to skip unchanged files. 10,000 files scanned in milliseconds.
*   **Language Polyglot**: First-class support for **9 languages** (Go, JS, TS, Python, Rust, Ruby, Java, C#, PHP, Kotlin).
*   **Zero Config**: Works out of the box with standard `.env` patterns.
*   **Safety First**: Built-in protection against path traversal and resource exhaustion.

---

## 📦 Installation

```bash
# Go install
go install github.com/Hossiy21/taso@latest

# Verify
taso version
```

---

## � Security

Taso is built with security in mind:

- ✅ **No sensitive data stored** — Only analyzes source code, never captures environment variable values
- ✅ **Protected against attacks** — Path traversal, resource exhaustion, symlink attacks blocked
- ✅ **Safe caching** — SHA-256 hashing, no secrets in cache
- ✅ **Audit logging** — Track all scans and issues

For detailed security information, see our [Security Policy](SECURITY.md).

---

## �🛠 Commands

### `taso ghost` — Find "Ghost" Variables
Scans source code to find variables accessed in code but missing from `.env`.

```bash
$ taso ghost

👻  2 ghost variable(s) found

  STRIPE_WEBHOOK_SECRET
    used in:     api/webhooks.go:47
    not in:      .env, .env.local

  DATABASE_REPLICA_URL
    used in:     db/pool.go:12
    not in:      .env
```

| Flag | Description |
|---|---|
| `--fix` | Interactively add missing variables to your `.env` |
| `--json` | Export findings for CI/CD pipelines |
| `--dir <path>` | Specify a custom directory to scan |

---

### `taso score` — Health Audit
Gives your project an environment health score (0–100) based on ghost variables, placeholder values, and security posture.

```bash
$ taso score

  Env Health Score
  [====================----------]  B  72/100

  ⚠  1 ghost variable(s) found — run 'taso ghost' to see them
  ⚠  3 empty or placeholder value(s) in your env files
```

---

### `taso snap` + `taso drift` — Drift Tracking
Track how your environment evolves over time.

```bash
taso snap     # Save a baseline of your current keys
taso drift    # See what keys were added, removed, or changed since the snapshot
```

---

## 🌍 Supported Languages

| Language | Analysis Method | Patterns Detected |
|---|---|---|
| **Go** | AST | `os.Getenv`, `os.LookupEnv` |
| **JS / TS** | AST | `process.env`, `import.meta.env`, **Destructuring** |
| **Python** | Regex+ | `os.environ`, `os.getenv`, `environ.get` |
| **Rust** | Regex+ | `env::var`, `env!`, `option_env!` |
| **Ruby** | Regex+ | `ENV["KEY"]`, `ENV.fetch` |
| **Java** | Regex+ | `System.getenv` |
| **C#** | Regex+ | `Environment.GetEnvironmentVariable` |
| **PHP** | Regex+ | `getenv`, `$_ENV`, `$_SERVER` |
| **Kotlin** | Regex+ | `System.getenv` |

---

## ⚙️ Configuration

Taso works with no config, but you can customize it with a `.taso.yaml`:

```yaml
ignored_dirs:
  - vendor
  - node_modules
  - .git
  - dist
  - custom_build
```

---

## 🛡️ CI/CD Integration

Taso is built for automation. Use the `--json` flag in your GitHub Actions or GitLab CI to fail builds if the environment score drops too low.

```bash
# Example CI Check
taso ghost --json | jq '.ghost_count == 0'
```

### GitHub Actions Example

```yaml
name: Environment Drift Check

on: [pull_request, push]

jobs:
  taso-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - run: go install github.com/Hossiy21/taso@latest
      - run: taso score
      - run: taso ghost --json
```

---

## 💡 Usage Examples

### Real-World Scenario 1: Microservices Architecture

**Problem:** Your team has 5 microservices, each with their own environment variables. A new developer forgets to set `KAFKA_BROKER_URL` in the payment service.

```bash
$ cd payment-service
$ taso ghost

👻  1 ghost variable(s) found

  KAFKA_BROKER_URL
    used in:     events/producer.go:34
    not in:      .env, .env.production
```

**Solution:** Run before deploying to catch missing variables instantly.

```bash
$ taso ghost --fix
? Add KAFKA_BROKER_URL to .env? (Y/n) y
✓ Added KAFKA_BROKER_URL to .env
```

### Real-World Scenario 2: Onboarding New Team Members

**Problem:** A new developer clones the repo and runs the app, but gets cryptic errors because they don't know which environment variables are required.

```bash
$ taso score

  Env Health Score
  [==============----]  C  65/100

  ⚠  3 ghost variable(s) found — run 'taso ghost' to see them
  ⚠  5 empty or placeholder value(s) in your env files

$ taso ghost
# Shows exactly what's missing — no guesswork!
```

### Real-World Scenario 3: Tracking Environment Drift Over Time

**Problem:** Your production environment has been stable for months, but something changed. Did someone add a secret? Remove a config?

```bash
# When you deploy, save a snapshot
$ taso snap
✓ Snapshot saved to .taso.snap

# Later, check for drift
$ taso drift

  Environment Drift Report
  
  Added Keys:
    - NEW_FEATURE_FLAG
    - ANALYTICS_TOKEN

  Removed Keys:
    - LEGACY_SERVICE_URL
```

### Real-World Scenario 4: Pre-commit Hook

Prevent environment variable issues before they reach CI/CD:

```bash
#!/bin/bash
# .git/hooks/pre-commit
taso ghost
if [ $? -ne 0 ]; then
  echo "❌ Ghost variables detected! Fix them before committing."
  exit 1
fi
```

---

## 🤝 Contributing

We welcome contributions from the community! Whether it's bug fixes, new language support, or documentation improvements, your help makes Taso better.

### How to Contribute

1. **Fork the repository** and create a feature branch:
   ```bash
   git clone https://github.com/Hossiy21/taso.git
   cd taso
   git checkout -b feature/your-feature-name
   ```

2. **Set up your development environment:**
   ```bash
   go mod download
   go test ./...
   ```

3. **Make your changes** and write tests:
   ```bash
   go test ./...  # Ensure all tests pass
   go fmt ./...   # Format your code
   ```

4. **Commit and push your changes:**
   ```bash
   git commit -m "feat: add support for new language"
   git push origin feature/your-feature-name
   ```

5. **Open a Pull Request** with a clear description of your changes.

### Development Guidelines

- **Code Style:** Follow Go's standard conventions. Use `gofmt` and `golint`.
- **Testing:** All new features must include unit and integration tests.
- **Documentation:** Update the README and code comments as needed.
- **Performance:** Avoid changes that negatively impact scanning speed.

### Areas We're Looking For Help

- **New Language Support:** Add AST-based or regex patterns for languages not yet supported.
- **Performance Optimizations:** Help us scan even faster.
- **Documentation:** Improve guides, examples, and tutorials.
- **Bug Fixes:** Report issues and submit fixes.
- **Integrations:** Add plugins for popular tools and frameworks.

### Getting Help

- **Questions?** Open a [GitHub Discussion](https://github.com/Hossiy21/taso/discussions)
- **Found a bug?** [Report it here](https://github.com/Hossiy21/taso/issues)

---

## 🗺️ Roadmap

### Phase 1: Foundation (Current)
- ✅ AST-based scanning for Go, JS/TS
- ✅ Environment health scoring
- ✅ Drift tracking
- ✅ High-performance caching
- ✅ Security hardening

### Phase 2: Expansion (Q3 2026)
- 🔄 Enhanced language support (prioritize Python, Rust, Java)
- 🔄 Web dashboard for visualization
- 🔄 Team collaboration features (shared snapshots, audit logs)
- 🔄 VS Code extension for real-time linting

### Phase 3: Intelligence (Q4 2026)
- 📋 ML-powered anomaly detection for environment variables
- 📋 Automated secret rotation recommendations
- 📋 Integration with HashiCorp Vault and AWS Secrets Manager
- 📋 Advanced CI/CD templates for all major platforms

### Phase 4: Enterprise (Q1 2027)
- 📋 Multi-repository scanning
- 📋 Advanced RBAC and audit logging
- 📋 On-premises deployment options
- 📋 SLA and premium support

### Community-Driven Features

Have a feature request? Vote on and suggest ideas in [GitHub Discussions](https://github.com/Hossiy21/taso/discussions/categories/feature-requests).

---

## 📄 License

MIT — Created by [Hossiy21](https://github.com/Hossiy21)
