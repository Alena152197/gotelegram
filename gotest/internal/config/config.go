package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config — главная структура конфигурации приложения
// Все поля заполняются из переменных окружения
type Config struct {
	Bot      BotConfig      // Настройки бота
	Database DatabaseConfig // Настройки базы данных
	Logging  LoggingConfig  // Настройки логирования
}

// BotConfig — настройки Telegram-бота
type BotConfig struct {
	Token    string  `envconfig:"BOT_TOKEN" required:"true"` // Токен бота (обязательный)
	Debug    bool    `envconfig:"BOT_DEBUG" default:"false"` // Режим отладки
	Timeout  int     `envconfig:"BOT_TIMEOUT" default:"60"`  // Таймаут запросов (секунды)
	AdminIDs []int64 `envconfig:"ADMIN_IDS"`                 // ID администраторов
}

// DatabaseConfig — настройки подключения к PostgreSQL
type DatabaseConfig struct {
	Host     string `envconfig:"DB_HOST" default:"localhost"`    // Адрес сервера БД
	Port     int    `envconfig:"DB_PORT" default:"5432"`         // Порт БД
	Name     string `envconfig:"DB_NAME" default:"telegram_bot"` // Имя базы данных
	User     string `envconfig:"DB_USER" default:"postgres"`     // Пользователь БД
	Password string `envconfig:"DB_PASSWORD" default:""`         // Пароль БД
	SSLMode  string `envconfig:"DB_SSL_MODE" default:"disable"`  // Режим SSL
}

// LoggingConfig — настройки логирования
type LoggingConfig struct {
	Level string `envconfig:"LOG_LEVEL" default:"info"`   // Уровень логирования (debug, info, warn, error)
	File  string `envconfig:"LOG_FILE" default:"bot.log"` // Файл для логов
}

// Load загружает конфигурацию из переменных окружения
// Сначала пытается прочитать файл .env, затем читает переменные окружения
func Load() (*Config, error) {
	// Пытаемся загрузить .env файл
	// Если файла нет — не страшно, будем читать из системных переменных
	_ = godotenv.Load()

	// Создаём пустую структуру конфигурации
	var cfg Config

	// Заполняем структуру из переменных окружения
	// Если обязательное поле отсутствует — вернётся ошибка
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}

	// Парсим список ID администраторов из строки
	// ADMIN_IDS="123456789,987654321" -> []int64{123456789, 987654321}
	if err := parseAdminIDs(&cfg); err != nil {
		return nil, err
	}

	// Возвращаем указатель на конфигурацию
	return &cfg, nil
}

// parseAdminIDs парсит строку ADMIN_IDS и заполняет BotConfig.AdminIDs
func parseAdminIDs(cfg *Config) error {
	// Получаем значение переменной окружения ADMIN_IDS
	adminIDsStr := os.Getenv("ADMIN_IDS")

	// Если переменная не задана — возвращаем nil (это не ошибка)
	if adminIDsStr == "" {
		cfg.Bot.AdminIDs = []int64{}
		return nil
	}

	// Разделяем строку по запятым: "123,456,789" -> ["123", "456", "789"]
	parts := strings.Split(adminIDsStr, ",")

	// Создаём срез для хранения ID
	adminIDs := make([]int64, 0, len(parts))

	// Перебираем каждую часть
	for _, part := range parts {
		// Убираем пробелы с начала и конца
		part = strings.TrimSpace(part)

		// Пропускаем пустые строки
		if part == "" {
			continue
		}

		// Преобразуем строку в число (int64)
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return err
		}

		// Добавляем ID в срез
		adminIDs = append(adminIDs, id)
	}

	// Сохраняем результат в конфигурацию
	cfg.Bot.AdminIDs = adminIDs

	return nil
}
