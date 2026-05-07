# taso

**Find what's silently wrong with your environment — before production does.**

Offline. No cloud account. One binary. Every language.

```
taso ghost      → finds env vars your code calls but don't exist in .env
taso score      → gives your project an env health score 0–100
taso snap       → saves a snapshot of your current env state
taso drift      → shows what changed since last snapshot
taso share      → generates a shareable team fingerprint (keys only, no values)
```

---

## Why taso?

Every other env tool compares `.env` files to `.env` files.

**Taso compares your source code to your `.env` files.**

It parses your actual Go, JS, Python, Rust, or Ruby code and finds every env var your code calls — then cross-checks against your env files. If you're calling `STRIPE_WEBHOOK_SECRET` in your code but it doesn't exist anywhere in your env files, your app will crash in production. Taso tells you *before* that happens.

---

## Installation

```bash
# Homebrew (macOS/Linux)
brew tap Hossiy21/tap
brew install taso

# Go install
go install github.com/Hossiy21/taso@latest

# Verify
taso version
```

---

## Commands

### `taso ghost` — the killer feature

Scans your source code and finds every env var your code calls that doesn't exist in your `.env` files.

```
$ taso ghost

Scanning source in .
Checking against: .env, .env.local
Found 12 env vars in source code

👻  3 ghost variable(s) found

  STRIPE_WEBHOOK_SECRET
    used in:     api/webhooks.go:47
    used in:     api/webhooks.go:89
    not in:      .env, .env.local

  DATABASE_REPLICA_URL
    used in:     db/pool.go:12
    not in:      .env, .env.local

  REDIS_TLS_CERT
    used in:     cache/client.go:33
    not in:      .env
```

Supports: Go (AST-based), JavaScript/TypeScript, Python, Rust, Ruby, Java.

```bash
taso ghost
taso ghost --dir ./src --env .env.production
taso ghost --json
```

---

### `taso score` — env health score

Scores your project's env health 0–100. Factors in ghost vars, empty/placeholder values, missing `.env.example`, and whether you've taken a snapshot.

```
$ taso score

  Env Health Score

  ████████████████░░░░░░░░░░░░░░  B  72/100

  2 issue(s) found:

  ⚠  1 ghost variable(s) found — run 'taso ghost' to see them
  ⚠  3 empty or placeholder value(s) in your env files
```

---

### `taso snap` + `taso drift` — change tracking

```bash
# Save baseline
taso snap

# One week later, see what changed
taso drift
```

```
$ taso drift

Snapshot taken: 6 days ago (2025-01-10 09:00)

⚡  2 change(s) detected

  + NEW_PAYMENT_PROVIDER          added
  ~ REDIS_URL                     value changed
```

---

### `taso share` — team fingerprint

Generates a shareable digest of your env structure — key names and types, no values. Share it with teammates so they can verify their setup matches yours.

```
$ taso share

  Team Fingerprint

  24 variables across 2 file(s)

  DATABASE_URL                             url
  NEXT_PUBLIC_API_URL                      url
  REDIS_URL                                url
  STRIPE_SECRET_KEY                        secret
  ...

  Fingerprint: 3f8a1c902d44
```

---

## Supported Languages

| Language | Patterns detected |
|---|---|
| Go | `os.Getenv("KEY")`, `os.LookupEnv("KEY")` |
| JavaScript/TypeScript | `process.env.KEY`, `process.env["KEY"]`, `import.meta.env.KEY` |
| Python | `os.environ["KEY"]`, `os.getenv("KEY")`, `os.environ.get("KEY")` |
| Rust | `env::var("KEY")`, `env!("KEY")`, `option_env!("KEY")` |
| Ruby | `ENV["KEY"]`, `ENV.fetch("KEY")` |
| Java | `System.getenv("KEY")` |

---

## JSON output

Every command supports `--json` for CI/CD and scripting:

```bash
taso ghost --json
taso score --json
taso drift --json
```

---

## CI/CD Integration

```yaml
- name: Check for ghost env vars
  run: taso ghost --json

- name: Env health score
  run: taso score --json
```

---

## How it differs from Razify

| | razify | taso |
|---|---|---|
| Compares | `.env` ↔ `.env.example` | **source code ↔ `.env`** |
| Finds | missing/leaked keys in env files | vars used in code but missing from env |
| Scans | env file values | actual source code (AST + regex) |
| Use case | env file hygiene | code-env gap detection |

They're complementary. Use both.

---

## License

MIT — [Hossiy21](https://github.com/Hossiy21)
