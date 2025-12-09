package config

import (
	"flag"
	"os"
	"strconv"
	"time"
)

const (
	defaultHost           = "localhost:8080"
	defaultDBUri          = "postgres://user_main:user_main@localhost:5432/gophermartdb?sslmode=disable"
	defaultAccrualAddress = "http://localhost:8090"
	defaultTimeout        = 10
	defaultCountProcess   = 10
	defaultBufferSize     = 1000
)

type Config struct {
	Host           string
	DBUri          string
	AccrualAddress string
	Timeout        time.Duration
	CountProcess   int
	BufferSize     int
}

func getEnvString(envKey, defaultValue string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(envKey string, defaultValue int) int {
	if v := os.Getenv(envKey); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}

func GetConfig() *Config {
	hostFlag := flag.String("a", defaultHost, "Адрес запуска HTTP-сервера")
	dbUriFlag := flag.String("d", defaultDBUri, "Адрес подключения к базе данных")
	accrualAddressFlag := flag.String("r", defaultAccrualAddress, "Aдрес системы расчёта начислений")
	countProcessFlag := flag.Int("p", defaultCountProcess, "Количество worker")
	bufferSizeFlag := flag.Int("b", defaultBufferSize, "Размер буфера каналов")
	timeoutFlag := flag.Int("t", defaultTimeout, "Таймаут запросов")
	flag.Parse()

	host := getEnvString("RUN_ADDRESS", *hostFlag)
	dbUri := getEnvString("DATABASE_URI", *dbUriFlag)
	accrualAddress := getEnvString("ACCRUAL_SYSTEM_ADDRESS", *accrualAddressFlag)
	countProcess := getEnvInt("COUNT_PROCESS", *countProcessFlag)
	bufferSize := getEnvInt("BUFFER_SIZE", *bufferSizeFlag)
	timeout := getEnvInt("TIMEOUT_REQUEST", *timeoutFlag)

	return &Config{
		Host:           host,
		DBUri:          dbUri,
		AccrualAddress: accrualAddress,
		Timeout:        time.Second * time.Duration(timeout),
		CountProcess:   countProcess,
		BufferSize:     bufferSize,
	}
}
