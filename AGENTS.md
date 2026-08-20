# Project Guidelines & Development Workflow

## Git Branch & PR Workflow Policy

- The main branch is protected and locked against direct pushes.
- **All future contributions, bug fixes, and features must:**
  1. Be created on a dedicated branch (e.g., dev, eat/..., ix/...).
  2. Be tested and verified (go test -v ./... and go vet ./...).
  3. Be pushed to the remote repository.
  4. Have a Pull Request (PR) opened targeting main (or dev if using GitFlow).

## Conventional Commits

Use standard conventional commit prefixes:
- eat: New features
- ix: Bug fixes
- docs: Documentation updates
- efactor: Code refactoring
- 	est: Adding or updating tests
- ci: CI/CD workflow updates