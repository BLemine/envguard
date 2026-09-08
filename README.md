# envguard

**Keep your `.env` files honest.**

A lightweight CLI tool that checks missing, undocumented, and empty environment variables, syncs local files from examples, scans YAML and properties config for env placeholders, audits git history for leaked secrets, and encrypts/decrypts env files for safer team sharing.

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

With Go (Go 1.25 or newer):

```bash
go install github.com/BLemine/envguard@v1.0.0
```

Use `@latest` instead to install the latest release. The executable is installed
in `GOBIN`, or `$(go env GOPATH)/bin` when `GOBIN` is unset; add that directory
to your `PATH` if `envguard` is not found.

Pre-built binaries require no Go installation. Download the matching asset from
[the v1.0.0 release](https://github.com/BLemine/envguard/releases/tag/v1.0.0):

| Platform | Asset |
|----------|-------|
| macOS Apple Silicon | `envguard-darwin-arm64` |
| macOS Intel | `envguard-darwin-amd64` |
| Linux x86-64 | `envguard-linux-amd64` |
| Windows x86-64 | `envguard-windows-amd64.exe` |

For example, install the Linux x86-64 binary into your user account:

```bash
curl -fL https://github.com/BLemine/envguard/releases/download/v1.0.0/envguard-linux-amd64 -o envguard
chmod +x envguard
mkdir -p "$HOME/.local/bin"
mv envguard "$HOME/.local/bin/envguard"
export PATH="$HOME/.local/bin:$PATH"
envguard --help
```

On macOS, substitute the matching asset name in the download URL. On Windows,
rename the downloaded file to `envguard.exe` and put it in a directory on `PATH`,
or run it directly as `.\envguard.exe --help` in PowerShell.
Git must be installed and available on `PATH` to use `audit`.

---

## Quick start

From a project directory containing `.env.example`:

```bash
envguard sync
# Fill in the empty values added to .env, then:
envguard check
```

Paths are relative to the current working directory. `check` and `validate`
read `.env` by default; they do not validate the current process environment.
Use `envguard <command> --help` to inspect that command's flags.

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

`check` exits with code `1` for missing or empty keys. Undocumented keys produce
warnings only and do not fail the check. Both input files must exist.

Custom paths:
```bash
envguard check --example=.env.staging.example --local=.env.staging
```

---

### `sync` — add missing keys to your `.env`

Adds missing keys from `.env.example` into your `.env` with empty values. **Existing keys are never overwritten.**
If the target file does not exist yet and keys need adding, `envguard` creates it for you.
It appends empty assignments, preserving existing values and comments; it does
not copy example values or fill keys that already exist but are empty.
Use `--example=.env.staging.example --local=.env.staging` for custom paths.

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

Encrypts your `.env` into a `.env.enc` file using **AES-256-GCM** with a key derived from your passphrase via **Argon2id**. envguard does not save the passphrase. Existing output files are overwritten without a `--force` flag; choose a different `--out` path to keep an earlier encrypted copy.

```bash
# Bash: read without echoing or placing the passphrase in shell history.
read -r -s -p "Passphrase: " ENVGUARD_PASSPHRASE
printf '\n'
export ENVGUARD_PASSPHRASE
envguard encrypt
unset ENVGUARD_PASSPHRASE
```

Avoid passing secrets directly on the command line if you can, since shell history and process inspection may expose them. Prefer CI secret injection or a shell prompt that exports `ENVGUARD_PASSPHRASE` only for the current session.

```
✓ .env encrypted → .env.enc (share this with your team)
```

Custom paths:
```bash
envguard encrypt --local=.env.staging --out=.env.staging.enc --passphrase=mysecret
```

In CI, inject a stored secret as `ENVGUARD_PASSPHRASE` rather than writing its
value into the workflow. A non-empty `--passphrase` takes precedence over the
environment variable. Encryption and decryption fail if neither supplies a passphrase.

---

### `decrypt` — decrypt a `.env.enc` file

Decrypts a `.env.enc` back into `.env`. Fails loudly if the passphrase is wrong. Will **not** overwrite an existing `.env` unless `--force` is passed. On Unix, decrypted output is restricted to owner read/write (`0600`), including when overwriting an existing file.

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

> **Note:** This repository ignores `*.enc`. Installing envguard does not change
> another project's `.gitignore`; configure that project explicitly. Share the
> passphrase separately from the encrypted file. Decryption also rejects
> corrupted ciphertext before writing output.

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

Use `--local=.env.staging` to validate another file. Validation checks for
missing keys or the empty string; it does not validate formats, credentials,
or whitespace-only values. Setting a key through GitHub Actions `env:` alone
does not put it into the file. See [GitHub Actions](#github-actions) for CI setup.

---

### `audit` — scan git history for leaked env files and secrets

Walks history reachable from all local refs, including merge commits, and flags committed `.env` files plus lines matching common secret patterns such as API keys, tokens, passwords, private keys, Slack tokens, GitHub tokens, and Stripe keys.

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

The audit uses a finite set of patterns and can produce false positives or miss
secrets. It scans committed additions, including merge diffs, in history
reachable from local refs; it does not fetch remote refs or inspect uncommitted
files. Shallow clones limit the history available to it. Template filenames
ending in `.example`, `.sample`, or `.template` are excluded from the env-file
check, but their contents still undergo pattern matching. A clean audit is not
a guarantee that the repository contains no secrets. Rotate exposed credentials;
removing them from history does not revoke them.

---

### `scan-config` — scan Spring Boot / Quarkus config files for env variable placeholders

Extracts `${VAR}` and `${VAR:default}` placeholders from YAML and `.properties` config files.
Required variables (no default) and optional variables (with default) are reported separately.
All YAML documents (including sections separated by `---`) are scanned. Properties
accept `=`, `:`, or whitespace separators. Nested fallback expressions are supported, including deeper chains:

```properties
database.url=${DATABASE_URL:${FALLBACK_DATABASE_URL}}
server.port=${SERVER_PORT:${DEFAULT_PORT:8080}}
```

For the first expression, `--strict` accepts either key; it fails if both are
absent. The second expression always has a literal fallback. When a fallback
contains several references, all must resolve unless the primary key is present.
Each supplied env file is checked independently. As with simple placeholders,
empty values count as defined; use `check` or `validate` to reject empty values.

Human output preserves the full default expression. JSON adds a recursive
`fallback` array to placeholders with nested references, and `--quiet` includes
nested variable names. Missing conditional requirements are reported as their
full `${...}` expression in `missing_local` / `missing_example`.

```bash
envguard scan-config --files=application.yml
```

```
Scanned: application.yml

Required variables:
  - DATABASE_URL                (spring.datasource.url)
  - DATABASE_PASSWORD           (spring.datasource.password)

Variables with defaults (nested fallbacks are conditional):
  - SERVER_PORT=8080    (server.port)
```

Use `--strict` with `--local` / `--example` to verify all required variables are defined:

```bash
envguard scan-config --files=application.yml --local=.env --example=.env.example --strict
```

```
Required variables found in config: 2
Optional variables found in config: 1

Required variables:
  - DATABASE_URL                (spring.datasource.url)
  - DATABASE_PASSWORD           (spring.datasource.password)

Variables with defaults (nested fallbacks are conditional):
  - SERVER_PORT=8080    (server.port)

Missing in .env:
  - DATABASE_PASSWORD

Missing in .env.example:
  - DATABASE_URL
```

Exits with code `1` when `--strict` is set and a required variable or fallback
expression cannot be satisfied by a provided env file. Without `--strict`,
missing requirements are reported but do not cause failure. `--strict` alone
performs no comparison: supply `--local`, `--example`, or both. Read and parse
errors fail regardless of strict mode.

The scanner recognizes uppercase env names matching `[A-Z_][A-Z0-9_]*`.
It is a placeholder scanner, not a complete Spring Boot or Quarkus configuration
resolver: it does not select active profiles, resolve property-name references,
expand YAML aliases, or join properties continuation lines. All YAML documents
are scanned regardless of profile. Malformed or unsupported placeholder syntax
may be skipped, so a successful scan does not validate the configuration syntax
of the target framework.

**Spring Boot `application.yml`**

```yaml
spring:
  datasource:
    url: ${DATABASE_URL}
    password: ${DATABASE_PASSWORD}
server:
  port: ${SERVER_PORT:8080}
```

**Quarkus `application.properties`**

```properties
quarkus.datasource.jdbc.url=${DATABASE_URL}
quarkus.datasource.password=${DATABASE_PASSWORD}
quarkus.http.port=${SERVER_PORT:8080}
```

Flags:

| Flag | Description |
|------|-------------|
| `--files` | Comma-separated list of config files to scan (required) |
| `--format` | Force format: `auto` \| `yaml` \| `props` (default: `auto`, detected from extension) |
| `--local` | Path to local `.env` file for comparison |
| `--example` | Path to `.env.example` file for comparison |
| `--strict` | Fail for unresolved requirements in supplied `--local` / `--example` files |
| `--json` | Output results as JSON (for CI integrations) |
| `--quiet` | Output only variable names, one per line |

---

## GitHub Actions

The CLI can run in GitHub Actions. v1.0.0 does not provide an `action.yml`, so
`uses: BLemine/envguard@v1.0.0` is not supported yet.

Save the following as `.github/workflows/envguard.yml` in the repository you
want to audit. It installs the pinned CLI release and fails when the audit finds
committed env files or matching secret patterns:

```yaml
name: Envguard
on: [push, pull_request]

permissions:
  contents: read

jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v6
        with:
          go-version: '1.25.x'
          cache: false

      - name: Install envguard
        run: go install github.com/BLemine/envguard@v1.0.0

      - name: Audit Git history
        run: envguard audit
```

[`fetch-depth: 0`](https://github.com/actions/checkout) fetches history for all
branches and tags. [`setup-go`](https://github.com/actions/setup-go) provides the
Go toolchain. This example does not require a local `.env` or production secrets.

For repositories with application config and a committed `.env.example`, add
this step after installation, adjusting the filenames to your project:

```yaml
- name: Check documented config requirements
  run: |
    envguard scan-config \
      --files=application.yml \
      --example=.env.example \
      --strict
```

To run `check` or `validate` instead, first provision a `.env` file in the job,
or point `--local` at an existing test env file. These commands do not read
GitHub Actions `env:` entries as the input being validated. Do not commit
production secrets just to make a CI check pass.

## Exit codes

| Command | Success (`0`) | Failure (`1`) |
|---------|---------------|---------------|
| `check` | Every example key has a non-empty local value; undocumented keys may remain | An example key is missing or empty in the local file |
| `validate` | Every requested key exists and is non-empty | Missing/empty requested keys |
| `scan-config` | Scan completes; comparisons pass, or strict mode is off | Unresolved requirements with `--strict` and a comparison file |
| `audit` | No matching findings in scanned history | Committed env files or matching patterns found |
| `sync`, `encrypt`, `decrypt` | Operation completes | Operation fails |

Invalid flags, missing required arguments, and input/output errors also exit
with code `1`. `scan-config --json` and `--quiet` affect output only; `--json`
takes precedence when both are supplied.

## Project structure

```
envguard/
├── cmd/
│   ├── root.go          # Cobra root command
│   ├── check.go         # envguard check
│   ├── sync.go          # envguard sync
│   ├── validate.go      # envguard validate
│   ├── audit.go         # envguard audit
│   ├── encrypt.go       # envguard encrypt
│   ├── decrypt.go       # envguard decrypt
│   ├── scan_config.go   # envguard scan-config
│   └── passphrase.go    # shared passphrase resolution
├── internal/
│   ├── parser/          # .env file parsing
│   ├── differ/          # diff logic
│   ├── reporter/        # colored terminal output
│   ├── auditor/         # git history scanning
│   ├── configscan/      # YAML / .properties placeholder extraction
│   └── crypto/          # AES-256-GCM + Argon2id
└── main.go
```

---

## Roadmap

- [x] `audit` — scan git history for accidentally committed secrets
- [x] `encrypt` / `decrypt` — encrypted `.env` for safe team sharing
- [x] `scan-config` — extract env placeholders from Spring Boot / Quarkus config files
- [ ] Batch processing of multiple `.env` pairs in one command (custom paths
  already work; `scan-config --files` already accepts multiple config files)
- [ ] GitHub Actions marketplace action

---

## Contributing

PRs welcome. Open an issue first for anything beyond small fixes.

## License

MIT
