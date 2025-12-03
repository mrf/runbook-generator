# TEST123

## Overview

This runbook covers: Git version control, Git branching. It contains 70 steps with 96 commands total.

## Prerequisites

- PostgreSQL client installed with database access
- AWS CLI installed and configured
- Make build tool installed
- Git CLI installed

## Steps

### Step 1: exit operations

```bash
exit
```

### Step 2: Git operations

```bash
git status
```

### Step 3: make operations

```bash
make build
```

### Step 4: File system operations

```bash
cd ../claude-code-router
```

### Step 5: ../branch-cleaner/bin/branch-cleaner operations

```bash
../branch-cleaner/bin/branch-cleaner
```

### Step 6: Commit and push changes

```bash
git status
git add index.js
git commit -m "More debug code"
git push
```

**Why:** git-commit

### Step 7: Git operations

```bash
git status
```

### Step 8: File system operations

```bash
cd ~/Projects/branch-cleaner
```

### Step 9: File system operations

```bash
ls
```

### Step 10: Commit and push changes

```bash
git status
git add *
```

**Why:** git-commit

### Step 11: Commit and push changes

```bash
git status
git commit -m "Simplify commit history loading to minimize data leakage and api cost"
git push
```

**Why:** git-commit

### Step 12: Git operations

```bash
git log -p
```

### Step 13: Git operations

```bash
git status
git diff
```

### Step 14: File system operations

```bash
rm .tool-versions
```

### Step 15: Git operations

```bash
git status
git diff
git status
```

### Step 16: File system operations

```bash
rm -rf deployment/modules/infrastructure/datadog_api_performance
```

### Step 17: Git operations

```bash
git status
git diff deployment/
git status
git diff deployment/main.tf
```

### Step 18: vim operations

```bash
vim deployment/main.tf
```

### Step 19: Git operations

```bash
git status
git diff deployment/staging/
git diff deployment/
```

### Step 20: Git operations

```bash
git diff
```

### Step 21: Commit and push changes

```bash
git status
git add deployment/
```

**Why:** git-commit

### Step 22: Branch management

```bash
git checkout -b mrf/CLOUDINFRA-1404-uptime-e2e-correct-fleetid
```

**Why:** git-branch

### Step 23: Commit and push changes

```bash
git status
git commit -m "CLOUDINFRA-1404: Correct fleet id for staging and production."
```

**Why:** git-commit

### Step 24: File system operations

```bash
cd ~/Projects/scratch
```

### Step 25: export operations

```bash
export AWS_PROFILE=staging-us
```

### Step 26: aws operations

```bash
aws rds describe-db-clusters --output text --query "DBClusters[?starts_with(DatabaseName,'db_ai_hayden_private')].Endpoint"
```

### Step 27: vim operations

```bash
vim ~/.aws/credentials
vim ~/.aws/config
```

### Step 28: export operations

```bash
export AWS_PROFILE=stage-us
```

### Step 29: aws operations

```bash
aws rds describe-db-clusters --output text --query "DBClusters[?starts_with(DatabaseName,'db_ai_hayden_private')].Endpoint"
aws sso login --profile production-us
```

### Step 30: aws operations

```bash
aws rds describe-db-clusters --output text --query "DBClusters[?starts_with(DatabaseName,'db_ai_hayden_private')].Endpoint"
```

### Step 31: DB_PASSWORD=$(aws operations

```bash
DB_PASSWORD=$(aws ssm get-parameter --name db_password --with-decryption --query "Parameter.Value" --output text)\
```

### Step 32: psql operations

```bash
psql "postgresql://masterstaging:<REDACTED>@db-ai-hayden-private-staging.cluster-cmg3erok1okq.us-west-2.rds.amazonaws.com:5432/db_ai_hayden_private_staging?sslmode=require"\
```

### Step 33: psql operations

```bash
psql "postgresql://masterstaging:<REDACTED>@vpn-proxy.rds.staging.console.hayden.ai:5432/db_ai_hayden_private_staging?sslmode=require"\
```

### Step 34: env operations

```bash
env | grep DB
```

### Step 35: aws operations

```bash
aws ssm get-parameter --name db_password --with-decryption --query "Parameter.Value" --output text
```

### Step 36: export operations

```bash
export DB_PASSWORD='<REDACTED>'\
```

### Step 37: psql operations

```bash
psql "postgresql://masterstaging:<REDACTED>@vpn-proxy.rds.staging.console.hayden.ai:5432/db_ai_hayden_private_staging?sslmode=require"\
```

