---
name: gosourcemapper-guide
description: Guidelines for Go (Gin) and React (Vite) development in the gosourcemapper project.
---

# AI Assistant Guide: Go (Gin) & React Project

This document serves as a guideline for GitHub Copilot, Antigravity, and other AI coding assistants to ensure consistency, maintainability, and adherence to best practices within this codebase.

## 1. Tech Stack Overview

- **Backend**: Go (Golang) with [Gin Web Framework](https://github.com/gin-gonic/gin).
- **Frontend**: React (Vite).
- **Documentation**: Markdown/Starlight (if applicable).

## 2. Backend (Go + Gin) Best Practices

### Project Structure (Standard Go Layout)

We follow the standard project layout:

```
├── cmd/
│   └── server/
│       └── main.go       # Entry point. Initializes the app.
├── internal/             # Private application code (not importable by other projects).
│   ├── config/           # Configuration loading (e.g., env vars).
│   ├── handlers/         # HTTP handlers (controllers).
│   ├── models/           # Domain models and data structures.
│   ├── repository/       # Database access layer.
│   ├── middleware/       # Gin middleware (logging, auth, CORS).
│   ├── router/           # Route definitions.
│   └── service/          # Business logic.
├── pkg/                  # Library code ok to use by external applications (if any).
├── docs/                 # Swagger/OpenAPI docs.
└── go.mod
```

### coding Guidelines

1.  **Dependency Injection**: Avoid global state. Inject services into handlers, and repositories into services.
    - _Example_: Define a `Handler` struct that holds references to `Service` interfaces.
2.  **Error Handling**:
    - Use custom error types to distinguish between client errors (4xx) and server errors (5xx).
    - Return errors up the stack; handle them centrally in the HTTP handler or a middleware.
    - Check _all_ errors.
3.  **Gin Specifics**:
    - Use `gin.Context` only in the Handler layer. Do not pass it down to services.
    - specific route groups (e.g., `v1 := router.Group("/v1")`).
    - Use Struct Tags for binding and validation (e.g., `binding:"required"`).
4.  **Configuration**: Use a strong configuration pattern (e.g., loading from environment variables into a struct).
5.  **Concurrency**: Use Goroutines and Channels responsibly. Always handle context cancellation (`ctx.Done()`).

### Naming Conventions

- **Interfaces**: simple descriptors (e.g., `Service`, `Repository`) or `er` suffix (e.g., `Reader`, `Writer`).
- **files**: snake_case.go
- **Variables**: camelCase.
- **Exported**: PascalCase.

## 3. Frontend (React + Vite) Best Practices

1.  **Component Structure**: Functional components with Hooks.
2.  **State Management**: Use `useState` / `useReducer` for local state; Context API or external libraries (Zustand/Redux) for global state if complex.
3.  **Styling**: Modular CSS or Tailwind (if configured). Avoid inline styles.
4.  **Performance**: Use `useMemo` and `useCallback` judiciously to prevent unnecessary re-renders.

## 4. Documentation

- Keep `docs/` and `documentation/` updated.
- Inline comments should explain _why_, not _what_.
- Exported functions must have Godoc comments.
