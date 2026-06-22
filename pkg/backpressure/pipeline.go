package backpressure

import (
	"context"
	"sync"
)

// Stage представляет один шаг обработки элемента.
// Если возвращается ошибка, элемент не передаётся дальше по конвейеру.
type Stage[T any] func(ctx context.Context, item T) error

// Pipeline обрабатывает поток элементов через цепочку стадий
// с ограниченным параллелизмом и обратным давлением.
type Pipeline[T any] struct {
	ctx     context.Context
	input   <-chan T
	stages  []Stage[T]
	workers []int // количество горутин для каждой стадии
	bufSize int

	out  chan T
	done chan struct{}
}

// Option настраивает Pipeline.
type Option[T any] func(*Pipeline[T])

// WithWorkers задаёт количество горутин для каждой стадии.
// Длина должна совпадать с количеством стадий. По умолчанию по 1 на стадию.
func WithWorkers[T any](workers ...int) Option[T] {
	return func(p *Pipeline[T]) {
		p.workers = workers
	}
}

// WithBufferSize устанавливает размер буфера каналов между стадиями.
// По умолчанию 64.
func WithBufferSize[T any](size int) Option[T] {
	return func(p *Pipeline[T]) {
		p.bufSize = size
	}
}

// NewPipeline создаёт новый конвейер. Он не запускается до вызова Start.
func NewPipeline[T any](ctx context.Context, input <-chan T, stages []Stage[T], opts ...Option[T]) *Pipeline[T] {
	p := &Pipeline[T]{
		ctx:     ctx,
		input:   input,
		stages:  stages,
		bufSize: 64,
	}
	for _, opt := range opts {
		opt(p)
	}
	if p.workers == nil {
		p.workers = make([]int, len(stages))
		for i := range p.workers {
			p.workers[i] = 1
		}
	}
	return p
}

// Start запускает конвейер и возвращает канал, из которого можно читать
// успешно обработанные элементы. Канал закрывается после завершения всех стадий.
func (p *Pipeline[T]) Start() <-chan T {
	p.out = make(chan T, p.bufSize)
	p.done = make(chan struct{})
	go func() {
		p.run()
		close(p.out)
		close(p.done)
	}()
	return p.out
}

// Wait блокируется до полной остановки конвейера.
func (p *Pipeline[T]) Wait() {
	<-p.done
}

// run выстраивает цепочку горутин и ожидает их завершения.
func (p *Pipeline[T]) run() {
	if len(p.stages) == 0 {
		// Сквозная передача без обработки
		for item := range p.input {
			select {
			case p.out <- item:
			case <-p.ctx.Done():
				return
			}
		}
		return
	}

	// Создаём каналы между стадиями
	chans := make([]chan T, len(p.stages)+1)
	for i := range chans {
		chans[i] = make(chan T, p.bufSize)
	}

	var wg sync.WaitGroup

	// Фидер: читает входной поток и передаёт в первый канал
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(chans[0])
		for item := range p.input {
			select {
			case chans[0] <- item:
			case <-p.ctx.Done():
				return
			}
		}
	}()

	// Для каждой стадии запускаем пул воркеров и замыкающий горутину,
	// которая закрывает выходной канал стадии после завершения всех воркеров.
	for i, stage := range p.stages {
		in := chans[i]
		out := chans[i+1]
		num := p.workers[i]

		var stageWg sync.WaitGroup
		for j := 0; j < num; j++ {
			stageWg.Add(1)
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer stageWg.Done()
				for {
					select {
					case item, ok := <-in:
						if !ok {
							return
						}
						if err := stage(p.ctx, item); err != nil {
							continue
						}
						select {
						case out <- item:
						case <-p.ctx.Done():
							return
						}
					case <-p.ctx.Done():
						return
					}
				}
			}()
		}

		// Горутина, закрывающая выходной канал после завершения всех воркеров стадии
		wg.Add(1)
		go func() {
			defer wg.Done()
			stageWg.Wait()
			close(out)
		}()
	}

	// Коллектор: читает из последнего канала и отправляет в p.out
	wg.Add(1)
	go func() {
		defer wg.Done()
		for item := range chans[len(p.stages)] {
			select {
			case p.out <- item:
			case <-p.ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
}
