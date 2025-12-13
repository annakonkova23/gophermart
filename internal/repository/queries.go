package repository

var selectAllBalance string = `
        SELECT balance, withdrawn, user_login 
        FROM accum_system.users_balance
    `

var updateUserBalance string = `
        INSERT INTO accum_system.users_balance (user_login, balance, withdrawn)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_login) DO UPDATE
        SET 
            balance = EXCLUDED.balance,
            withdrawn = EXCLUDED.withdrawn
    `
var insertwithdrawals string = `
        INSERT INTO accum_system.orders_withdrawals (user_login, number, sum, processed_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (number) DO NOTHING
    `

var selectWithdrawals string = `
        SELECT  number,
    			sum,
    			processed_at
        FROM accum_system.orders_withdrawals
        WHERE user_login = $1
    `

var selectBalanceUser string = `
        SELECT balance, withdrawn
        FROM accum_system.users_balance
        WHERE user_login = $1
    `

var updateOrderStatus string = `
        INSERT INTO accum_system.orders_status (number, status, accrual, uploaded_at)
        VALUES ($1, $2, $3, NOW())
        ON CONFLICT (number) DO UPDATE
        SET 
            status = EXCLUDED.status,
            accrual = EXCLUDED.accrual,
            uploaded_at = NOW()
    `

var selectOrderStatus string = `
        SELECT 
            o.number AS number,
            os.status,
            os.accrual,
            os.uploaded_at
        FROM accum_system.orders o
        JOIN accum_system.orders_status os ON o.number = os.number
        WHERE o.user_login = $1
        ORDER BY os.uploaded_at DESC
    `

var execCreateOrder string = `SELECT status, error_message FROM accum_system.create_order($1, $2)`

var selectNewOrderStatus string = `
        SELECT 
            o.number AS number,
            os.status,
            os.accrual,
            os.uploaded_at,
			o.user_login
        FROM accum_system.orders o
        JOIN accum_system.orders_status os ON o.number = os.number
        WHERE status not in ('PROCESSED', 'INVALID')
    `

var selectUserLogin string = `SELECT login, password FROM accum_system.users WHERE login = $1`

var insertUser string = `
		INSERT INTO accum_system.users (login, password)
		VALUES ($1, $2);
	`
