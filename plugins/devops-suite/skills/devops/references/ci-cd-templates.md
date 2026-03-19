# CI/CD Pipeline Templates

Copy-paste ready CI/CD pipeline configurations for GitHub Actions and GitLab CI. Each template is production-tested and follows best practices for caching, security, and deployment.

## GitHub Actions Templates

### Node.js Full Pipeline (Lint → Test → Build → Deploy)

```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: ${{ github.ref != 'refs/heads/main' }}

permissions:
  contents: read

jobs:
  lint:
    name: Lint & Format
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm

      - run: npm ci

      - name: ESLint
        run: npx eslint . --max-warnings 0

      - name: Prettier
        run: npx prettier --check .

      - name: TypeScript
        run: npx tsc --noEmit

  test:
    name: Test
    runs-on: ubuntu-latest
    needs: [lint]
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: testdb
        ports:
          - 5432:5432
        options: >-
          --health-cmd "pg_isready -U test"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm

      - run: npm ci

      - name: Run Migrations
        run: npx drizzle-kit push
        env:
          DATABASE_URL: postgresql://test:test@localhost:5432/testdb

      - name: Unit Tests
        run: npx vitest run --coverage
        env:
          DATABASE_URL: postgresql://test:test@localhost:5432/testdb
          REDIS_URL: redis://localhost:6379
          NODE_ENV: test

      - name: Upload Coverage
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: coverage-report
          path: coverage/
          retention-days: 5

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [test]
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm

      - run: npm ci
      - run: npm run build

      - uses: actions/upload-artifact@v4
        with:
          name: build-output
          path: dist/
          retention-days: 3

  deploy-staging:
    name: Deploy Staging
    runs-on: ubuntu-latest
    needs: [build]
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    environment:
      name: staging
      url: https://staging.example.com
    permissions:
      contents: read
      id-token: write
    steps:
      - uses: actions/checkout@v4

      - uses: actions/download-artifact@v4
        with:
          name: build-output
          path: dist/

      - name: Deploy
        run: |
          echo "Deploy to staging environment"
          # heroku container:push web --app my-app-staging
          # vercel --token ${{ secrets.VERCEL_TOKEN }}
          # fly deploy --app my-app-staging
        env:
          DEPLOY_TOKEN: ${{ secrets.STAGING_DEPLOY_TOKEN }}

      - name: Smoke Test
        run: |
          sleep 15
          for i in 1 2 3 4 5; do
            if curl -sf https://staging.example.com/health; then
              echo "Health check passed"
              exit 0
            fi
            echo "Attempt $i failed, retrying in 5s..."
            sleep 5
          done
          echo "Smoke test failed"
          exit 1

  deploy-production:
    name: Deploy Production
    runs-on: ubuntu-latest
    needs: [build]
    if: startsWith(github.ref, 'refs/tags/v')
    environment:
      name: production
      url: https://example.com
    permissions:
      contents: read
      id-token: write
    steps:
      - uses: actions/checkout@v4

      - uses: actions/download-artifact@v4
        with:
          name: build-output
          path: dist/

      - name: Deploy
        run: echo "Deploy to production"
        env:
          DEPLOY_TOKEN: ${{ secrets.PRODUCTION_DEPLOY_TOKEN }}

      - name: Smoke Test
        run: |
          sleep 20
          curl -sf https://example.com/health || exit 1
```

### Python (FastAPI/Django) Pipeline

```yaml
# .github/workflows/ci.yml
name: Python CI/CD

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

jobs:
  lint:
    name: Lint & Type Check
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-python@v5
        with:
          python-version: "3.12"
          cache: pip

      - run: pip install ruff mypy

      - name: Ruff Format Check
        run: ruff format --check .

      - name: Ruff Lint
        run: ruff check .

      - name: Type Check
        run: mypy src/ --ignore-missing-imports

  test:
    name: Test (Python ${{ matrix.python-version }})
    runs-on: ubuntu-latest
    needs: [lint]
    strategy:
      fail-fast: false
      matrix:
        python-version: ["3.11", "3.12"]
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: test
        ports:
          - 5432:5432
        options: >-
          --health-cmd "pg_isready -U test"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-python@v5
        with:
          python-version: ${{ matrix.python-version }}
          cache: pip

      - run: pip install -r requirements.txt -r requirements-dev.txt

      - name: Run Tests
        run: pytest --cov=src --cov-report=xml -v
        env:
          DATABASE_URL: postgresql://test:test@localhost:5432/test

      - name: Upload Coverage
        if: matrix.python-version == '3.12'
        uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: coverage.xml

  build-docker:
    name: Build Docker Image
    runs-on: ubuntu-latest
    needs: [test]
    if: github.ref == 'refs/heads/main'
    permissions:
      packages: write
    steps:
      - uses: actions/checkout@v4

      - uses: docker/setup-buildx-action@v3

      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ghcr.io/${{ github.repository }}:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

### Docker Build & Push to Multiple Registries

```yaml
# .github/workflows/docker.yml
name: Docker Build & Push

