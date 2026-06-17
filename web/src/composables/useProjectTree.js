// Builds the folder tree of the GENERATED project (not of the J Initializr app
// itself) from the current dependency selection. The layout below — cmd/api,
// internal/config, internal/http, internal/database — is the output project's
// structure. Do not confuse it with the app's internal/ packages.
//
// IMPORTANT: this preview must mirror what internal/generator actually emits.
// The base file set and the dep→file mapping below are kept in sync with
// internal/catalog (the FileSpec entries) and the templates. If you change what
// the generator produces, update this tree too.
//
// Base (always present): cmd/api/main.go, internal/config/config.go,
// internal/http/server.go, go.mod, Makefile, .gitignore, README.md.
// (No go.sum — it isn't generated; the README tells the user to run
// `go mod tidy`.)
//   pgx       → internal/database/db.go
//   sqlc      → db/schema.sql, db/query.sql, sqlc.yaml
//   godotenv  → .env.example
//   air       → .air.toml

function node(name, { comment = null, dir = false, children = null } = {}) {
  return { name, comment, dir, children }
}

export function buildProjectTree(selectedIds, projectName) {
  const has = (id) => selectedIds.includes(id)

  const internalChildren = [
    node('config/', { dir: true, children: [node('config.go', { comment: 'tree.config' })] }),
    node('http/', { dir: true, children: [node('server.go', { comment: 'tree.http' })] }),
  ]
  if (has('pgx')) {
    internalChildren.push(
      node('database/', { dir: true, children: [node('db.go', { comment: 'tree.database' })] }),
    )
  }

  const root = [
    node('cmd/', {
      dir: true,
      children: [
        node('api/', { dir: true, children: [node('main.go', { comment: 'tree.main' })] }),
      ],
    }),
    node('internal/', { dir: true, children: internalChildren }),
  ]

  if (has('sqlc')) {
    root.push(
      node('db/', {
        dir: true,
        children: [
          node('schema.sql', { comment: 'tree.dbSchema' }),
          node('query.sql', { comment: 'tree.dbQuery' }),
        ],
      }),
    )
  }
  if (has('sqlc')) root.push(node('sqlc.yaml', { comment: 'tree.sqlc' }))
  if (has('air')) root.push(node('.air.toml', { comment: 'tree.air' }))
  if (has('godotenv')) root.push(node('.env.example', { comment: 'tree.env' }))

  root.push(node('.gitignore', { comment: 'tree.gitignore' }))
  root.push(node('go.mod', { comment: 'tree.gomod' }))
  root.push(node('Makefile', { comment: 'tree.makefile' }))
  root.push(node('README.md', { comment: 'tree.readme' }))

  // Flatten into display lines with ASCII connector prefixes.
  const lines = []
  let fileCount = 0

  lines.push({ prefix: '', name: `${projectName || 'my-app'}/`, comment: null, dir: true })

  function walk(nodes, prefix) {
    nodes.forEach((n, i) => {
      const last = i === nodes.length - 1
      lines.push({
        prefix: prefix + (last ? '└─ ' : '├─ '),
        name: n.name,
        comment: n.comment,
        dir: n.dir,
      })
      if (!n.dir) fileCount += 1
      if (n.children) walk(n.children, prefix + (last ? '   ' : '│  '))
    })
  }

  walk(root, '')

  return { lines, fileCount }
}
