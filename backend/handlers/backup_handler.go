package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"shoes-store-backend/models"
)

// CreateBackupHandler создает полный бэкап базы данных
// @Summary Создание бэкапа базы данных
// @Description Создает полный бэкап базы данных и сохраняет его в папку проекта
// @Tags Admin Backup
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.BackupResponse
// @Failure 400 {string} string "Ошибка валидации"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /admin/backup [post]
func CreateBackupHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("🔄 Начинаем создание бэкапа базы данных...")

	// Получаем настройки подключения к БД из переменных окружения
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:1@localhost:5432/ShoesStoreDB"
	}

	// Парсим URL для получения параметров подключения
	dbParams, err := parseDatabaseURL(databaseURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка парсинга URL базы данных: %v", err), http.StatusInternalServerError)
		return
	}

	// Создаем имя файла с текущей датой и временем
	timestamp := time.Now().Format("20060102_150405")
	backupFilename := fmt.Sprintf("shoes_store_backup_%s.sql", timestamp)

	// Получаем путь для сохранения бэкапов
	backupDir, err := getBackupPath()
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка получения пути для бэкапов: %v", err), http.StatusInternalServerError)
		return
	}

	backupPath := filepath.Join(backupDir, backupFilename)

	fmt.Printf("📁 Путь к бэкапу: %s\n", backupPath)
	fmt.Printf("🗄️ Параметры БД: %s@%s:%s/%s\n", dbParams.Username, dbParams.Host, dbParams.Port, dbParams.Database)

	// Проверяем версию pg_dump и получаем правильный путь
	pgDumpPath := checkPgDumpVersion()

	// Команда для создания бэкапа
	pgDumpCmd := []string{
		"--host=" + dbParams.Host,
		"--port=" + dbParams.Port,
		"--username=" + dbParams.Username,
		"--dbname=" + dbParams.Database,
		"--verbose",
		"--clean",
		"--no-owner",
		"--no-privileges",
		"--no-tablespaces", // Добавляем для совместимости
		"--file=" + backupPath,
	}

	// Устанавливаем переменную окружения для пароля
	env := os.Environ()
	env = append(env, "PGPASSWORD="+dbParams.Password)

	// Выполняем команду бэкапа
	fmt.Printf("⚙️ Выполняем команду: %s\n", pgDumpPath)
	cmd := exec.Command(pgDumpPath, pgDumpCmd...)
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Ошибка выполнения pg_dump: %v\n", err)
		fmt.Printf("📋 Вывод команды: %s\n", string(output))
		
		// Проверяем, если это ошибка версии
		if strings.Contains(string(output), "server version mismatch") {
			fmt.Println("🔄 Пробуем альтернативный способ с игнорированием версии...")
			
			// Пробуем с флагом --no-sync для игнорирования версии
			pgDumpCmdAlt := []string{
				"--host=" + dbParams.Host,
				"--port=" + dbParams.Port,
				"--username=" + dbParams.Username,
				"--dbname=" + dbParams.Database,
				"--verbose",
				"--clean",
				"--no-owner",
				"--no-privileges",
				"--no-tablespaces",
				"--no-sync", // Игнорируем версию
				"--file=" + backupPath,
			}
			
			cmdAlt := exec.Command(pgDumpPath, pgDumpCmdAlt...)
			cmdAlt.Env = env
			
			outputAlt, errAlt := cmdAlt.CombinedOutput()
			if errAlt != nil {
				fmt.Printf("❌ Альтернативный способ тоже не сработал: %v\n", errAlt)
				fmt.Printf("📋 Вывод: %s\n", string(outputAlt))
				
				// Последняя попытка - простой бэкап без дополнительных флагов
				fmt.Println("🔄 Последняя попытка - простой бэкап...")
				pgDumpCmdSimple := []string{
					"--host=" + dbParams.Host,
					"--port=" + dbParams.Port,
					"--username=" + dbParams.Username,
					"--dbname=" + dbParams.Database,
					"--file=" + backupPath,
				}
				
				cmdSimple := exec.Command(pgDumpPath, pgDumpCmdSimple...)
				cmdSimple.Env = env
				
				outputSimple, errSimple := cmdSimple.CombinedOutput()
				if errSimple != nil {
					fmt.Printf("❌ Все попытки не удались: %v\n", errSimple)
					fmt.Printf("📋 Вывод: %s\n", string(outputSimple))
					http.Error(w, fmt.Sprintf("Ошибка создания бэкапа. Возможно, нужно обновить pg_dump до версии PostgreSQL 17. Ошибка: %v\nВывод: %s", errSimple, string(outputSimple)), http.StatusInternalServerError)
					return
				}
				
				fmt.Println("✅ Простой бэкап создан успешно!")
			} else {
				fmt.Println("✅ Альтернативный бэкап создан успешно!")
			}
		} else {
			http.Error(w, fmt.Sprintf("Ошибка создания бэкапа: %v\nВывод: %s", err, string(output)), http.StatusInternalServerError)
			return
		}
	} else {
		fmt.Println("✅ Стандартный бэкап создан успешно!")
	}

	// Проверяем, что файл создался
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		http.Error(w, "Файл бэкапа не был создан", http.StatusInternalServerError)
		return
	}

	// Получаем размер файла
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка получения информации о файле: %v", err), http.StatusInternalServerError)
		return
	}

	fileSize := fileInfo.Size()
	fileSizeMB := float64(fileSize) / (1024 * 1024)

	fmt.Printf("✅ Бэкап успешно создан!\n")
	fmt.Printf("📄 Файл: %s\n", backupFilename)
	fmt.Printf("📊 Размер: %.2f MB (%d байт)\n", fileSizeMB, fileSize)

	// Логируем создание бэкапа
	LogUserAction(r, "CREATE", "backup", 0, fmt.Sprintf("Создан бэкап базы данных: %s (%.2f MB)", backupFilename, fileSizeMB))

	response := models.BackupResponse{
		Message:  fmt.Sprintf("Бэкап успешно создан. Размер файла: %.2f MB (%d байт)", fileSizeMB, fileSize),
		Success:  true,
		FilePath: backupPath,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DatabaseParams представляет параметры подключения к базе данных
type DatabaseParams struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

// parseDatabaseURL парсит URL базы данных и возвращает параметры подключения
func parseDatabaseURL(databaseURL string) (*DatabaseParams, error) {
	// Формат: postgresql://username:password@host:port/database
	// или: postgres://username:password@host:port/database
	
	// Убираем префикс postgresql:// или postgres://
	url := strings.TrimPrefix(databaseURL, "postgresql://")
	url = strings.TrimPrefix(url, "postgres://")
	
	if !strings.Contains(url, "@") {
		return nil, fmt.Errorf("неверный формат URL базы данных")
	}
	
	// Разделяем на части: auth@host/database
	parts := strings.Split(url, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("неверный формат URL базы данных")
	}
	
	authPart := parts[0]
	hostPart := parts[1]
	
	// Парсим аутентификацию: username:password
	authParts := strings.Split(authPart, ":")
	if len(authParts) != 2 {
		return nil, fmt.Errorf("неверный формат аутентификации")
	}
	
	username := authParts[0]
	password := authParts[1]
	
	// Парсим хост и базу данных: host:port/database
	hostDbParts := strings.Split(hostPart, "/")
	if len(hostDbParts) != 2 {
		return nil, fmt.Errorf("неверный формат хоста и базы данных")
	}
	
	hostPort := hostDbParts[0]
	database := hostDbParts[1]

	// Добавляем эту проверку
	if strings.Contains(database, "?") {
		database = strings.Split(database, "?")[0]
	}

	
	// Парсим хост и порт
	var host, port string
	if strings.Contains(hostPort, ":") {
		hostPortParts := strings.Split(hostPort, ":")
		host = hostPortParts[0]
		port = hostPortParts[1]
	} else {
		host = hostPort
		port = "5432" // Порт по умолчанию для PostgreSQL
	}
	
	return &DatabaseParams{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: database,
	}, nil
}

// getBackupPath возвращает путь для сохранения бэкапов
func getBackupPath() (string, error) {
    backupDir := os.Getenv("BACKUP_PATH")
    if backupDir == "" {
        // fallback на старый путь
        projectPath := "C:\\shoes-store"
        backupDir = filepath.Join(projectPath, "backups")
    }

    if err := os.MkdirAll(backupDir, 0755); err != nil {
        return "", fmt.Errorf("ошибка создания папки backups: %v", err)
    }

    fmt.Printf("📁 Папка для бэкапов: %s\n", backupDir)
    return backupDir, nil
}


// GetBackupInfoHandler получает информацию о доступных бэкапах
// @Summary Получение информации о бэкапах
// @Description Получает список всех доступных файлов бэкапа в папке проекта
// @Tags Admin Backup
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.BackupListResponse
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /admin/backup/info [get]
func GetBackupInfoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("📋 Получаем информацию о бэкапах...")

	backupDir, err := getBackupPath()
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка получения пути для бэкапов: %v", err), http.StatusInternalServerError)
		return
	}

	// Читаем содержимое директории
	files, err := os.ReadDir(backupDir)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения директории: %v", err), http.StatusInternalServerError)
		return
	}

	var backupFiles []models.BackupInfo

	// Ищем файлы бэкапа
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "shoes_store_backup_") && strings.HasSuffix(file.Name(), ".sql") {
			filePath := filepath.Join(backupDir, file.Name())
			
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				fmt.Printf("⚠️ Ошибка получения информации о файле %s: %v\n", file.Name(), err)
				continue
			}

			backupInfo := models.BackupInfo{
				Filename:  file.Name(),
				Path:      filePath,
				SizeBytes: fileInfo.Size(),
				SizeMB:    float64(fileInfo.Size()) / (1024 * 1024),
				Created:   fileInfo.ModTime(),
			}

			backupFiles = append(backupFiles, backupInfo)
		}
	}

	// Сортируем по дате создания (новые сверху)
	for i := 0; i < len(backupFiles)-1; i++ {
		for j := i + 1; j < len(backupFiles); j++ {
			if backupFiles[i].Created.Before(backupFiles[j].Created) {
				backupFiles[i], backupFiles[j] = backupFiles[j], backupFiles[i]
			}
		}
	}

	fmt.Printf("📊 Найдено файлов бэкапа: %d\n", len(backupFiles))

	response := models.BackupListResponse{
		BackupFiles: backupFiles,
		TotalFiles:  len(backupFiles),
		DesktopPath: backupDir,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteBackupHandler удаляет файл бэкапа
// @Summary Удаление файла бэкапа
// @Description Удаляет указанный файл бэкапа из папки проекта
// @Tags Admin Backup
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param filename path string true "Имя файла бэкапа"
// @Success 200 {object} models.BackupDeleteResponse
// @Failure 400 {string} string "Неверный формат файла"
// @Failure 404 {string} string "Файл бэкапа не найден"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /admin/backup/{filename} [delete]
func DeleteBackupHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем имя файла из URL
	filename := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	
	fmt.Printf("🗑️ Удаляем файл бэкапа: %s\n", filename)

	// Проверяем формат имени файла
	if !strings.HasPrefix(filename, "shoes_store_backup_") || !strings.HasSuffix(filename, ".sql") {
		http.Error(w, "Неверный формат файла. Файл должен начинаться с 'shoes_store_backup_' и заканчиваться на '.sql'", http.StatusBadRequest)
		return
	}

	backupDir, err := getBackupPath()
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка получения пути для бэкапов: %v", err), http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(backupDir, filename)

	// Проверяем, что файл существует
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Файл бэкапа не найден", http.StatusNotFound)
		return
	}

	// Удаляем файл
	err = os.Remove(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка удаления файла: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Printf("✅ Файл бэкапа %s успешно удален\n", filename)

	// Логируем удаление бэкапа
	LogUserAction(r, "DELETE", "backup", 0, fmt.Sprintf("Удален файл бэкапа: %s", filename))

	response := models.BackupDeleteResponse{
		Message: fmt.Sprintf("Файл бэкапа %s удален", filename),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DownloadBackupHandler скачивает файл бэкапа
// @Summary Скачать файл бэкапа
// @Description Скачивает указанный файл бэкапа
// @Tags Admin Backup
// @Produce application/octet-stream
// @Security BearerAuth
// @Param filename path string true "Имя файла бэкапа"
// @Success 200 {file} file "SQL файл бэкапа"
// @Failure 400 {string} string "Неверный формат файла"
// @Failure 404 {string} string "Файл бэкапа не найден"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /admin/backup/download/{filename} [get]
func DownloadBackupHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем имя файла из URL через mux
	vars := mux.Vars(r)
	filename := vars["filename"]
	
	fmt.Printf("📥 Скачиваем файл бэкапа: %s\n", filename)

	// Проверяем формат имени файла
	if !strings.HasPrefix(filename, "shoes_store_backup_") || !strings.HasSuffix(filename, ".sql") {
		http.Error(w, "Неверный формат файла. Файл должен начинаться с 'shoes_store_backup_' и заканчиваться на '.sql'", http.StatusBadRequest)
		return
	}

	backupDir, err := getBackupPath()
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка получения пути для бэкапов: %v", err), http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(backupDir, filename)

	// Проверяем, что файл существует
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Файл бэкапа не найден", http.StatusNotFound)
		return
	}

	// Устанавливаем заголовки для скачивания файла
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Transfer-Encoding", "binary")

	// Открываем и отправляем файл
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка открытия файла: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Копируем файл в ответ
	http.ServeFile(w, r, filePath)
}

