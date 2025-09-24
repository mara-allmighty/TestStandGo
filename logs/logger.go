package logs

import (
	"log"
	"os"
	"path/filepath"
)

func SetLogger(filePath string) *os.File {
	// Создаём директорию, если она не существует
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		panic("Не удалось создать директорию для логов: " + err.Error())
	}

	// Открываем файл с правильными флагами: запись + создать + дописать
	logFile, err := os.OpenFile(
		filePath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, // ← обязательно O_WRONLY!
		0o666,
	)
	if err != nil {
		panic("Не удалось открыть лог-файл: " + err.Error())
	}

	// Перенаправляем стандартный логгер Go
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("Логгер успешно инициализирован")

	// ВАЖНО: возвращаем файл, чтобы main мог его закрыть при завершении
	return logFile
}
