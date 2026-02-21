---
description: Ensure code is formatted and tested before committing
---

Before opening a PR or committing code:

1. **Format Code**: Always run formatting to ensure gofumpt styling compliance.
```bash
make fumpt
```

2. **Run Linter**: Ensure no linting errors are introduced via golangci-lint.
```bash
make lint
```

3. **Run Tests**: Verify test coverage and pass status.
```bash
make test
```

// turbo-all
If any errors occur, fix them before moving forward with commits or PR creation!