on:
  push:
    branches: [main]
    tags: ["v*"]

jobs:
  docker:
    name: Build & Push
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4

      - uses: docker/setup-qemu-action@v3

      - uses: docker/setup-buildx-action@v3

      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - uses: docker/login-action@v3
        if: startsWith(github.ref, 'refs/tags/v')
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}

      - uses: docker/metadata-action@v5
        id: meta
        with:
          images: |
            ghcr.io/${{ github.repository }}
            ${{ startsWith(github.ref, 'refs/tags/v') && format('{0}/{1}', secrets.DOCKERHUB_USERNAME, github.event.repository.name) || '' }}
          tags: |
            type=ref,event=branch
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha,prefix=

      - uses: docker/build-push-action@v5
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            APP_VERSION=${{ github.sha }}
```

### Terraform Infrastructure Pipeline

```yaml
# .github/workflows/terraform.yml
name: Terraform

on:
  push:
    branches: [main]
    paths:
      - "infra/**"
  pull_request:
    paths:
      - "infra/**"

permissions:
  contents: read
  pull-requests: write
  id-token: write

jobs:
  plan:
    name: Terraform Plan
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: infra
    steps:
      - uses: actions/checkout@v4

      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: "1.7"
          terraform_wrapper: false

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: us-east-1

      - name: Terraform Init
        run: terraform init

      - name: Terraform Format
        run: terraform fmt -check -recursive

      - name: Terraform Validate
        run: terraform validate

      - name: Terraform Plan
        id: plan
        run: terraform plan -no-color -out=tfplan
        continue-on-error: true

      - name: Comment PR
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v7
        with:
          script: |
            const output = `#### Terraform Plan 📖\`${{ steps.plan.outcome }}\`

            <details><summary>Show Plan</summary>

            \`\`\`terraform
            ${{ steps.plan.outputs.stdout }}
            \`\`\`

            </details>

            *Pushed by: @${{ github.actor }}, Action: \`${{ github.event_name }}\`*`;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: output
            });

      - name: Upload Plan
        uses: actions/upload-artifact@v4
        with:
          name: tfplan
          path: infra/tfplan

  apply:
    name: Terraform Apply
    runs-on: ubuntu-latest
    needs: [plan]
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    environment: production
    defaults:
      run:
        working-directory: infra
    steps:
      - uses: actions/checkout@v4

      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: "1.7"

      - uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: us-east-1

      - uses: actions/download-artifact@v4
        with:
          name: tfplan
          path: infra/

      - run: terraform init
      - run: terraform apply -auto-approve tfplan
```

### Matrix Testing (Multi-Version, Multi-OS)

```yaml
# .github/workflows/matrix.yml
name: Matrix Tests

on:
  pull_request:
    branches: [main]

jobs:
  test:
    name: Test (${{ matrix.os }}, Node ${{ matrix.node }})
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        node: [18, 20, 22]
        exclude:
          # Skip older Node on Windows (slower, less common)
          - os: windows-latest
            node: 18
          # Skip macOS for Node 18 (save costs)
          - os: macos-latest
            node: 18
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node }}
          cache: npm

      - run: npm ci

      - name: Run Tests
        run: npm test
        env:
          CI: true
```

### Reusable Workflow (Called by Other Workflows)

```yaml
# .github/workflows/reusable-deploy.yml
name: Reusable Deploy

on:
  workflow_call:
    inputs:
      environment:
        required: true
        type: string
      url:
        required: true
        type: string
    secrets:
      deploy_token:
        required: true

