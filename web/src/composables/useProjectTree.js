// Renders the GENERATED project's folder tree for the "Explore" modal.
//
// The list of files is NOT decided here — it comes from the backend's
// POST /api/preview, which returns the exact paths the generator would pack.
// This keeps internal/generator the single source of truth; the frontend only
// turns a flat, sorted path list into an ASCII tree for display.
//
// The only project knowledge that lives here is COMMENTS: the human-friendly
// blurb shown next to each file. That's pure presentation (and i18n), so it
// belongs on the client.

const COMMENTS = {
  'cmd/api/main.go': 'tree.main',
  'internal/config/config.go': 'tree.config',
  'internal/http/server.go': 'tree.http',
  'internal/database/db.go': 'tree.database',
  'db/schema.sql': 'tree.dbSchema',
  'db/query.sql': 'tree.dbQuery',
  'sqlc.yaml': 'tree.sqlc',
  '.air.toml': 'tree.air',
  '.env.example': 'tree.env',
  'go.mod': 'tree.gomod',
  Makefile: 'tree.makefile',
  '.gitignore': 'tree.gitignore',
  'README.md': 'tree.readme',
}

// buildProjectTree turns a flat list of file paths (relative to the project
// root) into display lines with ASCII connector prefixes.
export function buildProjectTree(paths, projectName) {
  const rootChildren = []

  const child = (list, name, dir) => {
    let node = list.find((n) => n.name === name && n.dir === dir)
    if (!node) {
      node = { name, dir, children: [], comment: null }
      list.push(node)
    }
    return node
  }

  for (const p of paths) {
    const parts = p.split('/')
    let list = rootChildren
    parts.forEach((part, idx) => {
      const isFile = idx === parts.length - 1
      const node = child(list, part, !isFile)
      if (isFile) node.comment = COMMENTS[p] ?? null
      list = node.children
    })
  }

  sortTree(rootChildren)

  const lines = [{ prefix: '', name: `${projectName || 'my-app'}/`, comment: null, dir: true }]
  let fileCount = 0

  const walk = (nodes, prefix) => {
    nodes.forEach((n, i) => {
      const last = i === nodes.length - 1
      lines.push({
        prefix: prefix + (last ? '└─ ' : '├─ '),
        name: n.dir ? `${n.name}/` : n.name,
        comment: n.comment,
        dir: n.dir,
      })
      if (!n.dir) fileCount += 1
      if (n.dir) walk(n.children, prefix + (last ? '   ' : '│  '))
    })
  }

  walk(rootChildren, '')

  return { lines, fileCount }
}

// sortTree orders each level directories-first, then alphabetically, so the
// tree reads conventionally regardless of the server's path ordering.
function sortTree(nodes) {
  nodes.sort((a, b) => {
    if (a.dir !== b.dir) return a.dir ? -1 : 1
    return a.name.localeCompare(b.name)
  })
  nodes.forEach((n) => n.dir && sortTree(n.children))
}
