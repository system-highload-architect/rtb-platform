-- Миграция: создание таблицы и хранимых функций

CREATE TABLE IF NOT EXISTS balances (
    campaign_id VARCHAR(255) PRIMARY KEY,
    amount BIGINT NOT NULL DEFAULT 0,
    scale INT NOT NULL DEFAULT 2
);

-- Функция получения баланса (необязательно, можно и прямым SELECT)
CREATE OR REPLACE FUNCTION get_balance(campaign_id_in VARCHAR)
RETURNS TABLE(amount BIGINT, scale INT) AS $$
BEGIN
    RETURN QUERY SELECT b.amount, b.scale FROM balances b WHERE b.campaign_id = campaign_id_in;
END;
$$ LANGUAGE plpgsql;

-- Функция списания средств
CREATE OR REPLACE FUNCTION debit_balance(
    campaign_id_in VARCHAR,
    amount_in BIGINT,
    OUT success BOOLEAN,
    OUT remaining BIGINT,
    OUT scale_out INT,
    OUT error_msg TEXT
) AS $$
DECLARE
    current_amount BIGINT;
    current_scale INT;
BEGIN
    -- Блокируем строку
    SELECT b.amount, b.scale INTO current_amount, current_scale
    FROM balances b
    WHERE b.campaign_id = campaign_id_in
    FOR UPDATE;

    IF NOT FOUND THEN
        success := false;
        error_msg := 'Campaign not found';
        RETURN;
    END IF;

    IF current_amount < amount_in THEN
        success := false;
        remaining := current_amount;
        scale_out := current_scale;
        error_msg := 'Insufficient funds';
        RETURN;
    END IF;

    -- Списываем
    UPDATE balances b SET amount = current_amount - amount_in
    WHERE b.campaign_id = campaign_id_in;

    success := true;
    remaining := current_amount - amount_in;
    scale_out := current_scale;
    error_msg := '';
END;
$$ LANGUAGE plpgsql;