jobs:
  deploy:
    name: Deploy to ${{ inputs.environment }}
    runs-on: ubuntu-latest
    environment:
      name: ${{ inputs.environment }}
      url: ${{ inputs.url }}
    steps:
      - uses: actions/checkout@v4

      - uses: actions/download-artifact@v4
        with:
          name: build-output
          path: dist/

      - name: Deploy
        run: echo "Deploying to ${{ inputs.environment }}"
        env:
          DEPLOY_TOKEN: ${{ secrets.deploy_token }}

      - name: Smoke Test
        run: |
          sleep 15
          curl -sf ${{ inputs.url }}/health || exit 1
```

Caller workflow:
```yaml
# .github/workflows/ci.yml
jobs:
  deploy-staging:
    needs: [build]
    uses: ./.github/workflows/reusable-deploy.yml
    with:
      environment: staging
      url: https://staging.example.com
    secrets:
      deploy_token: ${{ secrets.STAGING_DEPLOY_TOKEN }}
```

### Release Pipeline (Semantic Versioning)

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    name: Create Release
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Generate Changelog
        id: changelog
        run: |
          PREVIOUS_TAG=$(git tag --sort=-version:refname | head -2 | tail -1)
          if [ -z "$PREVIOUS_TAG" ]; then
            CHANGELOG=$(git log --pretty=format:"- %s (%h)" HEAD)
          else
            CHANGELOG=$(git log --pretty=format:"- %s (%h)" ${PREVIOUS_TAG}..HEAD)
          fi
          echo "changelog<<EOF" >> $GITHUB_OUTPUT
          echo "$CHANGELOG" >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT

      - name: Create Release
        uses: actions/github-script@v7
        with:
          script: |
            await github.rest.repos.createRelease({
              owner: context.repo.owner,
              repo: context.repo.repo,
              tag_name: '${{ github.ref_name }}',
              name: 'Release ${{ github.ref_name }}',
              body: `## Changes\n\n${{ steps.changelog.outputs.changelog }}`,
              draft: false,
              prerelease: ${{ contains(github.ref_name, '-') }},
            });
```

### Security Scanning

```yaml
# .github/workflows/security.yml
name: Security Scan

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: "0 6 * * 1"  # Weekly Monday 6am UTC

permissions:
  security-events: write
  contents: read

jobs:
  codeql:
    name: CodeQL Analysis
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: github/codeql-action/init@v3
        with:
          languages: javascript-typescript

      - uses: github/codeql-action/analyze@v3

  dependency-audit:
    name: Dependency Audit
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm

      - run: npm ci
      - run: npm audit --audit-level=high

  trivy:
    name: Trivy Container Scan
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4

      - uses: docker/build-push-action@v5
        with:
          context: .
          push: false
          tags: scan-target:latest
          load: true

      - uses: aquasecurity/trivy-action@master
        with:
          image-ref: scan-target:latest
          format: sarif
          output: trivy-results.sarif
          severity: CRITICAL,HIGH

      - uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: trivy-results.sarif
```

## GitLab CI Templates

### Full Node.js Pipeline

```yaml
# .gitlab-ci.yml
stages:
  - lint
  - test
  - build
  - deploy

variables:
  NODE_VERSION: "20"
  POSTGRES_USER: test
  POSTGRES_PASSWORD: test
  POSTGRES_DB: test
  npm_config_cache: "$CI_PROJECT_DIR/.npm"

# Global cache for npm
cache:
  key: "${CI_COMMIT_REF_SLUG}-npm"
  paths:
    - .npm/
  policy: pull

# Lint stage
lint:
  stage: lint
  image: node:${NODE_VERSION}-slim
  cache:
    key: "${CI_COMMIT_REF_SLUG}-npm"
    paths:
      - .npm/
    policy: pull-push
  script:
    - npm ci --cache .npm
    - npx eslint .
    - npx prettier --check .
    - npx tsc --noEmit
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_COMMIT_BRANCH == "main"

