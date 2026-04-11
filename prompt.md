You are implementing a Go CLI project called plain-cli. Your job is to complete exactly one task.

Read the following files before doing anything else:
- `api.md` — the Plain app API reference
- `spec.md` — the project specification and design decisions
- `todo.md` — the implementation checklist

Find the first unchecked item in `todo.md` (a line beginning with `- [ ]`). Complete that item and only that item. Do not start or partially implement any other item.

When you are done with the implementation:

1. Run `gofumpt -l -w .` and fix any formatting issues before continuing. Fix all reported issues regardless of whether you believe you caused them.
2. Run `go build ./...` and fix any issues before continuing. Fix all reported issues regardless of whether you believe you caused them.
3. Run `go vet ./...` and fix any issues before continuing. Fix all reported issues regardless of whether you believe you caused them.
4. Run `golangci-lint run ./...` and fix any issues before continuing. Repeat until the lint run is clean. Fix all reported issues regardless of whether you believe you caused them.
5. Mark the completed todo item as done by changing its `- [ ]` to `- [x]` in `todo.md`.
6. Delete any temporary files you created during implementation.
7. Stage all changes and verify there are no untracked files (`git status` should show nothing untracked). If any untracked files exist that are not part of the implementation, delete them.
8. Commit all staged changes with a clear, specific commit message describing what was implemented.
