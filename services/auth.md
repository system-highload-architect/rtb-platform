# Auth Service

## Назначение
Сервис аутентификации и авторизации. Управляет учётными записями пользователей, выдаёт JWT-токены (access + refresh) и проверяет их валидность через gRPC.

## Архитектура
- `cmd/main.go` – инициализация, gRPC-сервер, хранилище пользователей, дефолтный администратор
- `internal/domain/user.go` – структура `User`, интерфейс `UserStore`, in‑memory реализация
- `internal/server/grpc.go` – реализация gRPC-сервера (`Register`, `Login`, `Validate`, `Refresh`)
- `configs/dev.yaml` – порт, секретный ключ JWT, TTL токенов

## Реализованные методы (proto `auth.proto`)

### Register
- Принимает `email`, `password`, `role`
- Проверяет, что пользователь не существует
- Хранит `SHA256` хеш пароля (в production нужно заменить на bcrypt)
- Возвращает `user_id`

### Login
- Проверяет email/пароль
- Генерирует access‑токен (15 мин) и refresh‑токен (72 часа)
- Возвращает `access_token`, `refresh_token`, `expires_in`

### Validate
- Проверяет JWT‑токен (подпись, срок действия)
- Возвращает `valid`, `user_id`, `role`, `expires_at`

### Refresh (упрощён)
- Принимает refresh‑токен, проверяет его валидность
- Выдаёт новую пару токенов (access + refresh)
- **Важно:** сейчас не реализована инвалидация refresh‑токенов (нет чёрного списка)

## Используемые пакеты
- `github.com/golang-jwt/jwt/v5` – работа с JWT
- Стандартный `crypto/sha256` – хеширование паролей
- Общие: `config`, `logger`, `metrics`, `shutdown`

## Хранилище
- In‑memory map (`inmemStore`) – при перезапуске данные теряются
- Создаётся администратор по умолчанию:
  - Email: `admin@rtb-platform.local`
  - Пароль: `Admin123!`

## Конфигурация (`dev.yaml`)
- Порт: `9004`
- JWT secret: `"super-secret-key-change-in-production"`
- access TTL: `15m`, refresh TTL: `72h`

## Что осталось разработать (TODO)
- [ ] Заменить SHA256 на bcrypt для хеширования паролей
- [ ] Полноценная ролевая модель (сейчас роль просто строка)
- [ ] Инвалидация refresh‑токенов (logout / смена пароля)
- [ ] Персистентное хранилище (PostgreSQL)
- [ ] Подключить `auth` к Gateway (сейчас отключён: `authMiddleware = nil`)
- [ ] Защитить маршруты в дашборде (PrivateRoute)
- [ ] Добавить эндпоинты для смены пароля, сброса и т.д.