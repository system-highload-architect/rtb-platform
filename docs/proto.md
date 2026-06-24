# Протоколы (Proto)

## auction.proto
- **AuctionService.Bid**: принимает `BidRequest` (device_id, ip, user_agent, lat/lng, campaign_ids, idempotency_key) и возвращает `BidResponse` (campaign_id, bid_price в Money, creative_url, error).

## accounting.proto
- **AccountingService.Debit**: списание средств (campaign_id, amount Money, bid_id), возвращает success, remaining_balance, error.
- **GetBalance**: запрос баланса кампании.

## analytics.proto
- **AnalyticsService.GetReport**: стриминг строк отчёта (dimensions, metrics).
- **Forecast**: прогноз временного ряда (history, horizon, параметры HoltWinters).
- **FactorAnalysis**: возвращает explained_variance_ratio.

## auth.proto
- Пока не реализован (сервис отсутствует).

## common/v1/money.proto
- Money: amount (int64, копейки), scale (int32, обычно 2).

## common/v1/user.proto
- UserProfile: device_id, ip, user_agent, lat/lng, features.