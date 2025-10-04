# Ospab Panel

Панель управления виртуальными машинами и гипервизорами (Proxmox, VMware, Hyper-V, KVM, Xen).

## Стек
- **Backend:** Go, Gorilla mux, MySQL, JWT, bcrypt, AES-GCM
- **Frontend:** React, Vite, Tailwind CSS
- **ORM:** Prisma (только миграции)

## Быстрый старт
1. Установите Go >= 1.20 и Node.js >= 18
2. Скопируйте `.env.example` → `.env` и укажите:
  - `DATABASE_URL="user:pass@tcp(localhost:3306)/dbname?parseTime=true"`
  - `SERVER_SECRET_KEY="ваш_32_символьный_ключ"`
3. Примените миграции:
  ```bash
  npx prisma migrate deploy
  ```
4. Установите зависимости и запустите фронтенд:
  ```bash
  cd web
  npm install
  npm run dev
  ```
5. Запустите бэкенд:
  ```bash
  go run ./cmd/server/main.go
  ```
6. Откройте [http://localhost:5173](http://localhost:5173)

# Ospab Panel

Панель управления виртуальными машинами и гипервизорами (Proxmox, VMware, Hyper-V, KVM, Xen).

## Стек
- **Backend:** Go, Gorilla mux, MySQL, JWT, bcrypt, AES-GCM
- **Frontend:** React, Vite, Tailwind CSS
- **ORM:** Prisma (только миграции)

## Быстрый старт
1. Установите Go >= 1.20 и Node.js >= 18
2. Скопируйте `.env.example` → `.env` и укажите:
  - `DATABASE_URL="user:pass@tcp(localhost:3306)/dbname?parseTime=true"`
  - `SERVER_SECRET_KEY="ваш_32_символьный_ключ"`
3. Примените миграции:
  ```bash
  npx prisma migrate deploy
  ```
4. Установите зависимости и запустите фронтенд:
  ```bash
  cd web
  npm install
  npm run dev
  ```
5. Запустите бэкенд:
  ```bash
  go run ./cmd/server/main.go
  ```
6. Откройте [http://localhost:5173](http://localhost:5173)

## Основные возможности
- Регистрация и вход (JWT)
- Добавление/редактирование серверов гипервизоров
- Список виртуальных машин/контейнеров
- Проверка подключения к гипервизору
- Шифрование паролей серверов (AES-GCM)

## API
- `POST /api/auth/login` — вход
- `POST /api/auth/register` — регистрация
- `GET/POST/PUT/DELETE /api/servers` — управление серверами
- `GET /api/servers/{id}/instances` — список VM/LXC
- `GET /api/hypervisors` — поддерживаемые типы
- `POST /api/hypervisors/check` — тест подключения
- `GET/PATCH /api/servers/{id}/connection` — параметры подключения
- `POST /api/servers/{id}/connection/check` — тест сохранённого подключения

## Пример .env
```
DATABASE_URL="user:pass@tcp(localhost:3306)/dbname?parseTime=true"
SERVER_SECRET_KEY="ваш_32_символьный_ключ"
PRISMA_MANAGED=1
```

## Структура
- `cmd/server/main.go` — запуск Go API
- `internal/` — сервисы, обработчики, бизнес-логика
- `web/` — фронтенд (React)
- `prisma/` — схема и миграции

## Безопасность
- Пароли пользователей — bcrypt + соль
- Пароли серверов — AES-GCM

## Лицензия
MIT
### Авторизация

- `POST /api/auth/login` - Вход в систему
- `POST /api/auth/register` - Регистрация пользователя

**Пример запроса:**
```json
{
  "username": "admin",
  "password": "admin"
}
```

**Ответ:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@ospab.panel"
  }
}
```

### Защищенные эндпоинты

Все запросы должны содержать заголовок: `Authorization: Bearer <token>`

- `GET /api/status` - Статус системы
- `GET /api/version` - Информация о версии

**Примеры:**

```bash
# Получение токена
curl -X POST http://localhost:5000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}'

# Использование токена
curl -X GET http://localhost:5000/api/status \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## Тестовые данные

### База данных
- **Пользователь БД**: ospab
- **Пароль БД**: ospab123
- **База данных**: ospab_panel

### Веб-интерфейс
Система автоматически создает тестового пользователя:
- **Логин**: admin
- **Пароль**: admin
- **Email**: admin@ospab.panel

## Структура проекта

```
.
├── cmd/server/          # Точка входа приложения
├── internal/
│   ├── api/            # HTTP обработчики и роутинг
│   ├── core/           # Бизнес-логика
│   │   └── user/       # Пользователи
│   └── infra/          # Инфраструктура
│       └── db/         # База данных
├── pkg/
│   └── auth/           # JWT авторизация
└── web/                 # React (Vite) фронтенд
  ├── src/             # Исходный код React
  ├── index.html       # Точка входа Vite
  └── dist/            # Собранные артефакты (при билде)
```

## Разработка

### Добавление новых эндпоинтов

1. Добавьте обработчик в `internal/api/handler.go`
2. Зарегистрируйте маршрут в `internal/api/router.go`
3. При необходимости добавьте middleware для авторизации

### База данных

Схема БД автоматически создается при запуске. Таблицы:

- `users` - Пользователи системы

## TODO

- [ ] Управление VPS
- [ ] Интеграция с гипервизорами (KVM, Docker)
- [ ] Расширенная система ролей
- [ ] Мониторинг ресурсов
- [ ] Логирование действий

## Лицензия

Open Source проект для грантов