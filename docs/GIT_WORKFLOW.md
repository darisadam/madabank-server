# Git Workflow & Branch Protection 🛡️

This repository follows a strict **Gitflow** inspired workflow to ensure code quality and stability.

## 🌿 Branching Strategy

| Branch | Protection Level | Purpose |
|--------|------------------|---------|
| `main` | 🔒 **Locked** | Production code. Strictly read-only. Only merge from `staging` via PR. |
| `staging` | 🔒 **Locked** | QA environment. Only merge from `develop` via PR. |
| `develop` | 🛡️ **Protected** | Main development branch. PR required from feature branches. |
| `feat/*` | 📝 Open | New features (e.g. `feat/refresh-token`). |
| `fix/*` | 📝 Open | Bug fixes (e.g. `fix/login-error`). |

## 🚫 Branch Protection Rules (How to Set Up)
Go to **Settings -> Branches -> Add Rule** for `main`, `staging`, `develop`:

1. **Require a pull request before merging**
   - [x] Require approvals (1)
2. **Require status checks to pass before merging**
   - [x] `Build & Push Docker Image` (CI)
   - [x] `test` (Unit Tests)
3. **Do not allow bypassing the above settings**

This ensures no one (including you) can accidentally push broken code to critical branches.

## 💬 Commit Convention
We follow **Conventional Commits**:

- `feat: add biometric login support`
- `fix: resolve db connection timeout`
- `docs: update api reference`
- `chore: update github actions versions`
- `refactor: optimize transaction logic`

## 🏷️ Release Process
To release to **Production**:

1. Merge `staging` into `main`.
2. Create a tag:
   ```bash
   git checkout main
   git pull origin main
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
3. The CI/CD pipeline will automatically deploy to AWS Production.
