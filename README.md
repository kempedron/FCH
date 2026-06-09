<p align="center">
  <img src="https://capsule-render.vercel.app/api?type=waving&color=0:6366F1,100:4F46E5&height=200&section=header&text=FCH&fontSize=80&fontColor=ffffff&animation=fadeIn" alt="FCH" />
</p>

<h3 align="center">
  Fast Chat Hub — Real-Time Messenger
</h3>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" alt="Go 1.26"/>
  <img src="https://img.shields.io/badge/WebSocket-010101?logo=socket.io" alt="WebSocket"/>
  <img src="https://img.shields.io/badge/PostgreSQL-15-4169E1?logo=postgresql&logoColor=white" alt="PostgreSQL"/>
  <img src="https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white" alt="Docker Compose"/>
  <img src="https://img.shields.io/badge/Swagger-85EA2D?logo=swagger&logoColor=black" alt="Swagger"/>
  <img src="https://img.shields.io/badge/JWT-auth-000000?logo=jsonwebtokens" alt="JWT Auth"/>
  <img src="https://img.shields.io/badge/license-MIT-yellow" alt="License"/>
</p>

<p align="center">
  <b>FCH</b> — современный мессенджер с real-time чатами, групповыми беседами<br/>
  и WebSocket-взаимодействием, построенный на микросервисной архитектуре Go.
