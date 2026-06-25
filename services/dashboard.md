# Dashboard (Frontend)

## Назначение
Веб-интерфейс для взаимодействия с RTB-платформой. Предоставляет страницы входа, регистрации, дашборд с аналитикой и демонстрационный аукцион.

## Стек
- **React 18** + **TypeScript**
- **Vite** (сборка, dev‑сервер с прокси)
- **Tailwind CSS** (стилизация)
- **Recharts** (график прогноза)
- **React Router v6** (маршрутизация)

## Структура
- `main.tsx` – точка входа, BrowserRouter с future‑флагами
- `App.tsx` – маршруты
- `api/gateway.ts` – функции для HTTP‑запросов к Gateway (REST и JSON‑RPC)
- `pages/Login.tsx` – страница входа (email/пароль)
- `pages/Register.tsx` – страница регистрации
- `pages/Dashboard.tsx` – главная страница с отчётом, прогнозом, факторным анализом, экспортом Excel
- `pages/Auction.tsx` – демо‑форма аукциона

## API‑клиент (`api/gateway.ts`)

### JSON‑RPC (через `/rpc`)
- `login(email, password)` → сохраняет токен в `localStorage`
- `register(email, password, role)` → перенаправляет на логин

### Защищённые REST‑запросы (через `authFetch`)
- `fetchReport(start, end)` → таблица отчёта
- `fetchForecast(history, horizon)` → график прогноза
- `fetchFactorAnalysis()` → факторный анализ
- `getExportUrl(start, end)` → ссылка на скачивание Excel

Токен передаётся в заголовке `Authorization: Bearer <token>`.

## Страницы и их состояние

### Login
- Поля: Email, Password
- Кнопка «Sign In»
- Ошибка при неверных данных
- Ссылка «Register» → `/register`
- После успеха – редирект на `/`

### Register
- Поля: Email, Password
- Кнопка «Register»
- Ошибка при неудаче
- Ссылка «Sign in» → `/login`
- После успеха – редирект на `/login`

### Dashboard
- Спиннеры при загрузке данных
- Таблица отчёта с агрегированными метриками (показов, кликов, spend) по кампаниям
- Линейный график прогноза (Recharts) на основе тестовых данных
- Факторный анализ (explained variance ratio) – статические данные
- Кнопка «Download Excel Report» (открывает файл)

### Auction
- Форма с полями Device ID, IP, Latitude, Longitude
- Кнопка «Send Bid» отправляет JSON‑RPC запрос `auction.bid`
- Ответ отображается в отформатированном JSON

## Защита маршрутов
- Пока **отсутствует** (любой может открыть Dashboard без токена)
- AuthMiddleware в Gateway отключён (`nil`), поэтому все API доступны
- Планируется: компонент `PrivateRoute`, который проверяет наличие токена в `localStorage` и редиректит на `/login`

## Что осталось разработать (TODO)
- [ ] Компонент `PrivateRoute` – защита маршрутов
- [ ] Кнопка «Logout» (очистка токена)
- [ ] Тёмная тема (переключатель)
- [ ] Адаптивная вёрстка для мобильных
- [ ] Загрузка реальных данных для прогноза и факторного анализа (сейчас используются статические/тестовые)
- [ ] Уведомления (toast) при ошибках
- [ ] Более детальная страница аукциона (история, лоты)
- [ ] Интеграция с реальным Auth после включения Gateway