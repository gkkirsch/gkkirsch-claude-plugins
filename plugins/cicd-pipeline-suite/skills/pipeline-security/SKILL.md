---
name: pipeline-security
description: >
  CI/CD pipeline security — supply chain protection, dependency scanning,
  SAST/DAST integration, secrets management, SBOM generation, signed commits,
  container image scanning, and pipeline hardening.
  Triggers: "pipeline security", "supply chain", "dependency scan", "sast",
  "dast", "sbom", "container scan", "signed commits", "ci security".
  NOT for: GitHub Actions workflow syntax (use github-actions-patterns).
version: 1.0.0
allowed-tools: Read, Glob, Grep, Bash, Write, Edit
---

# Pipeline Security

## Dependency Scanning

```yaml
# .github/workflows/security.yml
name: Security Scan

on:
  push:
    branches: [main]
  pull_request:
  schedule:
    - cron: '0 6 * * 1'  # Weekly Monday 6am UTC

permissions:
  contents: read
  security-events: write

jobs:
  dependency-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # npm audit
      - uses: actions/setup-node@v4
        with: { node-version: 20 }
      - run: npm ci
      - name: npm audit
        run: |
          npm audit --production --audit-level=high 2>&1 | tee audit.txt
          if grep -q "found 0 vulnerabilities" audit.txt; then
            echo "No vulnerabilities found"
          else
            echo "::warning::Vulnerabilities detected — review audit.txt"
          fi

      # Trivy vulnerability scanner
      - name: Trivy filesystem scan
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          severity: 'HIGH,CRITICAL'
          format: 'sarif'
          output: 'trivy-results.sarif'

      - name: Upload scan results
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: 'trivy-results.sarif'

  license-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20 }
      - run: npm ci

      - name: Check licenses
        run: |
          npx license-checker --production --failOn "GPL-2.0;GPL-3.0;AGPL-3.0" \
            --excludePackages "package1;package2" \
            --json > licenses.json

          COPYLEFT=$(cat licenses.json | jq '[to_entries[] | select(.value.licenses | test("GPL|AGPL"))] | length')
          if [ "$COPYLEFT" -gt 0 ]; then
            echo "::error::Copyleft licenses detected in production dependencies"
            cat licenses.json | jq '[to_entries[] | select(.value.licenses | test("GPL|AGPL"))]'
            exit 1
          fi
```

## SAST (Static Application Security Testing)

```yaml
  codeql:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        language: ['javascript-typescript']
    steps:
      - uses: actions/checkout@v4

      - name: Initialize CodeQL
        uses: github/codeql-action/init@v3
        with:
          languages: ${{ matrix.language }}
          queries: +security-extended

      - name: Autobuild
        uses: github/codeql-action/autobuild@v3

      - name: Perform analysis
        uses: github/codeql-action/analyze@v3

  semgrep:
    runs-on: ubuntu-latest
    container:
      image: semgrep/semgrep
    steps:
      - uses: actions/checkout@v4
      - name: Run Semgrep
        run: |
          semgrep scan \
            --config auto \
            --config p/owasp-top-ten \
            --config p/nodejs \
            --sarif --output semgrep.sarif \
            --error --severity ERROR
        env:
          SEMGREP_APP_TOKEN: ${{ secrets.SEMGREP_APP_TOKEN }}

      - uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: semgrep.sarif
```

## Container Image Scanning

```yaml
  image-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Build image
        run: docker build -t myapp:scan .

      # Trivy container scan
      - name: Scan image with Trivy
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: 'myapp:scan'
          severity: 'HIGH,CRITICAL'
          exit-code: 1
          ignore-unfixed: true
          format: 'table'

      # Grype (alternative scanner)
      - name: Scan image with Grype
        uses: anchore/scan-action@v4
        with:
          image: 'myapp:scan'
          fail-build: true
          severity-cutoff: high
          output-format: sarif

      # Hadolint — Dockerfile linter
      - name: Lint Dockerfile
        uses: hadolint/hadolint-action@v3.1.0
        with:
          dockerfile: Dockerfile
          failure-threshold: warning
```

## SBOM Generation

```yaml
  sbom:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Generate SBOM with Syft
        uses: anchore/sbom-action@v0
        with:
          image: myapp:${{ github.sha }}
          format: spdx-json
          output-file: sbom.spdx.json

      - name: Attest SBOM
        uses: actions/attest-sbom@v1
        with:
          subject-name: registry.example.com/myapp
          subject-digest: ${{ steps.build.outputs.digest }}
          sbom-path: sbom.spdx.json
          push-to-registry: true

      - name: Upload SBOM to release
        uses: softprops/action-gh-release@v2
        if: startsWith(github.ref, 'refs/tags/')
        with:
          files: sbom.spdx.json
```

## Secrets Management

```yaml
# Using GitHub Environments with protection rules
jobs:
  deploy:
    runs-on: ubuntu-latest
    environment:
      name: production
      # Requires: manual approval, specific reviewers, branch protection
    steps:
      - name: Deploy
        env:
          # Secrets scoped to environment
          DB_URL: ${{ secrets.PRODUCTION_DB_URL }}
          API_KEY: ${{ secrets.PRODUCTION_API_KEY }}
        run: ./deploy.sh
```

