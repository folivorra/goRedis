package logger

import (
	"io"
	"log"
	"os"
)

var (
	ErrorLogger   *log.Logger
	InfoLogger    *log.Logger
	WarningLogger *log.Logger
)

func Init(logFilename string) error {
	file, err := os.OpenFile(logFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	multiWriter := io.MultiWriter(os.Stdout, file)
	// мультирайтер который будет писать и в терминал и в файл

	InfoLogger = log.New(multiWriter, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(multiWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(multiWriter, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	// сущности логгеров с нужным префиксом, датой, временем и файлом + строкой

	return nil
}
