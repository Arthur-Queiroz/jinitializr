# CLAUDE.md — J Initializr

Contexto do projeto para o Claude Code. Leia isto antes de qualquer tarefa.

## O que é

Web app inspirado no [start.spring.io](https://start.spring.io), porém para Go. O usuário configura um projeto Go via formulário (module path, versão, router, dependências) e baixa um `.zip` com o projeto scaffoldado pronto para rodar.

O gerador produz **apenas o backend Go** — não gera frontend. O J Initializr em si tem um frontend Vue, mas isso é separado do que ele gera (assim como o Spring Initializr é escrito em Java mas gera só o projeto Spring).

## Filosofia central

**Go faz quase tudo com a std lib.** O app usa `net/http`, `text/template`, `archive/zip` e `embed` — zero dependências obrigatórias na v0. Isso não é só preferência: é a mensagem do projeto. O J Initializr existe para demonstrar que Go não precisa de framework para o básico, então o próprio app não deve depender de frameworks. Dogfooding é regra, não sugestão.

Se uma tarefa parecer mais fácil adicionando uma dependência externa, pare e considere a versão stdlib primeiro. Quase sempre o custo da stdlib é pequeno e o ganho de coerência vale mais.

## Princípio do output gerado

O projeto que o J Initializr gera segue o oposto da abundância: **mínimo de coisas no esqueleto base, tudo opt-in.** Espelha o Spring Initializr, onde a tela abre com "No dependency selected".

- Router é escolha obrigatória, default = stdlib (`net/http`). Opções: stdlib / Chi / Gin.
- Dependências (pgx, godotenv, sqlc, Air) são todas opt-in. Nada vem marcado.
- Banco de dados é opcional.
- A liberdade do usuário manda. Não imponha estrutura além do mínimo que roda.

## Arquitetura

```
j-initializr/
├── cmd/server/main.go        # bootstrap: DI manual, sobe HTTP
├── assets.go                 # go:embed do web/dist (raiz: embed não sobe "..")
├── internal/
│   ├── handler/              # HTTP: parse, chama generator, devolve zip/json
│   ├── generator/            # orquestra: config → mapa de arquivos
│   ├── template/             # renderiza .tmpl (text/template) + go:embed
│   │   └── templates/        # arquivos .tmpl (common/ + layouts/{stdlib,chi,gin}/)
│   ├── zipper/               # map[path][]byte → zip em memória
│   ├── model/                # tipos centrais, sem lógica
│   └── catalog/              # registry de routers e dependências
├── web/                      # Vue 3 + Vite (dist/ embarcado via go:embed)
├── go.mod
├── Makefile
└── Dockerfile
```

> Nota: os templates moram em `internal/template/templates/` (não na raiz como o
> rascunho inicial sugeria). O `//go:embed` não consegue referenciar diretórios
> pais (`..`), então a árvore embarcada precisa ficar dentro do pacote que a usa.

### Fluxo de dependências entre pacotes

`cmd/server` → injeta tudo → `handler` → `generator` → (`template`, `zipper`, `catalog`). `model` é folha (não depende de nada). Nunca crie import cíclico; se precisar, o tipo provavelmente pertence a `model`.

### Camadas

| Pacote | Responsabilidade |
|---|---|
| `cmd/server` | Bootstrap, injeção de dependência manual |
| `handler` | HTTP: parse request, chama generator, devolve zip/json |
| `generator` | Orquestra geração: config → mapa de arquivos |
| `template` | Renderiza `.tmpl` com dados do config |
| `zipper` | Monta `.zip` em memória |
| `model` | Tipos centrais (ProjectConfig, Dependency, Router), sem lógica |
| `catalog` | Registry de routers e dependências |

## Convenções de código

- **Injeção de dependência manual.** Sem framework DI. `main.go` constrói tudo e injeta. Cada pacote expõe um `New(...)` que recebe suas dependências.
- **Tipos centrais em `model/`**, só structs e enums, zero lógica de negócio.
- **`//go:embed`** para templates e frontend. Zero file I/O em runtime — tudo embarcado no binário.
- **Roteamento stdlib** com method+path do Go 1.22+: `mux.HandleFunc("POST /api/generate", h.Generate)`.
- **Middleware via composição** de `func(http.Handler) http.Handler`, não framework.
- **Erros**: wrap com `fmt.Errorf("...: %w", err)`. Handlers traduzem erro → status HTTP; pacotes internos retornam erro, nunca escrevem resposta HTTP.
- **Nomes**: pacotes em minúsculo singular (`handler`, não `handlers`). Sem stutter (`catalog.Catalog`, não `catalog.CatalogStruct`).
- Rode `gofmt`/`goimports` sempre. `go vet` deve passar limpo.

## API

### `GET /api/catalog`
Retorna routers e dependências disponíveis. Chamado uma vez no load do frontend. A busca de dependências é **client-side** — o backend não tem endpoint de search, só devolve o catálogo inteiro e o Vue filtra em memória.

### `POST /api/generate`
Recebe `ProjectConfig` no body JSON, retorna `application/zip`.

## Tipos centrais (model/)

```go
type Router string
const (
    RouterStdlib Router = "stdlib"
    RouterChi    Router = "chi"
    RouterGin    Router = "gin"
)

type ProjectConfig struct {
    ModulePath  string
    ProjectName string
    GoVersion   string
    Router      Router
    Deps        []Dependency
}

type Dependency struct {
    ID        string   // "pgx"
    Name      string   // "PostgreSQL (pgx)"
    Category  string   // "database", "config", "observability", "tooling"
    GoModule  string   // "github.com/jackc/pgx/v5" — vazio se for tool/stdlib
    Version   string   // "v5.7.1"
    Files     []string // templates extras que essa dep traz
    Conflicts []string // dependências incompatíveis
}
```

## Catálogo de dependências (v0)
Cada dependência é um **conjunto de mutações** no projeto gerado, não só uma linha no `go.mod`. Pode trazer imports, arquivos de config, targets de Makefile e blocos de código.

| Dep | Tipo | go.mod? | O que injeta |
|---|---|---|---|
| pgx | lib runtime | sim | import + pool em `internal/database/db.go` |
| godotenv | lib runtime | sim | `godotenv.Load()` + `.env.example` |
| slog | stdlib | não | **base** — logger estruturado mínimo em todo projeto |
| sqlc | tool de build | não | `sqlc.yaml` + `db/{schema,query}.sql` + Makefile target |
| Air | tool de dev | não | `.air.toml` + Makefile target `make dev` |

Adicionar uma dep nova = uma entrada no `catalog.go` + seus templates. Não cria rotas, não mexe no frontend.

`slog` é o único item da lista que vai no **esqueleto base** (não é opt-in): logger com handler de texto, configurado no `main.go`. O setup de observabilidade completo (handler JSON, request logging, request ID) fica como dep opt-in futura.

## Estado atual
Fase de design concluída. Decisões fechadas:
- Frontend: Vue 3 estático (Vite), embarcado no binário
- App: stdlib puro, zero deps obrigatórias
- Output: stdlib default, Chi/Gin opt-in
- Catálogo: endpoint único, busca client-side, deps categorizadas
- `slog`: logger estruturado mínimo no esqueleto base (decidido)

### Critério do esqueleto base

Entra no base apenas o que é (a) stdlib, (b) transversal a 100% dos projetos e (c) sem alternativa de opinião razoável na forma mínima. `slog` (logger de boot mínimo) passa nos três. `Makefile` opinativo falha em (c). Healthcheck falha em (b). Use este critério para qualquer proposta futura de adicionar algo ao base.

- i18n: UI bilíngue PT/EN via `vue-i18n` (client-side, `web/src/locales/`). Não afeta o output gerado — projeto gerado é sempre em inglês. Go não participa da i18n.

## Documentos relacionados

- `architecture.md` (Obsidian vault) — documento de arquitetura completo
- `AGENTS.md` — regras operacionais para agentes de código
- `frontend.md` — spec do frontend Vue (estrutura, deps, design tokens)
