# OSPAB Panel 0.1-alpha

Панель управления хостингом с открытым исходным кодом. Включает в себя API и веб-интерфейс для управления виртуальными серверами.

## Особенности

- 🔐 JWT авторизация
- 🗄️ MySQL база данных  
- 🌐 REST API
- 📱 Веб-интерфейс
- 🔧 Конфигурация через .env файл

## Быстрый старт

### 1. Установка зависимостей

```bash
go mod tidy
```

### 2. Настройка базы данных

#### Вариант 1: Автоматическая настройка
Выполните SQL команды из файла `setup_mysql.sql`:

```bash
mysql -u root -p < setup_mysql.sql
```

#### Вариант 2: Ручная настройка
1. Откройте MySQL консоль:
```bash
mysql -u root -p
```

2. Выполните команды:
```sql
CREATE DATABASE IF NOT EXISTS ospab_panel CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'ospab'@'localhost' IDENTIFIED BY 'ospab123';
GRANT ALL PRIVILEGES ON ospab_panel.* TO 'ospab'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

### 3. Настройка переменных окружения

Файл `.env` уже создан с базовыми настройками. Измените параметры подключения к БД:

```env
# Server Configuration
API_PORT=5000
WEB_PORT=8080

# JWT Configuration
JWT_SECRET=ospab-panel-secret-key-change-in-production

# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_NAME=ospab_panel
DB_USER=root
DB_PASSWORD=password
```

### (Опционально) Prisma для миграций

Если хотите управлять схемой БД через Prisma:
1. Добавьте переменную `DATABASE_URL`, составив её из параметров выше, например:
```
DATABASE_URL="mysql://root:password@localhost:3306/ospab_panel?charset=utf8mb4&parseTime=True&loc=Local"
```
2. Установите prisma (в каталоге `web` или корне, если используете общий package.json):
```
npm install --save-dev prisma
npx prisma migrate dev --name init
npx prisma studio   # просмотр данных
```

Схема расположена в `prisma/schema.prisma`. Go-код по‑прежнему использует `database/sql`; Prisma здесь может применяться только для миграций / инспекции. Полноценная работа через Prisma потребует отдельного Node sidecar сервиса либо перехода на Go ORM (sqlc/ent/gorm).

### 4. Запуск

```bash
go run cmd/server/main.go
```

Серверы будут доступны по адресам:
- **API**: http://localhost:5000
- **Веб-интерфейс**: http://localhost:8080

## API Endpoints

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