</p>

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Quick Start](#quick-start)
- [API Documentation](#api-documentation)
- [WebSocket](#websocket)
- [Project Structure](#project-structure)
- [Screenshots](#screenshots)
- [License](#license)

---

## Overview

**FCH (Fast Chat Hub)** — это real-time мессенджер, разработанный на Go с использованием микросервисной архитектуры. Проект поддерживает личные и групповые чаты, WebSocket для мгновенного обмена сообщениями, JWT-аутентификацию, CSRF-защиту и инвайт-коды для групповых чатов.

Фронтенд серверный (Go templates) с тёмной и светлой темой. API задокументирован через Swagger/OpenAPI.

---

## Architecture

```
                         ┌──────────────┐
                         │  API Gateway  │  (Go — :8080)
                         └──────┬───────┘
                                │
            ┌───────────────────┼───────────────────┐
            ▼                   ▼                   ▼
    ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
    │  Web Service   │   │ User Service  │   │ Chat Service  │
    │  (Go) :8081    │   │  (Go) :8082   │   │  (Go) :8083   │
    └───────────────┘   └───────────────┘   └───────┬───────┘
                                                     │
                                            ┌────────┴────────┐
                                            │   PostgreSQL    │
                                            │     :5432       │
                                            └─────────────────┘
```

**WebSocket Flow:**
```
Client ──WS──▶ API Gateway ──proxy──▶ Chat Service ──▶ Hub (sync.Map)
                                                        │
                                           ┌────────────┴────────────┐
                                           ▼                        ▼
                                      BroadcastToChat          SendTo
                                      (всем, кроме автора)  (direct message)
```

### Services

| Service | Port | Description |
|---------|------|-------------|
| `api-gateway` | 8080 | Reverse proxy, JWT + CSRF middleware, Swagger UI |
| `web-service` | 8081 | Главная страница с поиском пользователей |
| `user-service` | 8082 | Регистрация, вход, поиск пользователей |
| `chat-service` | 8083 | Чаты, WebSocket, группы, сообщения |
| `db` (PostgreSQL) | 5433 | Основная база данных |

---

## Features

- **💬 Real-time чаты** — WebSocket с auto-reconnect
- **👥 Групповые чаты** — создание, приглашение по invite-коду
- **🔍 Поиск пользователей** — находите собеседников по нику
- **🔐 JWT + CSRF** — безопасная аутентификация и защита форм
- **🌗 Dark/Light тема** — переключаемая цветовая схема
- **📄 Swagger API** — интерактивная документация на `/swagger/`

---

## Tech Stack

- **Language:** Go 1.26
- **HTTP Router:** [gorilla/mux](https://github.com/gorilla/mux)
- **WebSocket:** [gorilla/websocket](https://github.com/gorilla/websocket)
- **Database:** PostgreSQL 15 + [GORM](https://gorm.io/)
- **Auth:** JWT (HS256) + bcrypt + [gorilla/csrf](https://github.com/gorilla/csrf)
- **Docs:** [swaggo/swag](https://github.com/swaggo/swag)
- **Deployment:** Docker Compose (multi-stage builds)

---

## Quick Start

```bash
# 1. Клонирование
git clone https://github.com/kempedron/FCH.git
cd FCH

# 2. Настройка окружения
cp .env.example .env
# Отредактируйте .env под свои нужды

# 3. Запуск
chmod +x start.sh
./start.sh
```

После запуска:
- **Main page:** [http://localhost:8080](http://localhost:8080)
- **Swagger UI:** [http://localhost:8080/swagger/](http://localhost:8080/swagger/)

---

## API Documentation

Swagger-документация доступна в формате OpenAPI:

- `docs/api-gateway/swagger.json`
- `docs/api-gateway/swagger.yaml`

Веб-интерфейс: `GET /swagger/` на API Gateway.

### Key Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/` | — | Главная страница |
| `GET/POST` | `/register` | — | Регистрация |
| `GET/POST` | `/login` | — | Вход |
| `GET` | `/search` | ✓ | Поиск пользователей |
| `GET` | `/my-chats` | ✓ | Список чатов |
| `GET` | `/chat/{id}` | ✓ | Страница чата |
| `POST` | `/chat/new/{userId}` | ✓ | Начать личный чат |
| `GET/POST` | `/chat/create/group` | ✓ | Создать группу |
| `POST` | `/chat/{id}/join/{code}` | ✓ | Присоединиться по коду |

---

## WebSocket

Подключение к WebSocket для real-time сообщений:

```javascript
const ws = new WebSocket(`ws://localhost:8080/ws/chat/${chatId}`);

ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    // { type: "message", chat_id, author_id, content, created_at }
};

ws.send(JSON.stringify({
    chat_id: 1,
    content: "Привет!"
}));
```

**WebSocket Hub** ([`internal/chat-service/handler/websocket.go`](internal/chat-service/handler/websocket.go)) использует `sync.Map` для:
- `Register` — подключение клиента
- `Unregister` — отключение клиента
- `SendTo` — отправка конкретному пользователю
- `BroadcastToChat` — рассылка всем участникам чата (кроме отправителя)

---

## Project Structure

```
FCH/
├── cmd/
│   ├── api-gateway/       # Входная точка Gateway
│   ├── chat-service/      # Чат-сервис (WebSocket + HTTP)
│   ├── user-service/      # Управление пользователями
│   └── web-service/       # Frontend сервис
├── internal/
│   ├── chat-service/      # Бизнес-логика чатов
│   │   ├── handler/       # HTTP + WebSocket хендлеры
│   │   └── service/       # Логика создания/управления чатами
│   ├── database/          # GORM инициализация
│   ├── jwt/               # JWT токены
│   ├── middleware/        # Middleware (JWT, helpers)
│   ├── models/            # GORM модели
│   ├── repository/        # Репозитории
│   ├── user-service/      # Бизнес-логика пользователей
│   └── web-service/       # Логика главной страницы
├── web/templates/         # HTML шаблоны
├── docs/api-gateway/      # Swagger спецификация
├── docker-compose.yml     # Оркестрация
└── start.sh               # Скрипт запуска
```

---


## License

Распространяется под лицензией **MIT**. См. файл [LICENSE](LICENSE) для получения дополнительной информации.

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/kempedron">kempedron</a>
</p>

<p align="center">
  <img src="https://capsule-render.vercel.app/api?type=waving&color=0:4F46E5,100:6366F1&height=120&section=footer" />
</p>
