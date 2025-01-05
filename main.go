package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GOLANG-NINJA/pingrobot/workerpool"
)

const (
	INTERVAL        = time.Second * 3
	REQUEST_TIMEOUT = time.Second * 2
	WORKERS_COUNT   = 3
)

var urls = []string{
	"https://workshop.zhashkevych.com/",
	"https://golang-ninja.com/",
	"https://zhashkevych.com/",
	"https://google.com/",
	"https://golang.org/",
	"https://github.com/vadimkiryanov/pingrobot",
}

func main() {
	// Создаем канал для получения результатов пинга от воркеров
	results := make(chan workerpool.Result)

	// Создаем новый пул воркеров с указанным количеством работников,
	// таймаутом для каждого запроса и каналом для результатов
	workerPool := workerpool.New(WORKERS_COUNT, REQUEST_TIMEOUT, results)

	// Инициализируем пул воркеров - запускаем всех работников
	workerPool.Init()

	// Запускаем генерацию задач пинга в отдельной горутине,
	// чтобы не блокировать основной поток
	go generateJobs(workerPool)

	// Запускаем обработку результатов пинга в отдельной горутине,
	// которая читает из канала results
	go proccessResults(results)

	// Создаем канал для обработки сигналов завершения
	// Размер буфера 1, чтобы гарантированно не пропустить сигнал
	quit := make(chan os.Signal, 1)

	// Регистрируем обработку сигналов SIGTERM и SIGINT
	// Это позволяет gracefully завершить программу по Ctrl+C или команде kill
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	// Блокируем основной поток до получения сигнала завершения
	<-quit

	// При получении сигнала завершения останавливаем пул воркеров
	workerPool.Stop()
}

// proccessResults обрабатывает результаты пинга, получаемые от воркеров
// через канал results
func proccessResults(results chan workerpool.Result) {
	// Запускаем анонимную горутину для непрерывной обработки результатов
	go func() {
		// Читаем результаты из канала до его закрытия
		for result := range results {
			// Выводим информацию о результате пинга в консоль
			fmt.Println(result.Info())
		}
	}()
}

// generateJobs генерирует задачи пинга для пула воркеров
func generateJobs(wp *workerpool.Pool) {
	// Бесконечный цикл генерации задач
	for {
		// Перебираем все URL из списка
		for _, url := range urls {
			// Отправляем новую задачу в пул воркеров
			wp.Push(workerpool.Job{URL: url})
		}

		// Ждем указанный интервал перед следующей серией задач
		time.Sleep(INTERVAL)
	}
}
