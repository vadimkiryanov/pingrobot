package workerpool

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// New создает новый пул воркеров
func New(workersCount int, timeout time.Duration, results chan Result) *Pool {
	return &Pool{
		worker:       newWorker(timeout),
		workersCount: workersCount,
		jobs:         make(chan Job),
		results:      results,
		wg:           new(sync.WaitGroup),
	}
}

// Job представляет собой задачу для обработки - содержит URL для пинга
type Job struct {
	URL string
}

// Result содержит результат выполнения пинга
type Result struct {
	URL          string        // URL, который пинговали
	StatusCode   int           // HTTP статус код ответа
	ResponseTime time.Duration // Время ответа
	Error        error         // Ошибка, если произошла
}

// Info форматирует результат пинга в читаемую строку
func (r Result) Info() string {
	if r.Error != nil {
		return fmt.Sprintf("[ERROR] - [%s] - %s", r.URL, r.Error.Error())
	}
	return fmt.Sprintf("[SUCCESS] - [%s] - Status: %d, Response Time: %s", r.URL, r.StatusCode, r.ResponseTime.String())
}

// Pool представляет пул воркеров
type Pool struct {
	worker       *worker         // Экземпляр воркера
	workersCount int             // Количество воркеров в пуле
	jobs         chan Job        // Канал для передачи задач
	results      chan Result     // Канал для получения результатов
	wg           *sync.WaitGroup // WaitGroup для синхронизации
	stopped      bool            // Флаг остановки пула
}

// Init запускает всех воркеров в пуле
func (p *Pool) Init() {
	for i := 0; i < p.workersCount; i++ {
		go p.initWorker(i)
	}
}

// Push добавляет новую задачу в пул
func (p *Pool) Push(j Job) {
	if p.stopped {
		return
	}
	p.jobs <- j
	p.wg.Add(1)
}

// Stop останавливает пул воркеров
func (p *Pool) Stop() {
	p.stopped = true
	close(p.jobs)
	p.wg.Wait()
	fmt.Println("ALL CHANNELS CLOSED")
}

// initWorker запускает отдельного воркера
func (p *Pool) initWorker(id int) {
	// Обрабатываем задачи, пока канал открыт
	for job := range p.jobs {
		time.Sleep(time.Second) // Искусственная задержка
		p.results <- p.worker.process(job)
		p.wg.Done()
	}

	log.Printf("[worker ID %d] finished proccesing", id)
}