# Test stage with service containers
test:
  stage: test
  image: node:${NODE_VERSION}-slim
  services:
    - name: postgres:16-alpine
      alias: postgres
    - name: redis:7-alpine
      alias: redis
  variables:
    DATABASE_URL: "postgresql://test:test@postgres:5432/test"
    REDIS_URL: "redis://redis:6379"
    NODE_ENV: test
  script:
    - npm ci --cache .npm
    - npx vitest run --coverage
  coverage: '/All files[^|]*\|[^|]*\s+([\d\.]+)/'
  artifacts:
    when: always
    reports:
      junit: junit.xml
      coverage_report:
        coverage_format: cobertura
        path: coverage/cobertura-coverage.xml
    paths:
      - coverage/
    expire_in: 7 days
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_COMMIT_BRANCH == "main"

# Build Docker image
build:
  stage: build
  image: docker:24
  services:
    - docker:24-dind
  variables:
    DOCKER_TLS_CERTDIR: "/certs"
  before_script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
  script:
    - docker build
      --cache-from $CI_REGISTRY_IMAGE:latest
      --tag $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
      --tag $CI_REGISTRY_IMAGE:latest
      --build-arg BUILDKIT_INLINE_CACHE=1
      .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - docker push $CI_REGISTRY_IMAGE:latest
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
    - if: $CI_COMMIT_TAG

# Deploy to staging
deploy-staging:
  stage: deploy
  image: alpine:latest
  environment:
    name: staging
    url: https://staging.example.com
  script:
    - echo "Deploy $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA to staging"
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
  when: on_success

# Deploy to production (manual)
deploy-production:
  stage: deploy
  image: alpine:latest
  environment:
    name: production
    url: https://example.com
  script:
    - echo "Deploy $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA to production"
  rules:
    - if: $CI_COMMIT_TAG =~ /^v\d+\.\d+\.\d+$/
  when: manual
  allow_failure: false
```

### GitLab CI with Review Apps

```yaml
review:
  stage: deploy
  image: alpine:latest
  environment:
    name: review/$CI_COMMIT_REF_SLUG
    url: https://$CI_COMMIT_REF_SLUG.review.example.com
    on_stop: stop-review
    auto_stop_in: 1 week
  script:
    - echo "Deploy review app for $CI_COMMIT_REF_SLUG"
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"

stop-review:
  stage: deploy
  image: alpine:latest
  environment:
    name: review/$CI_COMMIT_REF_SLUG
    action: stop
  script:
    - echo "Tear down review app for $CI_COMMIT_REF_SLUG"
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
      when: manual
  allow_failure: true
```

## Caching Strategies Reference

### GitHub Actions Cache Keys

```yaml
# Best practice: hash lockfile for cache key
cache:
  key: npm-${{ runner.os }}-${{ hashFiles('**/package-lock.json') }}
  restore-keys: |
    npm-${{ runner.os }}-

# Turborepo cache
cache:
  key: turbo-${{ runner.os }}-${{ hashFiles('**/pnpm-lock.yaml') }}-${{ github.sha }}
  restore-keys: |
    turbo-${{ runner.os }}-${{ hashFiles('**/pnpm-lock.yaml') }}-
    turbo-${{ runner.os }}-

# Docker layer cache with BuildKit
cache-from: type=gha
cache-to: type=gha,mode=max
```

### Dependency Cache Sizes (Typical)

| Package Manager | Typical Cache Size | Cache Location |
|----------------|-------------------|----------------|
| npm | 50-200MB | ~/.npm |
| pnpm | 100-300MB | ~/.local/share/pnpm/store |
| pip | 20-100MB | ~/.cache/pip |
| Go modules | 50-200MB | ~/go/pkg/mod |
| Cargo | 200-500MB | ~/.cargo/registry |
| Maven | 100-400MB | ~/.m2/repository |
| Gradle | 100-300MB | ~/.gradle/caches |

## Deployment Stage Templates

### Heroku Deploy

```yaml
deploy-heroku:
  steps:
    - uses: akhileshns/heroku-deploy@v3.13.15
      with:
        heroku_api_key: ${{ secrets.HEROKU_API_KEY }}
        heroku_app_name: my-app
        heroku_email: ${{ secrets.HEROKU_EMAIL }}
```

### Vercel Deploy

```yaml
deploy-vercel:
  steps:
    - name: Install Vercel CLI
      run: npm install -g vercel

    - name: Pull Vercel Settings
      run: vercel pull --yes --environment=production --token=${{ secrets.VERCEL_TOKEN }}

    - name: Build
      run: vercel build --prod --token=${{ secrets.VERCEL_TOKEN }}

    - name: Deploy
      run: vercel deploy --prebuilt --prod --token=${{ secrets.VERCEL_TOKEN }}