// RestoreBackupHandler восстанавливает базу данных из файла бэкапа
// @Summary Восстановить базу данных из бэкапа
// @Description Восстанавливает базу данных из указанного файла бэкапа
// @Tags Admin Backup
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "SQL файл бэкапа"
// @Success 200 {object} models.BackupResponse
// @Failure 400 {string} string "Неверный формат файла"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /admin/backup/restore [post]
func RestoreBackupHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("🔄 Начинаем восстановление базы данных из бэкапа...")

	// Получаем файл из формы
	err := r.ParseMultipartForm(100 << 20) // 100 MB max
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка парсинга формы: %v", err), http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка получения файла: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Проверяем расширение файла
	if !strings.HasSuffix(fileHeader.Filename, ".sql") {
		http.Error(w, "Файл должен иметь расширение .sql", http.StatusBadRequest)
		return
	}

	// Получаем настройки подключения к БД из переменных окружения
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:1@localhost:5432/ShoesStoreDB"
	}

	// Парсим URL для получения параметров подключения
	dbParams, err := parseDatabaseURL(databaseURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка парсинга URL базы данных: %v", err), http.StatusInternalServerError)
		return
	}

	// Создаем временный файл для сохранения загруженного бэкапа
	tempDir := os.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "restore_backup_*.sql")
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка создания временного файла: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Копируем содержимое загруженного файла во временный файл
	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка сохранения файла: %v", err), http.StatusInternalServerError)
		return
	}
	tempFile.Close()

	// Проверяем версию psql и получаем правильный путь
	psqlPath := findPsqlPath()

	// Команда для восстановления базы данных
	psqlCmd := []string{
		"--host=" + dbParams.Host,
		"--port=" + dbParams.Port,
		"--username=" + dbParams.Username,
		"--dbname=" + dbParams.Database,
		"--file=" + tempFile.Name(),
	}

	// Устанавливаем переменную окружения для пароля
	env := os.Environ()
	env = append(env, "PGPASSWORD="+dbParams.Password)

	// Выполняем команду восстановления
	fmt.Printf("⚙️ Выполняем команду восстановления: %s\n", psqlPath)
	cmd := exec.Command(psqlPath, psqlCmd...)
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Ошибка выполнения psql: %v\n", err)
		fmt.Printf("📋 Вывод команды: %s\n", string(output))
		http.Error(w, fmt.Sprintf("Ошибка восстановления базы данных: %v\nВывод: %s", err, string(output)), http.StatusInternalServerError)
		return
	}

	fmt.Println("✅ База данных успешно восстановлена из бэкапа!")

	// Логируем восстановление бэкапа
	LogUserAction(r, "RESTORE", "backup", 0, fmt.Sprintf("Восстановлена база данных из файла: %s", fileHeader.Filename))

	response := models.BackupResponse{
		Message: fmt.Sprintf("База данных успешно восстановлена из файла: %s", fileHeader.Filename),
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// findPsqlPath находит правильный путь к psql
func findPsqlPath() string {
	// Возможные пути к psql на Windows
	possiblePaths := []string{
		"psql", // Обычный путь в PATH
		"C:\\Program Files\\PostgreSQL\\17\\bin\\psql.exe",
		"C:\\Program Files (x86)\\PostgreSQL\\17\\bin\\psql.exe",
		"C:\\PostgreSQL\\17\\bin\\psql.exe",
		"C:\\Users\\mreax\\AppData\\Local\\Programs\\PostgreSQL\\17\\bin\\psql.exe",
	}
	
	for _, path := range possiblePaths {
		cmd := exec.Command(path, "--version")
		output, err := cmd.Output()
		if err == nil {
			version := strings.TrimSpace(string(output))
			fmt.Printf("🔍 Найден psql: %s - %s\n", path, version)
			return path
		}
	}
	
	fmt.Printf("⚠️ Не найден psql, используем стандартный путь\n")
	return "psql"
}

// findPgDumpPath находит правильный путь к pg_dump версии 17
func findPgDumpPath() string {
	// Возможные пути к pg_dump на Windows
	possiblePaths := []string{
		"pg_dump", // Обычный путь в PATH
		"C:\\Program Files\\PostgreSQL\\17\\bin\\pg_dump.exe",
		"C:\\Program Files (x86)\\PostgreSQL\\17\\bin\\pg_dump.exe",
		"C:\\PostgreSQL\\17\\bin\\pg_dump.exe",
		"C:\\Users\\mreax\\AppData\\Local\\Programs\\PostgreSQL\\17\\bin\\pg_dump.exe",
	}
	
	for _, path := range possiblePaths {
		cmd := exec.Command(path, "--version")
		output, err := cmd.Output()
		if err == nil {
			version := strings.TrimSpace(string(output))
			fmt.Printf("🔍 Найден pg_dump: %s - %s\n", path, version)
			
			// Проверяем, что это версия 17
			if strings.Contains(version, "17.") {
				fmt.Printf("✅ Используем правильную версию: %s\n", path)
				return path
			}
		}
	}
	
	fmt.Printf("⚠️ Не найдена версия pg_dump 17, используем стандартный путь\n")
	return "pg_dump"
}

// checkPgDumpVersion проверяет версию pg_dump и выводит предупреждение если нужно
func checkPgDumpVersion() string {
	pgDumpPath := findPgDumpPath()
	
	cmd := exec.Command(pgDumpPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("⚠️ Не удалось проверить версию pg_dump: %v\n", err)
		return pgDumpPath
	}
	
	version := strings.TrimSpace(string(output))
	fmt.Printf("🔍 Версия pg_dump: %s\n", version)
	
	// Проверяем, если версия старая
	if strings.Contains(version, "11.") || strings.Contains(version, "12.") || 
	   strings.Contains(version, "13.") || strings.Contains(version, "14.") ||
	   strings.Contains(version, "15.") || strings.Contains(version, "16.") {
		fmt.Printf("⚠️ ВНИМАНИЕ: У вас старая версия pg_dump (%s), а PostgreSQL 17.4\n", version)
		fmt.Printf("💡 Рекомендации:\n")
		fmt.Printf("   1. Обновите pg_dump до версии 17.x\n")
		fmt.Printf("   2. Или используйте pg_dump из PostgreSQL 17\n")
		fmt.Printf("   3. Система попробует создать бэкап с обходными путями\n")
	}
	
	return pgDumpPath
}
