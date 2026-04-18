# envguard

**Keep your `.env` files honest.**

A lightweight CLI tool that catches missing, undocumented, and empty environment variables before they cause problems in production.

---

## The problem

Every project has a `.env.example` committed to git and a real `.env` that's gitignored. This breaks in three ways:

- A new dev clones the repo and has no idea which keys are missing
- Someone adds a new key to `.env` but forgets to update `.env.example`
- A key exists but is empty — and nobody notices until production

`envguard` catches all of this.

---

## Install

```bash
go install github.com/BLemine/envguard@latest
```

Or download a pre-built binary from [Releases](https://github.com/BLemine/envguard/releases).

---

## Usage

### `check` — diff your `.env` against `.env.example`

```bash
envguard check
```

```
Comparing .env.example → .env

  ✓ DATABASE_URL                            ok
  ✓ JWT_SECRET                              ok
  ✗ STRIPE_SECRET_KEY                       missing from your .env
  ✗ REDIS_URL                               missing from your .env
  ⚠ DEBUG_MODE                              in .env but not in .env.example

Summary
  ✓ 2 ok
  ✗ 2 missing
  ⚠ 1 undocumented

✗ Check failed — run `envguard sync` to fill missing keys
```

Custom paths:
```bash
envguard check --example=.env.staging.example --local=.env.staging
```

---

### `sync` — add missing keys to your `.env`

Adds missing keys from `.env.example` into your `.env` with empty values. **Existing keys are never overwritten.**

```bash
envguard sync
```

```
Sync result
  + STRIPE_SECRET_KEY                       added (empty value)
  + REDIS_URL                               added (empty value)
  ~ DATABASE_URL                            already exists, skipped
```

---

### `validate` — assert required keys are non-empty

Perfect for CI pipelines. Exits with code 1 if any required key is missing or empty.

```bash
envguard validate --required=DATABASE_URL,API_KEY,JWT_SECRET
```

```
  ✓ DATABASE_URL                            set
  ✓ JWT_SECRET                              set
  ✗ API_KEY                                 missing or empty

✗ Validation failed — 1 required key(s) missing or empty
```

Use in GitHub Actions:
```yaml
- name: Validate env
  run: envguard validate --required=DATABASE_URL,API_KEY,JWT_SECRET
```

---

## Project structure

```
envguard/
├── cmd/
│   ├── root.go       # Cobra root command
│   ├── check.go      # envguard check
│   ├── sync.go       # envguard sync
│   └── validate.go   # envguard validate
├── internal/
│   ├── parser/       # .env file parsing
│   ├── differ/       # diff logic
│   └── reporter/     # colored terminal output
└── main.go
```

---

## Roadmap

- [ ] `audit` — scan git history for accidentally committed secrets
- [ ] `encrypt` / `decrypt` — encrypted `.env` for safe team sharing
- [ ] Multi-file support (`.env.staging`, `.env.test`, `.env.production`)
- [ ] GitHub Actions marketplace action

---

## Contributing

PRs welcome. Open an issue first for anything beyond small fixes.

## License

MIT