```bash
# scripts/rotate-secrets.sh — Secret rotation pattern
#!/bin/bash
set -euo pipefail

# Generate new secret
NEW_SECRET=$(openssl rand -hex 32)

# Update in secrets manager (AWS example)
aws secretsmanager update-secret \
  --secret-id "production/api/jwt-secret" \
  --secret-string "$NEW_SECRET"

# Update GitHub Actions secret
gh secret set JWT_SECRET \
  --body "$NEW_SECRET" \
  --env production \
  --repo owner/repo

# Trigger deploy to pick up new secret
gh workflow run deploy.yml \
  --field environment=production \
  --field version=current
```

```typescript
// src/config/secrets.ts — Runtime secret loading
import {
  SecretsManagerClient,
  GetSecretValueCommand,
} from "@aws-sdk/client-secrets-manager";

const client = new SecretsManagerClient({ region: "us-east-1" });
const secretCache = new Map<string, { value: string; expiresAt: number }>();

export async function getSecret(secretId: string): Promise<string> {
  const cached = secretCache.get(secretId);
  if (cached && cached.expiresAt > Date.now()) {
    return cached.value;
  }

  const result = await client.send(
    new GetSecretValueCommand({ SecretId: secretId })
  );

  const value = result.SecretString!;
  secretCache.set(secretId, {
    value,
    expiresAt: Date.now() + 5 * 60 * 1000, // 5 min cache
  });

  return value;
}
```

## Signed Commits and Tags

```yaml
  verify-signatures:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Verify commit signatures
        run: |
          # Check that the latest commit is signed
          if ! git verify-commit HEAD 2>/dev/null; then
            echo "::warning::Latest commit is not signed"
          fi

          # For tags: verify tag signature
          if [[ "$GITHUB_REF" == refs/tags/* ]]; then
            TAG="${GITHUB_REF#refs/tags/}"
            if ! git verify-tag "$TAG" 2>/dev/null; then
              echo "::error::Tag $TAG is not signed"
              exit 1
            fi
          fi

      - name: Enforce signed commits on PR
        if: github.event_name == 'pull_request'
        run: |
          UNSIGNED=$(git log --format='%H %G?' origin/main..HEAD | grep -v ' G$' | grep -v ' U$' || true)
          if [ -n "$UNSIGNED" ]; then
            echo "::warning::Unsigned commits found in PR:"
            echo "$UNSIGNED"
          fi
```

## Pipeline Hardening

```yaml
# Minimal permissions — always set at job level
permissions:
  contents: read  # Never write unless needed

# Pin action versions to full SHA (not tags)
steps:
  - uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.1

# Restrict workflow triggers
on:
  pull_request:
    branches: [main]
    paths:
      - 'src/**'
      - 'package.json'
      - '.github/workflows/ci.yml'
    # Ignore PRs from forks for secret-using workflows
    types: [opened, synchronize, reopened]
```

```yaml
# .github/workflows/scorecard.yml — OpenSSF Scorecard
name: Scorecard
on:
  push:
    branches: [main]
  schedule:
    - cron: '0 4 * * 1'

permissions: read-all

jobs:
  analysis:
    runs-on: ubuntu-latest
    permissions:
      security-events: write
    steps:
      - uses: actions/checkout@v4
        with:
          persist-credentials: false

      - name: Run Scorecard
        uses: ossf/scorecard-action@v2
        with:
          results_file: results.sarif
          results_format: sarif

      - name: Upload results
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif
```

## Pre-commit Hooks

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.5.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-json
      - id: check-added-large-files
        args: ['--maxkb=500']
      - id: detect-private-key
      - id: check-merge-conflict

  - repo: https://github.com/gitleaks/gitleaks
    rev: v8.18.0
    hooks:
      - id: gitleaks  # Detect secrets in commits

  - repo: local
    hooks:
      - id: npm-test
        name: Run tests
        entry: npm test
        language: system
        pass_filenames: false
        stages: [push]  # Only on push, not every commit
```

## Gotchas

1. **`pull_request` from forks can't access secrets** — For security, GitHub doesn't expose repository secrets to workflows triggered by fork PRs. If your CI needs secrets (e.g., to run integration tests), use `pull_request_target` instead — but be extremely careful because it runs with the base repo's permissions and could execute malicious PR code.

2. **Action version pinning by tag is unsafe** — `uses: actions/checkout@v4` can be silently replaced if a maintainer force-pushes a new commit to the `v4` tag. Pin to the full commit SHA: `uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11`. Use Dependabot or Renovate to automate SHA updates.

3. **GitHub-hosted runners share a public IP pool** — All GitHub-hosted runners exit through a shared set of IP ranges. IP-based allowlists (for database access, API whitelisting) are unreliable because other repos share the same IPs. Use OIDC tokens or VPN connections instead.

4. **`GITHUB_TOKEN` permissions are too broad by default** — The default token has write access to contents, packages, and more. Always set explicit `permissions` at the workflow or job level with the minimum required access. Use `permissions: read-all` at the top level and override per-job.

5. **Secret scanning doesn't catch all patterns** — GitHub's built-in secret scanning catches known provider patterns (AWS keys, GitHub tokens) but misses custom secrets, internal API keys, and database connection strings. Add Gitleaks or TruffleHog to your pipeline for broader pattern matching.

6. **`npm audit` exit codes are unreliable for CI** — `npm audit` returns exit code 1 for ANY vulnerability, including low-severity issues with no fix available. This causes unnecessary CI failures. Use `--audit-level=high` to only fail on high/critical, and `--production` to skip dev dependencies.
