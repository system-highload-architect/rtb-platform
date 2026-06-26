# 🇬🇧 Auth SRS / 🇷🇺 Техническое задание на Auth

## 🇬🇧 Overview / 🇷🇺 Обзор

Auth is the authentication and authorisation service. It manages user accounts, issues JWT access and refresh tokens, and validates tokens on behalf of the Gateway. This document defines the functional and non‑functional requirements implemented in Auth.
Auth — сервис аутентификации и авторизации. Он управляет учётными записями пользователей, выпускает JWT access‑ и refresh‑токены и проверяет токены по запросу Gateway. Этот документ определяет реализованные функциональные и нефункциональные требования к Auth.

---

## 🇬🇧 Functional Requirements / 🇷🇺 Функциональные требования

### FR‑AUTH‑001: User Registration
**🇬🇧** Auth MUST provide a `Register` gRPC method that creates a new user account. The request MUST include `email`, `password`, and `role`. The password MUST be hashed before storage. If the email already exists, an error MUST be returned.  
**🇷🇺** Auth ДОЛЖЕН предоставлять gRPC‑метод `Register`, создающий новую учётную запись. Запрос ДОЛЖЕН содержать `email`, `password` и `role`. Пароль ДОЛЖЕН быть хеширован перед сохранением. Если email уже существует, ДОЛЖНА возвращаться ошибка.

### FR‑AUTH‑002: User Login
**🇬🇧** Auth MUST provide a `Login` gRPC method that verifies credentials and returns a pair of tokens: a short‑lived access token and a long‑lived refresh token. If credentials are invalid, an error MUST be returned.  
**🇷🇺** Auth ДОЛЖЕН предоставлять gRPC‑метод `Login`, который проверяет учётные данные и возвращает пару токенов: короткоживущий access‑токен и долгоживущий refresh‑токен. Если учётные данные неверны, ДОЛЖНА возвращаться ошибка.

### FR‑AUTH‑003: JWT Access Token
**🇬🇧** The access token MUST be a signed JWT (HS256) containing the user ID, email, and role. It MUST have a configurable expiration time (default 15 minutes).  
**🇷🇺** Access‑токен ДОЛЖЕН быть подписанным JWT (HS256), содержащим ID пользователя, email и роль. Он ДОЛЖЕН иметь настраиваемое время истечения (по умолчанию 15 минут).

### FR‑AUTH‑004: JWT Refresh Token
**🇬🇧** The refresh token MUST be a signed JWT with a longer expiration time (default 72 hours). It MUST be accepted by the `Refresh` method to issue a new pair of tokens.  
**🇷🇺** Refresh‑токен ДОЛЖЕН быть подписанным JWT с более длительным временем истечения (по умолчанию 72 часа). Он ДОЛЖЕН приниматься методом `Refresh` для выдачи новой пары токенов.

### FR‑AUTH‑005: Token Validation
**🇬🇧** Auth MUST provide a `Validate` gRPC method that checks the signature and expiration of an access token. It MUST return the token’s validity, user ID, role, and expiration time.  
**🇷🇺** Auth ДОЛЖЕН предоставлять gRPC‑метод `Validate`, который проверяет подпись и срок действия access‑токена. Он ДОЛЖЕН возвращать действительность токена, ID пользователя, роль и время истечения.

### FR‑AUTH‑006: Token Refresh
**🇬🇧** Auth MUST provide a `Refresh` gRPC method that accepts a refresh token, validates it, and returns a new access token and refresh token pair. Invalid or expired refresh tokens MUST return an error.  
**🇷🇺** Auth ДОЛЖЕН предоставлять gRPC‑метод `Refresh`, который принимает refresh‑токен, проверяет его и возвращает новую пару access‑ и refresh‑токенов. Недействительные или истёкшие refresh‑токены ДОЛЖНЫ возвращать ошибку.

### FR‑AUTH‑007: Default Administrator
**🇬🇧** Auth MUST create a default administrator account at startup if the user store is empty. The credentials MUST be configurable or use well‑known defaults for development.  
**🇷🇺** Auth ДОЛЖЕН создавать учётную запись администратора по умолчанию при запуске, если хранилище пользователей пусто. Учётные данные ДОЛЖНЫ быть настраиваемыми или использовать общеизвестные значения для разработки.

### FR‑AUTH‑008: Password Hashing
**🇬🇧** Passwords MUST be hashed using SHA‑256 in the current implementation. A production‑ready deployment MUST replace this with bcrypt or argon2.  
**🇷🇺** Пароли ДОЛЖНЫ хешироваться с использованием SHA‑256 в текущей реализации. Для production ДОЛЖЕН использоваться bcrypt или argon2.

---

## 🇬🇧 Non‑Functional Requirements / 🇷🇺 Нефункциональные требования

- **NFR‑AUTH‑001**: JWT secret MUST be configurable via environment variable and MUST NOT be hard‑coded in production.  
  Секрет JWT ДОЛЖЕН настраиваться через переменную окружения и НЕ ДОЛЖЕН быть жёстко задан в production.
- **NFR‑AUTH‑002**: Token expiration times MUST be configurable independently for access and refresh tokens.  
  Время истечения токенов ДОЛЖНО настраиваться независимо для access‑ и refresh‑токенов.
- **NFR‑AUTH‑003**: The service MUST support graceful shutdown with a maximum total timeout of 30 seconds.  
  Сервис ДОЛЖЕН поддерживать корректное завершение работы с общим таймаутом 30 секунд.

---

## 🇬🇧 Data Flow / 🇷🇺 Поток данных

### Login

```mermaid
sequenceDiagram
    participant Gateway
    participant Auth
    participant Store[In‑Memory Store]

    Gateway->>Auth: Login(email, password)
    Auth->>Store: Get user by email
    Store-->>Auth: user
    Auth->>Auth: Verify password hash
    Auth->>Auth: Generate access + refresh token
    Auth-->>Gateway: LoginResponse{access_token, refresh_token}