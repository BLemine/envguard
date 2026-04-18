# envguard

**Keep your `.env` files honest.**

A lightweight CLI tool that checks missing, undocumented, and empty environment variables, syncs local files from examples, audits git history for leaked secrets, and encrypts/decrypts env files for safer team sharing.

---

## The problem

Every project has a `.env.example` committed to git and a real `.env` that's gitignored. This breaks in three ways:

- A new dev clones the repo and has no idea which keys are missing
- Someone adds a new key to `.env` but forgets to update `.env.example`
- A key exists but is empty — and nobody notices until production

`envguard` catches all of this, and helps you recover when it doesn't.

---

## Install

With Homebrew:

```bash
brew tap BLemine/tap
brew install BLemine/tap/envguard
envguard --help
```

With Go:

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
  ✗ STRIPE_SECRET_KEY                       missing from .env
  ✗ REDIS_URL                               missing from .env
  ⚠ DEBUG_MODE                              in .env but not in .env.example

Summary
  ✓ 2 ok
  ✗ 2 missing
  ⚠ 1 undocumented

✗ Check failed — fix missing or empty keys before continuing
```

Custom paths:
```bash
envguard check --example=.env.staging.example --local=.env.staging
```

---

### `sync` — add missing keys to your `.env`

Adds missing keys from `.env.example` into your `.env` with empty values. **Existing keys are never overwritten.**
If the target file does not exist yet, `envguard` creates it for you.

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

### `encrypt` — encrypt a `.env` file for safe sharing

Encrypts your `.env` into a `.env.enc` file using **AES-256-GCM** with a key derived from your passphrase via **Argon2id**. The passphrase is never stored anywhere.

```bash
envguard encrypt --passphrase=mysecret
```

Avoid passing secrets directly on the command line if you can, since shell history and process inspection may expose them. Prefer CI secret injection or a shell prompt that exports `ENVGUARD_PASSPHRASE` only for the current session.

```
✓ .env encrypted → .env.enc (share this with your team)
```

Custom paths:
```bash
envguard encrypt --local=.env.staging --out=.env.staging.enc --passphrase=mysecret
```

Via environment variable (useful in CI):
```bash
ENVGUARD_PASSPHRASE=mysecret envguard encrypt
```

---

### `decrypt` — decrypt a `.env.enc` file

Decrypts a `.env.enc` back into `.env`. Fails loudly if the passphrase is wrong. Will **not** overwrite an existing `.env` unless `--force` is passed.

```bash
envguard decrypt --passphrase=mysecret
```

```
✓ .env.enc decrypted → .env
```

Custom paths or force overwrite:
```bash
envguard decrypt --local=.env.staging.enc --out=.env.staging --passphrase=mysecret
envguard decrypt --passphrase=mysecret --force
```

> **Note:** `.env.enc` is gitignored by default. Commit it intentionally only after confirming your team knows the passphrase out-of-band.

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

### `audit` — scan git history for leaked env files and secrets

Walks the full git history and flags committed `.env` files plus lines matching common secret patterns such as API keys, tokens, passwords, private keys, Slack tokens, GitHub tokens, and Stripe keys.

```bash
envguard audit
```

Custom repository path:
```bash
envguard audit --repo /path/to/other/repo
```

Example output:
```text
Audit: scanning git history for secrets and .env files

.env Files Committed to History
  ⚠ e92d843  .env

Secret Patterns Detected
  ✗ e92d843  .env                                Generic API Key
         API_KEY=[REDACTED]

Summary
  ⚠ 1 .env file(s) found in history
  ✗ 1 secret pattern(s) detected

✗ Audit failed — sensitive data may be exposed in git history
  Tip: use `git filter-repo` or BFG Repo Cleaner to scrub history
```

`audit` exits with code `1` when it finds leaked env files or matching secret patterns, which makes it suitable for CI or pre-release checks without dumping the secret value into logs.

---

## Project structure

```
envguard/
├── cmd/
│   ├── root.go       # Cobra root command
│   ├── check.go      # envguard check
│   ├── sync.go       # envguard sync
│   ├── validate.go   # envguard validate
│   ├── audit.go      # envguard audit
│   ├── encrypt.go    # envguard encrypt
│   ├── decrypt.go    # envguard decrypt
│   └── passphrase.go # shared passphrase resolution
├── internal/
│   ├── parser/       # .env file parsing
│   ├── differ/       # diff logic
│   ├── reporter/     # colored terminal output
│   ├── auditor/      # git history scanning
│   └── crypto/       # AES-256-GCM + Argon2id
└── main.go
```

---

## Roadmap

- [x] `audit` — scan git history for accidentally committed secrets
- [x] `encrypt` / `decrypt` — encrypted `.env` for safe team sharing
- [ ] Multi-file support (`.env.staging`, `.env.test`, `.env.production`)
- [ ] GitHub Actions marketplace action

---

## Contributing

PRs welcome. Open an issue first for anything beyond small fixes.

## License

MIT
