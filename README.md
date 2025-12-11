# JKH Inspection System

Система осмотра жилых помещений ЖКХ - Full Stack приложение.

## Архитектура

- **Backend**: Go + Gin (порт 8080)
- **Frontend**: React + TypeScript + Vite (порт 3000)

## Почему разные порты?

Это стандартная практика в разработке:
- **Backend (8080)**: REST API сервер, обрабатывает запросы от frontend
- **Frontend (3000)**: Dev сервер Vite, проксирует API запросы на backend

В production они могут быть на одном домене через nginx, но в разработке удобнее раздельно.

## Быстрый старт

### 1. Установка зависимостей

```bash
# Установить зависимости frontend
npm run install:frontend

# Или вручную:
cd frontend && npm install
```

### 2. Запуск всего проекта одной командой

```bash
npm run dev
```

Эта команда запустит одновременно:
- Backend на `http://localhost:8080`
- Frontend на `http://localhost:3000`

### 3. Запуск по отдельности (если нужно)

**Backend:**
```bash
go run main.go
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```
## API

Backend API доступен на `http://localhost:8080/api/v1`
Frontend автоматически проксирует запросы через Vite proxy.

## Разработка

- Backend: `http://localhost:8080`
- Frontend: `http://localhost:3000`
- Swagger: `http://localhost:8080/swagger/index.html`

