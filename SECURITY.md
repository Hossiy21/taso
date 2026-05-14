# Security Policy

## Reporting Security Vulnerabilities

We take security seriously. If you discover a security vulnerability in Taso, please **do not** open a public GitHub issue. Instead, please follow these steps:

1. **Report privately**: Email security concerns to [security@taso-project.dev](mailto:security@taso-project.dev) or open a [GitHub Security Advisory](https://github.com/Hossiy21/taso/security/advisories).
2. **Include details**:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

3. **Timeline**: We will:
   - Acknowledge receipt within 48 hours
   - Provide an initial assessment within 5 business days
   - Release a fix in a timely manner
   - Credit the reporter (unless you prefer anonymity)

---

## Security Considerations

### What Taso Does (Safe)

Taso **only analyzes source code** to find environment variable references. It:

- ✅ Scans source files for environment variable access patterns
- ✅ Reads `.env` files to compare against actual code usage
- ✅ Does **NOT** capture, store, or transmit the actual values of environment variables
- ✅ Uses SHA-256 hashing for cache validation (no sensitive data in cache)
- ✅ Logs only metadata about scans (file counts, durations, errors)

### What Taso Protects Against

Taso is hardened against common attack vectors:

#### 1. **Path Traversal Attacks** ✅
- Validates all file paths before access
- Rejects paths attempting to escape the scan directory
- Blocks symbolic links to prevent circumvention
- Rejects absolute paths to system directories

#### 2. **Resource Exhaustion / Denial of Service** ✅
- Enforces file size limits (50MB per file maximum)
- Enforces total scan timeout (5 minutes)
- Monitors cumulative memory and bytes scanned
- Gracefully skips files that exceed limits with logging

#### 3. **Code Injection** ✅
- Uses proper AST parsing for Go, JavaScript, TypeScript (no regex injection risks)
- For other languages, uses anchored regex patterns to prevent injection
- Never evaluates or executes code

#### 4. **Symlink Following** ✅
- Explicitly rejects symlinks in scan directories and `.env` files
- Prevents attackers from redirecting scans to sensitive system files

#### 5. **Null Byte Injection** ✅
- Validates all paths for null bytes before processing
- Prevents bypassing file access controls

---

## Deployment Security Best Practices

### For CI/CD Pipelines

```yaml
# ✅ RECOMMENDED: Run on ephemeral, isolated containers
- name: Security Check
  runs-on: ubuntu-latest
  container:
    image: golang:1.23-alpine
  steps:
    - uses: actions/checkout@v3
    - run: go install github.com/Hossiy21/taso@latest
    - run: taso ghost
```

### File Permissions

Ensure proper file permissions when running Taso:

```bash
# ✅ Recommended: Restrict .env files to owner only
chmod 600 .env .env.local .env.production

# ✅ Recommended: Make .taso cache directory restricted
chmod 700 .taso/
```

### Secrets Management

**Never commit `.env` files to version control.** Instead:

1. **Use `.env.example`** (without values):
   ```bash
   # .env.example
   DATABASE_URL=
   API_KEY=
   STRIPE_SECRET_KEY=
   ```

2. **Use secret management tools**:
   - GitHub Secrets for CI/CD
   - HashiCorp Vault for production
   - AWS Secrets Manager / Azure Key Vault for cloud deployments

3. **Use Taso to validate secrets are configured**:
   ```bash
   taso ghost  # Ensure no secrets are missing
   taso score  # Monitor overall env health
   ```

### Production Deployment

When running Taso in production or on sensitive environments:

1. **Use restricted service accounts** with minimal file system access
2. **Run in sandboxed containers** (e.g., Docker with `--read-only`)
3. **Audit logs** for compliance using `--json` output
4. **Monitor resource usage** to detect DoS attempts:
   ```bash
   # Example monitoring
   timeout 300 taso ghost --json > results.json
   ```

---

## Known Limitations

### What Taso Cannot Detect

- **Dynamic environment variable access** (e.g., `os.Getenv(myVar)` where `myVar` is not a literal string)
- **Variables passed via command-line arguments** (only source code scanning)
- **Runtime-injected variables** (e.g., Docker secrets, Kubernetes ConfigMaps)

For these cases, consider complementary tools like:
- Static analysis frameworks (SonarQube, Trivy)
- Secrets scanning (TruffleHog, GitGuardian)
- Supply chain security (Snyk, Dependabot)

### Cache Security

The `.taso` directory contains:
- **Hashes** (SHA-256) of scanned files
- **Cached results** (environment variable names, not values)
- **No sensitive values** are ever cached

However, you should:
- Add `.taso/` to your `.gitignore`
- Restrict permissions: `chmod 700 .taso/`

---

## Secure Development Practices

### For Taso Contributors

- **Code Review**: All changes require peer review before merging
- **Dependency Management**: 
  - Use `go mod verify` to validate dependencies
  - Keep dependencies up-to-date with `go get -u`
  - Check for vulnerabilities with `go list -json ./... | nancy sleuth`
- **Testing**: All new features must include security-focused tests
- **Fuzzing**: Critical parsing functions should have fuzzing tests

### Dependency Safety

Taso dependencies include:
- **Cobra** (CLI framework) — actively maintained
- **Viper** (configuration) — actively maintained
- **Godotenv** (.env parsing) — stable, minimal scope
- **Lipgloss** (terminal UI) — actively maintained
- **Goja** (JS/TS AST parsing) — actively maintained

All dependencies are vetted for:
- Active maintenance
- Security issue responsiveness
- Minimal external dependencies
- License compatibility (permissive licenses preferred)

---

## Security Roadmap

### Current (Phase 1)
- ✅ Path traversal prevention
- ✅ Resource exhaustion protection
- ✅ Symlink rejection
- ✅ Null byte filtering
- ✅ Security testing

### Q3 2026 (Phase 2)
- 🔄 SBOM (Software Bill of Materials) generation
- 🔄 Secrets detection in `.env` files (warn about hardcoded secrets)
- 🔄 GPG signing of releases
- 🔄 Enhanced audit logging for compliance

### Q4 2026 (Phase 3)
- 📋 Integration with SIEM systems
- 📋 Automated vulnerability scanning in CI/CD
- 📋 Security policy templates for enterprises

### Q1 2027 (Phase 4)
- 📋 Enterprise deployment options
- 📋 Advanced RBAC and audit logging
- 📋 Multi-repository scanning

---

## Versioning and Releases

### Semantic Versioning

Taso follows **Semantic Versioning** (SemVer) for all releases:

- **Major** (X.0.0): Breaking changes or major features
- **Minor** (0.X.0): New features, backward compatible
- **Patch** (0.0.X): Bug fixes, security patches

Example: `v1.2.3` = Major 1, Minor 2, Patch 3

### Release Process

1. **Update Version**:
   ```bash
   # Update version in code/docs
   # Tag the release in Git
   git tag -a v1.2.3 -m "Release version 1.2.3"
   git push origin v1.2.3
   ```

2. **GitHub Releases**:
   - Pre-built binaries for all platforms (Linux, macOS, Windows)
   - Detailed release notes with:
     - New features
     - Bug fixes
     - Security patches
     - Migration guide (if breaking changes)
   - Checksums (SHA256) for binary verification

3. **Supported Platforms**:
   - Linux: amd64, arm64
   - macOS: amd64 (Intel), arm64 (Apple Silicon)
   - Windows: amd64, arm64

### Using GoReleaser

Configure `.goreleaser.yaml` to automate releases:

```yaml
version: 2

builds:
  - id: taso
    main: ./main.go
    binary: taso
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

archives:
  - id: default
    name_template: 'taso_{{ .Version }}_{{ .Os }}_{{ .Arch }}'
    files:
      - README.md
      - SECURITY.md
      - LICENSE

checksum:
  name_template: 'checksums.txt'

release:
  github:
    owner: Hossiy21
    name: taso
  prerelease: auto
```

### Installation from Releases

Users can install pre-built binaries:

```bash
# Using go install (latest)
go install github.com/Hossiy21/taso@latest

# Using go install (specific version)
go install github.com/Hossiy21/taso@v1.2.3

# Download binary directly
wget https://github.com/Hossiy21/taso/releases/download/v1.2.3/taso_1.2.3_linux_amd64
chmod +x taso_1.2.3_linux_amd64
sudo mv taso_1.2.3_linux_amd64 /usr/local/bin/taso
```

---

## Dependency Auditing

### Regular Audits

Perform security audits regularly:

```bash
# Check for vulnerabilities
go list -json ./... | nancy sleuth

# Update dependencies
go get -u ./...

# Verify dependencies
go mod verify

# Check for outdated packages
go list -u -m all
```

### Dependency Management Best Practices

1. **Minimize Direct Dependencies**:
   - Only add dependencies that are essential
   - Prefer native Go implementations when available
   - Replace indirect dependencies with direct ones if critical

2. **Vet All Dependencies**:
   - Check GitHub repository activity
   - Review security advisories
   - Verify license compatibility
   - Check dependency tree for conflicts

3. **Update Strategy**:
   - Monthly: Check for security patches (urgent)
   - Quarterly: Update to latest minor versions
   - Annually: Evaluate major version upgrades

4. **Current Dependencies**:
   - **Cobra** (CLI): Core dependency, well-maintained
   - **Viper** (Config): Configuration management
   - **Godotenv** (.env): Minimal, proven .env parser
   - **Lipgloss** (UI): Terminal styling
   - **Goja** (JS/TS AST): JavaScript parser

### Automated Dependency Scanning

Enable in GitHub:

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
    security-updates-only: false
    open-pull-requests-limit: 5
    reviewers:
      - "Hossiy21"
```

### Vulnerability Disclosure

If a dependency has a security issue:

1. Assess impact on Taso
2. Update dependency immediately
3. Release patch version
4. Update security advisory
5. Credit the researcher

---

## Compliance

Taso is designed to support compliance with:
- **SOC 2** audit requirements (configuration management)
- **ISO 27001** (information security controls)
- **GDPR** (data minimization — no personal data collected)
- **HIPAA** (audit trails available via `--json`)
- **PCI-DSS** (environment variable tracking)

---

## Questions?

For security questions or concerns:
- 📧 Email: yhosaina@outlook.com
- 🐛 Report a bug: [GitHub Issues](https://github.com/Hossiy21/taso/issues)
- 💬 Ask a question: [GitHub Discussions](https://github.com/Hossiy21/taso/discussions)

---

**Last Updated**: May 14, 2026
**Version**: 1.0.0
