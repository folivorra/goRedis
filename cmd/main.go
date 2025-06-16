package main

import (
	"context"
	"github.com/folivorra/goRedis/internal/cli"
	"github.com/folivorra/goRedis/internal/config"
	"github.com/folivorra/goRedis/internal/controller"
	"github.com/folivorra/goRedis/internal/logger"
	"github.com/folivorra/goRedis/internal/persist"
	"github.com/folivorra/goRedis/internal/storage"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfgPath := os.Getenv("APP_CONFIG")
	if cfgPath == "" {
		cfgPath = "/app/config/app_config.yaml"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	err = logger.Init(cfg.Logger.LogFile)
	if err != nil {
		log.Fatal("Failed to init logger: ", err)
	}

	logger.InfoLogger.Println("Init logger")

	store := storage.NewInMemoryStorage() // хранилище

	rdb := storage.NewRedisClient() // клиент редиса
	logger.InfoLogger.Println("Init redis client")

	redisPersister := persist.NewRedisPersister(rdb, cfg.Storage.RedisKey)
	filePersister := persist.NewFilePersister(cfg.Storage.DumpFile)
	// сущности для дампинга и лоада из файла и из редиса

	data, err := redisPersister.Load()
	if err != nil || data == nil {
		logger.WarningLogger.Println("Failed to load data from redis:", err)
		data, err = filePersister.Load()
		if err != nil {
			logger.WarningLogger.Println("Failed to load data from file:", err)
		} else {
			logger.InfoLogger.Println("Loaded data from file")
		}
	} else {
		logger.InfoLogger.Println("Loaded data from redis")
	}
	if data != nil {
		store.Replace(data)
	}
	// сначал идем в редис за дампом, если ошибка, то идем в файл
	// если чтение из файла дает ошибку, то оставляем пустую мапу
	// если все хорошо и мы получили данные(даже пустые), то вписываем их в store

	itemController := controller.NewItemController(store) // контроллер
	router := mux.NewRouter()                             // маршрутизатор
	itemController.RegisterRoutes(router)                 // регистрация маршрутов по заданным методам
	logger.InfoLogger.Println("Creating controller, router and register routes were finished")

	srv := &http.Server{ // создаем объект сервера
		Addr:    cfg.Server.Port, // который будет слушать 8080 порт
		Handler: router,          // и обрабатываться маршрутизатором router
	}

	go func() {
		logger.InfoLogger.Println("Starting server on port 8080")
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.ErrorLogger.Printf("Failed to listen: %v", err)
		}
	}()

	// запускаем сервер в горутине чтобы не заблокироваться в main
	// сервер слушает порт и при возникновении ошибки (кроме ошибки graceful shutdown) аварийно завершает работу

	ctx, cancel := context.WithCancel(context.Background())
	// контекст для отмены долгоиграющих горутин

	go redisPersister.DumpForTTL(ctx, store)
	// дампер по ттл
	// дамп во время работы сервиса задает ттл в 2 минуты
	// в то же время дамп по окончанию работы сервиса сохраняет данные бессрочно

	go cli.RunCLI(ctx, store)
	// cli интерфейс - чтобы воспользоваться нужно зайти в терминал контейнера (в котором работает уже сервис)
	// с помощью команды docker attach myapp, чтобы выйти последовательно нажать Ctrl+P, Ctrl+Q

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.InfoLogger.Println("Shutdown server ...")
	cancel()
	// создаем канал прерывания, чтобы корректно обрабатывать нажатие Ctrl+C, вызывая GS

	snapshot := store.Snapshot()
	if err = redisPersister.Dump(snapshot, 0); err != nil {
		logger.ErrorLogger.Println("Failed to dump snapshot in redis")
	} else {
		logger.InfoLogger.Println("Snapshot dumped in redis")
	}
	if err = filePersister.Dump(snapshot); err != nil {
		logger.ErrorLogger.Println("Failed to dump snapshot in file")
	} else {
		logger.InfoLogger.Println("Snapshot dumped in file")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	// задаем контекст отмены в 5 секунд, откладывая освобождение ресурсов

	if err = srv.Shutdown(shutdownCtx); err != nil {
		logger.ErrorLogger.Println("Failed to gracefully shutdown server")
	} else {
		logger.InfoLogger.Println("Server gracefully shutdown")
		logger.InfoLogger.Println("Server exiting")
	}
	// даем серверу мягко завершится за эти 5 секунд, иначе завершаем аварийно
}
