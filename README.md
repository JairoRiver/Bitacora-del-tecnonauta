# Bitácora del Tecnonauta

Blog personal sobre tecnología. Monorepo con un backend en Go que expone una API REST y un panel de administración, y un frontend estático en Astro que consume esa API en tiempo de build.

## Estructura

```
bitacora_tech/
├── backend/   # API REST + admin UI (Go)
└── frontend/  # Blog estático (Astro)
```

---

## Backend

### Tecnologías

| Herramienta | Rol |
|---|---|
| Go 1.25 | Lenguaje |
| PostgreSQL 18 | Base de datos |
| [chi](https://github.com/go-chi/chi) | Router HTTP |
| [sqlc](https://sqlc.dev) | Generación de código desde SQL |
| [golang-migrate](https://github.com/golang-migrate/migrate) | Migraciones |
| [templ](https://templ.guide) | Templates HTML del admin |
| [zerolog](https://github.com/rs/zerolog) | Logging estructurado |
| [Viper](https://github.com/spf13/viper) | Configuración |
| Docker / docker compose | Postgres en desarrollo |
| [Task](https://taskfile.dev) | Runner de tareas |

### Requisitos previos

- Go ≥ 1.25
- Docker + docker compose
- [Task](https://taskfile.dev/#/installation)
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html)
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- [templ CLI](https://templ.guide/quick-start/installation) (`go install github.com/a-h/templ/cmd/templ@latest`)
- [air](https://github.com/air-verse/air) (opcional, solo para hot reload)

### Configuración

```bash
cd backend
cp config.yaml.example config.yaml
# Edita config.yaml con tus credenciales
```

Los campos obligatorios son `database.user`, `database.password` y `admin.auth_secret`.

> `config.yaml` está en `.gitignore` y nunca se commitea.

### Puesta en marcha

```bash
cd backend

# 1. Levantar Postgres
task db:up

# 2. Aplicar migraciones
task db:migrate

# 3. Crear usuario admin
task admin:create-user -- --username admin --password tu_contraseña

# 4. Arrancar el servidor (con hot reload)
task dev

# — o compilar y ejecutar directamente —
task build
./bin/server
```

El servidor escucha en `http://localhost:8080` por defecto.

### API pública

Consumida por el frontend de Astro en tiempo de build.

| Método | Ruta | Descripción |
|---|---|---|
| GET | `/health` | Estado del servidor |
| GET | `/api/posts` | Lista de posts |
| GET | `/api/posts/{slug}` | Detalle de un post |
| GET | `/api/categories` | Lista de categorías |
| GET | `/api/categories/{slug}` | Posts de una categoría |

### Panel de administración

Disponible en `http://localhost:8080/admin/`.

| Ruta | Descripción |
|---|---|
| `/admin/login` | Formulario de acceso |
| `/admin/` | Dashboard — lista de posts |
| `/admin/posts/new` | Crear post |
| `/admin/posts/{id}/edit` | Editar post |

El editor de contenido soporta cuatro tipos de bloque: **párrafo**, **código**, **imagen** y **tabla**.

### Tareas disponibles

```
task db:up              # Levanta el contenedor de Postgres
task db:down            # Para y elimina el contenedor
task db:migrate         # Aplica migraciones pendientes
task db:migrate:down    # Revierte la última migración
task db:migrate:status  # Versión actual de la DB
task db:seed            # Importa posts.json a la DB (migración inicial)
task admin:create-user  # Crea el usuario admin
task generate           # Regenera código sqlc + templ
task build              # Compila el binario
task dev                # Arranca con hot reload (air)
task test               # Ejecuta todos los tests
task test:coverage      # Tests con informe de cobertura
task tidy               # go mod tidy
```

### Tests

Los tests de integración usan [testcontainers-go](https://golang.testcontainers.org/) — levantan Postgres en Docker automáticamente, no necesitan configuración extra.

```bash
task test
```

---

## Frontend

### Tecnologías

| Herramienta | Rol |
|---|---|
| [Astro](https://astro.build) | Framework SSG |
| [Tailwind CSS](https://tailwindcss.com) | Estilos |
| TypeScript | Tipado |

### Requisitos previos

- Node.js ≥ 18
- npm

### Configuración

```bash
cd frontend
cp .env.example .env
# Edita .env con la URL del backend
```

```env
API_URL=http://localhost:8080
```

### Puesta en marcha

```bash
cd frontend
npm install

# Desarrollo local
npm run dev

# Build estático (requiere el backend en marcha)
npm run build

# Previsualizar el build
npm run preview
```

El frontend genera un sitio completamente estático en `dist/`. En producción se sube ese directorio a cualquier hosting estático (Netlify, Vercel, S3, etc.).

> El backend debe estar accesible desde la máquina donde se ejecuta `npm run build`, ya que Astro hace fetch a la API en tiempo de compilación.

---

## Flujo de trabajo habitual

```
# Terminal 1 — backend
cd backend && task dev

# Terminal 2 — frontend
cd frontend && npm run dev
```

Para publicar contenido nuevo:

1. Accede al admin en `http://localhost:8080/admin/`
2. Crea o edita el post
3. Ejecuta `npm run build` en el frontend
4. Despliega el directorio `frontend/dist/`
