# Contributing to MadaBank Server

## 🌿 Branching Strategy & Workflow

We follow a strict **Gitflow-inspired** workflow to ensure code quality and system stability.

| Branch | Protection Level | Purpose |
|--------|------------------|---------|
| `main` | 🔒 **Locked** | **Production**. Strictly read-only. Only merge from `staging`. |
| `staging` | 🔒 **Locked** | **QA**. Only merge from `develop`. |
| `develop` | 🛡️ **Protected** | **Integration**. Main dev branch. PR required. |
| `feat/*` | 📝 Open | New features (e.g. `feat/refresh-token`). |
| `fix/*` | 📝 Open | Bug fixes (e.g. `fix/login-error`). |

### 🛡️ Protection Rules
For `main`, `staging`, and `develop`, the following are **enforced**:
1.  **Pull Request Required**: Direct pushes are blocked.
2.  **Status Checks Must Pass**: CI (Lint, Test, Docker Build) must succeed.

---

## 📝 Pull Request Process

1.  **Create a Branch**: `git checkout -b feat/my-new-feature`
2.  **Commit Changes**: Follow the [Commit Convention](#-commit-convention) below.
3.  **Push**: `git push origin feat/my-new-feature`
4.  **Open PR**: 
    *   Features -> target `develop`
    *   Hotfixes -> target `main` (rare)
5.  **Review**: Wait for CI checks. Address feedback.
6.  **Merge**: Squash and merge is preferred to keep history clean.

---

## 💬 Commit Convention

We follow **Conventional Commits** to automate releases and changelogs.

Format: `<type>(<scope>): <description>`

*   `feat`: A new feature
    *   *Example*: `feat(auth): implement refresh token flow`
*   `fix`: A bug fix
    *   *Example*: `fix(db): resolve connection timeout on heavy load`
*   `docs`: Documentation only changes
    *   *Example*: `docs: update API Swagger definition`
*   `chore`: Maintenance, config, CI/CD (no product code change)
    *   *Example*: `chore: upgrade go version to 1.24`
*   `refactor`: Code change that neither fixes a bug nor adds a feature
    *   *Example*: `refactor: optimize transaction service logic`

---

## 📦 Releases & Tags

To release to **Production**:

1.  Ensure `staging` is merged into `main`.
2.  Create and push a semantic version tag:
    ```bash
    git checkout main
    git pull
    git tag -a v1.1.0 -m "Release v1.1.0: Production Ready"
    git push origin v1.1.0
    ```
3.  The CI/CD pipeline will detect the tag and trigger the deployment workflow (if manual approval flow is configured).

---

## 🛠️ CI/CD Pipeline

*   **CI (Continuous Integration)**: Runs on every push. Checks Code Quality (golangci-lint), Unit Tests, Security (Trivy/Gosec), and attempts a Docker Build.
*   **CD (Continuous Deployment)**: Automatically deploys to the corresponding environment (`Dev`, `Staging`, `Prod`) when code lands on the protected branch.
