# create-saas-app

A CLI tool that scaffolds a new multitenant Go + React SaaS project from the
[saas-starter](https://github.com/edsonmubezi/saas-starter) template in one
command.

## Install

```sh
go install github.com/edsonmubezi/saas-starter/create-saas-app@latest
```

## Usage

```sh
create-saas-app init <slug> [flags]
```

### Positional argument

| Argument | Description |
|---|---|
| `<slug>` | Project identifier — lowercase letters, digits, and hyphens; no leading/trailing hyphens (e.g. `todo-app`) |

### Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--display-name` | `-d` | Title-cased slug | Human-readable project name shown in the UI and emails |
| `--github-user` | `-g` | prompted | GitHub username used in the Go module path |
| `--author` | `-a` | prompted | Author or company name |
| `--template` | | `https://github.com/edsonmubezi/saas-starter` | Template repository to clone |
| `--output-dir` | | `./<slug>` | Destination directory |

### Example

```sh
create-saas-app init todo-app \
  --display-name "Todo App" \
  --github-user johndoe \
  --author "John Doe"
```

If `--github-user` or `--author` are omitted the CLI prompts for them
interactively.

## What it does

1. **Clones** `saas-starter` (shallow clone, `--depth=1`)
2. **Deletes** template-only files (`.env`, `go.sum`, `frontend/node_modules`, `.git`, `create-saas-app/`, etc.)
3. **Renames** every template string across all Go, TypeScript, YAML, JSON, Markdown, and config files:
   - `saas-starter` → `<slug>`
   - `github.com/edsonmubezi/myapp` → `github.com/<github-user>/<slug>`
   - `SaaS Starter` → `<display-name>`
   - DB names, container names, Dockerfile paths, and more
4. **Creates** `.env` and `frontend/.env` from the committed `.env.example` files
5. **Runs** `go mod tidy`, `npm install`, and `git init && git commit`
6. **Prints** next-step instructions

## Slug rules

The slug must match `^[a-z0-9][a-z0-9-]*[a-z0-9]$`:

- Lowercase letters and digits only
- Hyphens allowed in the middle
- No leading or trailing hyphens
- Minimum 2 characters

Valid: `todo-app`, `my-crm`, `invoicer42`
Invalid: `-app`, `app-`, `My_App`, `a`

## Verification

After running the command, confirm zero template strings remain:

```sh
grep -r "edsonmubezi\|saas-starter\|SaaS Starter\|myapp" todo-app/ \
  --include="*.go" --include="*.ts" --include="*.tsx" \
  --include="*.yaml" --include="*.json" --include="*.mod"
# → zero results
```

And that the project builds:

```sh
cd todo-app && go build ./...
```

## License

MIT
