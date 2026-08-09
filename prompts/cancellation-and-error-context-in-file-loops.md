---
status: draft
created: "2026-08-09T19:15:00Z"
---

<summary>
- Two loops walk every markdown file in a vault doing real disk work per file
- Neither checks whether the caller has cancelled, so a cancelled run keeps reading
- The cancellation value is already available in both functions, just unused
- One of the loops also reports failures without saying which file failed
- No behaviour change on the success path
</summary>

<objective>
Both vault-file loops stop promptly when the caller cancels, and the indexing loop's errors name the file that failed.
</objective>

<context>
Read `CLAUDE.md` for project conventions if present (this repo has none; conventions come from sibling bborbe repos and the code below).

Files to read before making changes (read ALL first):
- `pkg/index/index.go` — `indexBuilder.Build`; the loop is `for _, file := range files` and does `os.ReadFile` then `b.parser.ParseAliases(ctx, ...)` per file
- `pkg/validator/validator.go` — `validator.Validate`; the loop is `for _, file := range files` and calls `v.parser.ParseFile(ctx, file)` per file

Both functions already take `ctx context.Context` as their first parameter. Both already import `github.com/bborbe/errors`.

Look at how the rest of this repo writes a cancellation guard before inventing one; match the existing style if a precedent exists.
</context>

<requirements>
1. In `pkg/index/index.go`, at the very top of the `for _, file := range files` loop body in `Build`, add a cancellation check that returns a wrapped `ctx.Err()` when the context is done.
2. In `pkg/validator/validator.go`, add the same check at the top of the `for _, file := range files` loop body in `Validate`.
3. Use a non-blocking `select` with a `default:` branch so the success path is unaffected:
   ```go
   select {
   case <-ctx.Done():
       return nil, errors.Wrap(ctx, ctx.Err(), "context cancelled")
   default:
   }
   ```
   Match each function's actual return signature — both return `(T, error)`, so `return nil, …` is correct, but verify rather than assume.
4. In `pkg/index/index.go` only, change the two error wraps inside that loop from `errors.Wrap` to `errors.Wrapf` so they name the file:
   - `errors.Wrap(ctx, err, "read file failed")` → `errors.Wrapf(ctx, err, "read file %s failed", file)`
   - `errors.Wrap(ctx, err, "parse aliases failed")` → `errors.Wrapf(ctx, err, "parse aliases %s failed", file)`
5. Do NOT touch the inner loops (`index.go` over `aliases`, `validator.go` over `links`). They are bounded and do no I/O, and the outer guard already covers them.
6. Do NOT change the `errors.Wrap` in `pkg/validator/validator.go` — it was not part of this finding. Leave it exactly as it is.
7. Add a bullet under `## Unreleased` in `CHANGELOG.md` using a conventional prefix. **The file currently has no `## Unreleased` section** — its top entry is `## v0.2.10`. Create the section directly above that entry; do not rename or edit the released one.
</requirements>

<constraints>
- Only change `pkg/index/index.go`, `pkg/validator/validator.go`, and `CHANGELOG.md`
- Do NOT commit — dark-factory handles git
- Existing tests must still pass
- Use `errors.Wrap`/`errors.Wrapf` from `github.com/bborbe/errors` — never `fmt.Errorf` or a bare `return err`
- Do not add a new dependency and do not restructure either function
</constraints>

<verification>
make precommit
</verification>
