CREATE SCHEMA IF NOT EXISTS accum_system AUTHORIZATION current_user;
   

GRANT USAGE, CREATE ON SCHEMA accum_system TO current_user;

CREATE TABLE IF NOT EXISTS accum_system.users (
    login VARCHAR(1000) PRIMARY KEY NOT NULL,
    password VARCHAR(1000) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
 );

 CREATE TABLE IF NOT EXISTS accum_system.orders (
    number VARCHAR(1000) PRIMARY KEY NOT NULL,
    user_login VARCHAR(1000) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
 );


CREATE TABLE IF NOT EXISTS accum_system.orders_status (
    number VARCHAR(1000) PRIMARY KEY NOT NULL,
    status VARCHAR(1000) NOT NULL,
    uploaded_at TIMESTAMP DEFAULT NOW(),
    accrual decimal(10,4)
);

CREATE TABLE IF NOT EXISTS accum_system.orders_withdrawals (
    user_login VARCHAR(1000) NOT NULL,
    number VARCHAR(1000) PRIMARY KEY NOT NULL,
    sum decimal(10,4) NOT NULL,
    processed_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS accum_system.users_balance (
    user_login VARCHAR(1000) PRIMARY KEY NOT NULL,
    balance decimal(10,4),
    withdrawn decimal(10,4)
);