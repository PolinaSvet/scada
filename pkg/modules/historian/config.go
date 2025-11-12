package historian

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// loadConfig загружает конфигурацию из файла
func (hist *Historian) loadConfig(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	var config HistorianConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	// Дешифруем пароли
	if err := hist.decryptPasswords(&config); err != nil {
		return fmt.Errorf("ошибка дешифровки паролей: %w", err)
	}

	hist.config = config

	if err := hist.validateConfig(); err != nil {
		return fmt.Errorf("ошибка валидации конфига: %w", err)
	}

	return nil
}

// "password": "enc:U2FsdGVkX1+WvFJzW1kQy8K5t6M8V9p7R2XbL3aN4cO0=",
// decryptPasswords дешифрует пароли в конфигурации
func (hist *Historian) decryptPasswords(config *HistorianConfig) error {
	if config.Alarm.Enable {
		decrypted, err := hist.decryptPassword(config.Alarm.ConnectSettings.Password)
		if err != nil {
			return fmt.Errorf("ошибка дешифровки пароля Alarm: %w", err)
		}
		config.Alarm.ConnectSettings.DecryptedPassword = decrypted
	}

	if config.Trend.Enable {
		decrypted, err := hist.decryptPassword(config.Trend.ConnectSettings.Password)
		if err != nil {
			return fmt.Errorf("ошибка дешифровки пароля Trend: %w", err)
		}
		config.Trend.ConnectSettings.DecryptedPassword = decrypted
	}

	return nil
}

// decryptPassword дешифрует пароль если он зашифрован
func (hist *Historian) decryptPassword(password string) (string, error) {
	// Если пароль начинается с "enc:", значит он зашифрован
	if len(password) > 4 && password[:4] == "enc:" {
		// Здесь вызываем функцию дешифровки
		decrypted, err := DecryptPassword(password[4:])
		if err != nil {
			return "", err
		}
		return decrypted, nil
	}
	// Если не зашифрован, возвращаем как есть
	return password, nil
}

// validateConfig валидирует конфигурацию
func (hist *Historian) validateConfig() error {
	if hist.config.ID == "" {
		return fmt.Errorf("ID не может быть пустым")
	}

	if hist.config.Alarm.Enable {
		if err := hist.validateConnectSettings(&hist.config.Alarm.ConnectSettings, "Alarm"); err != nil {
			return err
		}
	}

	if hist.config.Trend.Enable {
		if err := hist.validateConnectSettings(&hist.config.Trend.ConnectSettings, "Trend"); err != nil {
			return err
		}
	}

	return nil
}

// validateConnectSettings валидирует настройки подключения
func (hist *Historian) validateConnectSettings(settings *ConnectSettings, service string) error {
	if settings.Host == "" {
		return fmt.Errorf("host не может быть пустым для %s", service)
	}
	if settings.Port == 0 {
		return fmt.Errorf("port не может быть 0 для %s", service)
	}
	if settings.User == "" {
		return fmt.Errorf("user не может быть пустым для %s", service)
	}
	if settings.DecryptedPassword == "" {
		return fmt.Errorf("password не может быть пустым для %s", service)
	}
	if settings.DBName == "" {
		return fmt.Errorf("dbname не может быть пустым для %s", service)
	}
	return nil
}

// IsEnabled проверяет, включены ли настройки подключения
func (c *ConnectSettings) IsEnabled() bool {
	return c.Host != "" && c.Port > 0 && c.User != "" && c.DecryptedPassword != "" && c.DBName != ""
}

// GetConnectionString возвращает строку подключения с дешифрованным паролем
func (c *ConnectSettings) GetConnectionString() string {
	password := c.DecryptedPassword
	if password == "" {
		password = c.Password // fallback
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, password, c.DBName, c.SSLMode)
}

// GetAlarmConnectionString возвращает строку подключения для Alarm с дешифрованным паролем
func (c *ConnectSettings) GetAlarmConnectionString() string {
	if c.DecryptedPassword != "" {
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.DecryptedPassword, c.DBName, c.SSLMode)
	}
	// Fallback на зашифрованный пароль (для обратной совместимости)
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetTrendConnectionString возвращает строку подключения для Trend с дешифрованным паролем
func (c *ConnectSettings) GetTrendConnectionString() string {
	if c.DecryptedPassword != "" {
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.DecryptedPassword, c.DBName, c.SSLMode)
	}
	// Fallback на зашифрованный пароль (для обратной совместимости)
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}
