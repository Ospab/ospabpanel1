-- SQL команды для настройки MySQL для OSPAB Panel
-- Выполните эти команды в MySQL консоли (mysql -u root -p)

-- 1. Создание базы данных
CREATE DATABASE IF NOT EXISTS ospab_panel CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 2. Создание пользователя с простым паролем
CREATE USER 'ospab'@'localhost' IDENTIFIED BY 'ospab123';

-- 3. Выдача всех прав на базу данных
GRANT ALL PRIVILEGES ON ospab_panel.* TO 'ospab'@'localhost';

-- 4. Обновление привилегий
FLUSH PRIVILEGES;

-- 5. Проверка созданного пользователя (опционально)
SELECT User, Host FROM mysql.user WHERE User = 'ospab';

-- 6. Показать все базы данных (опционально)
SHOW DATABASES;

-- Структура таблицы users (расширенная)
CREATE TABLE IF NOT EXISTS users (
	id INT AUTO_INCREMENT PRIMARY KEY,
	username VARCHAR(64) NOT NULL UNIQUE,
	email VARCHAR(128) NOT NULL UNIQUE,
	password_hash VARCHAR(255) NOT NULL,
	password_salt VARCHAR(64) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Таблица серверов гипервизоров (чувствительные поля будут зашифрованы / захешированы на уровне приложения позже)
CREATE TABLE IF NOT EXISTS servers (
	id INT AUTO_INCREMENT PRIMARY KEY,
	name VARCHAR(128) NOT NULL,
	host VARCHAR(255) NOT NULL,
	port INT NOT NULL DEFAULT 0,
	type CHAR(3) NOT NULL,
	username_enc TEXT NOT NULL,
	password_enc TEXT NOT NULL,
	user_id INT NOT NULL,
	is_active TINYINT(1) NOT NULL DEFAULT 1,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	INDEX (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Индексы безопасности
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_servers_user ON servers(user_id);

-- Использование:
-- 1. Откройте MySQL консоль: mysql -u root -p
-- 2. Введите пароль root пользователя
-- 3. Скопируйте и выполните команды выше
-- 4. Выйдите из MySQL: EXIT;
-- 5. Теперь можете запускать OSPAB Panel