CREATE OR REPLACE FUNCTION accum_system.create_order(
    p_number VARCHAR,
    p_user_login VARCHAR
)
RETURNS TABLE(status TEXT, error_message TEXT)
LANGUAGE plpgsql
AS $$
DECLARE
    existing_login VARCHAR;
BEGIN
    -- Проверяем, существует ли заказ
    SELECT user_login INTO existing_login
    FROM accum_system.orders
    WHERE number = p_number;

    IF FOUND THEN
        IF existing_login = p_user_login THEN
            RETURN QUERY SELECT 'already_exists'::TEXT, 'Order already exists for this user'::TEXT;
        ELSE
            RETURN QUERY SELECT 'conflict'::TEXT, 'Order belongs to another user'::TEXT;
        END IF;
    ELSE
        -- Вставляем в orders
        INSERT INTO accum_system.orders (number, user_login)
        VALUES (p_number, p_user_login);

        -- Вставляем в orders_status со статусом 'NEW'
        INSERT INTO accum_system.orders_status (number, status)
        VALUES (p_number, 'NEW');

        RETURN QUERY SELECT 'success'::TEXT, NULL::TEXT;
    END IF;

    RETURN;
END;
$$;