```

### Fly.io Deploy

```yaml
deploy-fly:
  steps:
    - uses: superfly/flyctl-actions/setup-flyctl@master

    - run: flyctl deploy --remote-only
      env:
        FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}
```

### AWS ECS Deploy

```yaml
deploy-ecs:
  steps:
    - uses: aws-actions/configure-aws-credentials@v4
      with:
        role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
        aws-region: us-east-1

    - uses: aws-actions/amazon-ecr-login@v2
      id: ecr

    - name: Push to ECR
      run: |
        docker tag myapp:latest ${{ steps.ecr.outputs.registry }}/myapp:${{ github.sha }}
        docker push ${{ steps.ecr.outputs.registry }}/myapp:${{ github.sha }}

    - name: Update ECS Service
      run: |
        aws ecs update-service \
          --cluster my-cluster \
          --service my-service \
          --force-new-deployment
```

### Cloudflare Workers Deploy

```yaml
deploy-cloudflare:
  steps:
    - name: Deploy Worker
      uses: cloudflare/wrangler-action@v3
      with:
        apiToken: ${{ secrets.CLOUDFLARE_API_TOKEN }}
        command: deploy
```

### Railway Deploy

```yaml
deploy-railway:
  steps:
    - name: Install Railway CLI
      run: npm install -g @railway/cli

    - name: Deploy
      run: railway up --detach
      env:
        RAILWAY_TOKEN: ${{ secrets.RAILWAY_TOKEN }}
```

## Monorepo Pipeline Patterns

### Turborepo with Path Filtering

```yaml
name: Monorepo CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  detect-changes:
    runs-on: ubuntu-latest
    outputs:
      api: ${{ steps.filter.outputs.api }}
      web: ${{ steps.filter.outputs.web }}
      packages: ${{ steps.filter.outputs.packages }}
    steps:
      - uses: actions/checkout@v4
      - uses: dorny/paths-filter@v3
        id: filter
        with:
          filters: |
            api:
              - 'apps/api/**'
              - 'packages/shared/**'
              - 'packages/db/**'
            web:
              - 'apps/web/**'
              - 'packages/ui/**'
              - 'packages/shared/**'
            packages:
              - 'packages/**'

  test-api:
    needs: detect-changes
    if: needs.detect-changes.outputs.api == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with:
          version: 9
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: pnpm
      - run: pnpm install --frozen-lockfile
      - run: pnpm turbo test --filter=api...

  test-web:
    needs: detect-changes
    if: needs.detect-changes.outputs.web == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with:
          version: 9
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: pnpm
      - run: pnpm install --frozen-lockfile
      - run: pnpm turbo test --filter=web...

  build-all:
    needs: [test-api, test-web]
    if: always() && !contains(needs.*.result, 'failure')
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with:
          version: 9
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: pnpm
      - name: Turbo Cache
        uses: actions/cache@v4
        with:
          path: .turbo
          key: turbo-${{ runner.os }}-${{ hashFiles('**/pnpm-lock.yaml') }}-${{ github.sha }}
          restore-keys: |
            turbo-${{ runner.os }}-${{ hashFiles('**/pnpm-lock.yaml') }}-
            turbo-${{ runner.os }}-
      - run: pnpm install --frozen-lockfile
      - run: pnpm turbo build
```

## Notification Templates

### Slack Notification on Failure

```yaml
notify-failure:
  runs-on: ubuntu-latest
  needs: [deploy-production]
  if: failure()
  steps:
    - name: Slack Notification
      uses: 8398a7/action-slack@v3
      with:
        status: failure
        fields: repo,message,commit,author,action,eventName,ref,workflow
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

### Custom Slack Message

```yaml
- name: Post to Slack
  run: |
    curl -X POST ${{ secrets.SLACK_WEBHOOK }} \
      -H 'Content-type: application/json' \
      -d '{
        "blocks": [
          {
            "type": "section",
            "text": {
              "type": "mrkdwn",
              "text": "✅ *${{ github.repository }}* deployed to production\n*Version*: `${{ github.ref_name }}`\n*By*: ${{ github.actor }}"
            }
          }
        ]
      }'
```
