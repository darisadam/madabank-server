# Contributing to MadaBank Server

## Branching Strategy

We follow a feature-branch workflow. Direct commits to `main` are protected.

- **`main`**: Production-ready code. Auto-deploys to Production environment.
- **`develop`**: Integration branch. Auto-deploys to Dev environment.
- **`staging`**: Pre-production branch. Auto-deploys to Staging environment.

### Feature Branches
Create branches from `develop` for new work:
- `feat/feature-name` (New features)
- `fix/bug-name` (Bug fixes)
- `chore/task-name` (Maintenance, config, docs)

## Pull Request Process

1.  **Create a Branch**: `git checkout -b feat/my-feature`
2.  **Commit Changes**: Keep commits atomic and messages descriptive.
3.  **Push**: `git push origin feat/my-feature`
4.  **Open PR**: Target `develop` for features/fixes.
5.  **Review**: Wait for CI checks (Lint, Test, Docker Build) to pass.
6.  **Merge**: Squash and merge is preferred.

## Releases & Tags

To release to Production manually (rollback/hotfix) or mark a version:
1.  Tag the commit on `main`: `git tag v1.0.0`
2.  Push tag: `git push origin v1.0.0`
3.  This triggers the Manual Production Deploy workflow.

## CI/CD Pipeline

- **CI**: Runs on every Push/PR. Checks Code Quality, Tests, Security, and Docker Build.
- **CD**: Runs on push to `develop` (Dev), `staging` (Staging), `main` (Prod).
