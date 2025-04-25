# Pull Request Submission Guidelines

To ensure the quality of the codebase and maintainability of the project, please follow these guidelines before submitting a Pull Request (PR):

## 1. PR Title and Description

- **Clear Title**: Concisely describe the main content of the PR, for example:
    - Fix: Correct error messages in user login
    - Feature: Add order export functionality

- **Detailed Description**: Include the following details in the description:
    - Purpose and background of this PR.
    - Detailed explanation of the changes.
    - For bug fixes, describe the steps to reproduce the issue.
    - For new features, explain how to use them.
    - Link related issues (if any) using keywords like `Closes #123`.

## 2. Code Checks Before Submission

- **Code Style**: Ensure the code adheres to the project's coding standards (e.g., ESLint, Prettier, or GoLint).
- **Functional Testing**: Fully test new features or bug fixes to ensure no missing functionality or regressions.
- **Unit Tests**: Write unit tests for added or modified functionality and ensure all tests pass.
- **Documentation Updates**: Update documentation if the PR includes new features or API changes.

## 3. Branch Strategy

- **Correct Branch**:
    - Develop new features based on `feature/*` branches.
    - Fix bugs based on `fix/*` branches.
    - Ensure the target branch of the PR aligns with the project's branching strategy.

- **Sync with Base Branch**: Before submitting the PR, ensure your branch is up-to-date with the target branch (e.g., `main` or `develop`).

## 4. Review Process

- **Small Commits**: Avoid submitting excessive changes in a single PR; break it into smaller logical units.

---

Thank you for your contribution!
