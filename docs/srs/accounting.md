# 🇬🇧 Accounting SRS / 🇷🇺 Техническое задание на Accounting

## 🇬🇧 Overview / 🇷🇺 Обзор

Accounting is the financial service responsible for storing campaign balances and processing debit operations. It guarantees idempotency so that the same auction never results in a double charge. This document defines the functional and non‑functional requirements implemented in Accounting.
Accounting — финансовый сервис, отвечающий за хранение балансов кампаний и обработку операций списания. Он гарантирует идемпотентность, чтобы один и тот же аукцион никогда не приводил к двойному списанию. Этот документ определяет реализованные функциональные и нефункциональные требования к Accounting.

---

## 🇬🇧 Functional Requirements / 🇷🇺 Функциональные требования

### FR‑ACC‑001: Balance Storage
**🇬🇧** Accounting MUST store the current balance for each campaign. The balance MUST be represented in minimal currency units (kopeks) using the `fixedpoint` package to avoid floating‑point errors.  
**🇷🇺** Accounting ДОЛЖЕН хранить текущий баланс каждой кампании. Баланс ДОЛЖЕН быть представлен в минимальных денежных единицах (копейках) с использованием пакета `fixedpoint`, чтобы избежать ошибок плавающей точки.

### FR‑ACC‑002: Debit Operation
**🇬🇧** Accounting MUST provide a `Debit` gRPC method that atomically deducts the requested amount from a campaign’s balance. It MUST check that the balance is sufficient before deducting. If the balance is insufficient, it MUST return an error.  
**🇷🇺** Accounting ДОЛЖЕН предоставлять gRPC‑метод `Debit`, который атомарно списывает запрошенную сумму с баланса кампании. Он ДОЛЖЕН проверять, что баланс достаточен перед списанием. Если баланс недостаточен, он ДОЛЖЕН возвращать ошибку.

### FR‑ACC‑003: Idempotency
**🇬🇧** Debit operations MUST be idempotent based on the `bid_id` field. If a `Debit` request with an already processed `bid_id` is received, Accounting MUST return success without modifying the balance.  
**🇷🇺** Операции списания ДОЛЖНЫ быть идемпотентными на основе поля `bid_id`. Если получен запрос `Debit` с уже обработанным `bid_id`, Accounting ДОЛЖЕН вернуть успех без изменения баланса.

### FR‑ACC‑004: Get Balance
**🇬🇧** Accounting MUST provide a `GetBalance` gRPC method that returns the current balance for a given campaign. If the campaign does not exist, it MUST return an error.  
**🇷🇺** Accounting ДОЛЖЕН предоставлять gRPC‑метод `GetBalance`, возвращающий текущий баланс для заданной кампании. Если кампания не существует, он ДОЛЖЕН возвращать ошибку.

### FR‑ACC‑005: Persistent Storage (PostgreSQL)
**🇬🇧** When configured with a database DSN, Accounting MUST use PostgreSQL to store balances. The `balances` table and the `debit_balance` stored function MUST be created automatically at startup if they do not already exist.  
**🇷🇺** При настройке DSN базы данных Accounting ДОЛЖЕН использовать PostgreSQL для хранения балансов. Таблица `balances` и хранимая функция `debit_balance` ДОЛЖНЫ создаваться автоматически при запуске, если они ещё не существуют.

### FR‑ACC‑006: In‑Memory Fallback
**🇬🇧** If no database DSN is provided, or the database connection fails, Accounting MUST fall back to an in‑memory store. This allows development and testing without PostgreSQL.  
**🇷🇺** Если DSN базы данных не указан или подключение не удалось, Accounting ДОЛЖЕН переключиться на in‑memory хранилище. Это позволяет вести разработку и тестирование без PostgreSQL.

### FR‑ACC‑007: Balance Initialisation
**🇬🇧** Accounting MUST initialise balances for predefined test campaigns at startup (`campaign-1`, `1001`, `1002`).  
**🇷🇺** Accounting ДОЛЖЕН инициализировать балансы для предопределённых тестовых кампаний при запуске (`campaign-1`, `1001`, `1002`).

---

## 🇬🇧 Non‑Functional Requirements / 🇷🇺 Нефункциональные требования

- **NFR‑ACC‑001**: Debit operations MUST be atomic – a race condition between concurrent debits MUST NOT result in an incorrect balance.  
  Операции списания ДОЛЖНЫ быть атомарными – гонка между параллельными списаниями НЕ ДОЛЖНА приводить к некорректному балансу.
- **NFR‑ACC‑002**: The stored function `debit_balance` MUST minimise network round‑trips by performing the check and update in a single call.  
  Хранимая функция `debit_balance` ДОЛЖНА минимизировать сетевые обмены, выполняя проверку и обновление в одном вызове.
- **NFR‑ACC‑003**: Idempotency keys MUST have a configurable TTL and MUST be evicted automatically after expiration.  
  Ключи идемпотентности ДОЛЖНЫ иметь настраиваемый TTL и ДОЛЖНЫ автоматически удаляться после истечения.
- **NFR‑ACC‑004**: Accounting MUST support graceful shutdown with a maximum total timeout of 30 seconds.  
  Accounting ДОЛЖЕН поддерживать корректное завершение работы с общим таймаутом 30 секунд.

---

## 🇬🇧 Data Flow / 🇷🇺 Поток данных

### Debit (with idempotency)

```mermaid
sequenceDiagram
    participant Client (Auction)
    participant Accounting
    participant PostgreSQL

    Client->>Accounting: Debit(campaign_id, amount, bid_id)
    Accounting->>Accounting: Check idempotent(bid_id)
    alt Already processed
        Accounting-->>Client: success (no double spend)
    else New request
        Accounting->>PostgreSQL: debit_balance(campaign_id, amount)
        PostgreSQL-->>Accounting: success + remaining balance / error
        Accounting-->>Client: DebitResponse
    end