### Step 38: env operations

```bash
env | grep DB
```

### Step 39: psql operations

```bash
psql "postgresql://masterstaging:<REDACTED>@vpn-proxy.rds.staging.console.hayden.ai:5432/db_ai_hayden_private_staging?sslmode=require"\
```

### Step 40: export operations

```bash
export PGPASSWORD="<REDACTED> rds generate-db-auth-token --profile "$AWS_PROFILE" --hostname "$PGENDPOINT" --port 5432 --region us-west-2 --username "$PGUSER")"\
```

### Step 41: env operations

```bash
env | grep PG
```

### Step 42: export operations

```bash
export PGENDPOINT="vpn-proxy.rds.staging.console.hayden.ai"
export PGENV="staging"
export PGUSER="masterstaging"
```

### Step 43: psql operations

```bash
psql
psql "host=$PGENDPOINT port=5432 sslmode=require dbname=db_ai_hayden_private_$PGENV user=$PGUSER password=<REDACTED>"
```

### Step 44: export operations

```bash
export PGENDPOINT=readonly.endpoint.proxy-cmg3erok1okq.us-west-2.rds.amazonaws.com
```

### Step 45: psql operations

```bash
psql "host=$PGENDPOINT port=5432 sslmode=require dbname=db_ai_hayden_private_$PGENV user=$PGUSER password=<REDACTED>"
```

### Step 46: env operations

```bash
env | grep PG
```

### Step 47: export operations

```bash
export PGENV="staging"\
```

### Step 48: psql operations

```bash
psql "host=$PGENDPOINT port=5432 sslmode=require dbname=db_ai_hayden_private_$PGENV user=$PGUSER password=<REDACTED>"
```

### Step 49: export operations

```bash
export PGENV="staging"\
```

### Step 50: psql operations

```bash
psql
```

### Step 51: export operations

```bash
export PGENDPOINT="vpn-proxy.rds.staging.console.hayden.ai"
export PGPASSWORD="<REDACTED>"
```

### Step 52: psql operations

```bash
psql "host=$PGENDPOINT port=5432 sslmode=require dbname=db_ai_hayden_private_$PGENV user=$PGUSER password=<REDACTED>"
```

### Step 53: aws operations

```bash
aws rds describe-db-clusters --output text --query "DBClusters[?starts_with(DatabaseName,'db_ai_hayden_private')].Endpoint"
```

### Step 54: export operations

```bash
export PGENDPOINT="db-ai-hayden-private-staging.cluster-cmg3erok1okq.us-west-2.rds.amazonaws.com"
```

### Step 55: psql operations

```bash
psql
psql "host=$PGENDPOINT port=5432 sslmode=require dbname=db_ai_hayden_private_$PGENV user=$PGUSER password=<REDACTED>"
```

### Step 56: env operations

```bash
env | grep PG
```

### Step 57: psql operations

```bash
psql "host=$PGENDPOINT port=5432 sslmode=require gssencmode=disable dbname=db_ai_hayden_private_$PGENV user=$PGUSER password=<REDACTED>"
```

### Step 58: psql operations

```bash
psql "host=$PGENDPOINT port=5432 sslmode=require gssencmode=disable dbname=db_ai_hayden_private_$PGENV user=$PGUSER" -W
```

### Step 59: File system operations

```bash
ls
```

### Step 60: Git operations

```bash
git status
git log -p
```

### Step 61: Commit and push changes

```bash
git push -u origin mrf/CLOUDINFRA-1404-uptime-e2e-correct-fleetid\
```

**Why:** git-commit

### Step 62: File system operations

```bash
cd ..
```

### Step 63: File system operations

```bash
ls
```

### Step 64: File system operations

```bash
cd ..
```

### Step 65: File system operations

```bash
ls
```

### Step 66: File system operations

```bash
mkdir runbook-generator
```

### Step 67: File system operations

```bash
cd runbook-generator
```

### Step 68: history operations

```bash
history
```

### Step 69: claude operations

```bash
claude
claude "Write me an ARCHITECTURE.md file for a tool that looks through local bash history between two provided timestamps or history numbers and then provides a detailed runbook document in markdown format so that someone else could follow those sames steps, tool should remove duplication and infer intent, tool should not include passwords or other sensitive authentication data"
```

### Step 70: Git operations

```bash
git config user.email "markferree@gmail.com"
git init
```

## Notes

- Generated from bash history on 2025-12-03 14:35:00
- Time range: commands #10000 to #10100
- Commands sanitized: 13
