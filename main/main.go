package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RingBufferInt struct {
	arrayint []int
	pos      int
	size     int
	mu       sync.Mutex
}

func NewRingBufferInt(size int) *RingBufferInt {
	return &RingBufferInt{make([]int, size), -1, size, sync.Mutex{}}
}

func (rbi *RingBufferInt) Push(element int) {
	rbi.mu.Lock()
	defer rbi.mu.Unlock()

	if rbi.pos == rbi.size-1 {
		for i := 1; i <= rbi.size-1; i++ {
			rbi.arrayint[i-1] = rbi.arrayint[i]
		}
		rbi.arrayint[rbi.pos] = element
	} else {
		rbi.pos++
		rbi.arrayint[rbi.pos] = element
	}

}

func (rbi *RingBufferInt) Get() []int {
	if rbi.pos < 0 {
		return nil
	}
	rbi.mu.Lock()
	defer rbi.mu.Unlock()
	var output []int = rbi.arrayint[:rbi.pos+1]
	rbi.pos = -1
	return output
}

func read(nextstep chan int, done chan bool) {

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	child := logger.With(
		slog.String("methd", "read"),
	)

	child.Info("Запущен метод считывания....")
	scanner := bufio.NewScanner(os.Stdin)
	var data string
	for scanner.Scan() {
		data = scanner.Text()
		if strings.EqualFold(data, "exit") {
			child.Info("Программа завершила работу")
			fmt.Println("Программа завершила работу")
			close(done)
			return
		}
		number, err := strconv.Atoi(data)
		if err != nil {
			child.Warn("Введены некорректные данные")
			fmt.Println("Программа обрабатывает только целые числа!")
			continue
		}
		nextstep <- number
	}
}

func notnegativeFilter(prevChan <-chan int, nextChan chan<- int, done chan bool) {
	slog.Info("Запущен фильтр отрицательных чисел, 1-й уровень", slog.String("methd", "notnegativeFilter"))
	for {
		select {
		case data := <-prevChan:
			if data >= 0 {
				nextChan <- data
			}
		case <-done:
			slog.Info("Завершение 1-го уровня", slog.String("methd", "notnegativeFilter"))
			return
		}
	}
}

func notDividedThreeFilter(prevChan <-chan int, nextChan chan<- int, done chan bool) {
	slog.Info("Запущен фильтр чисел кратных трём, 2-й уровень", slog.String("methd", "notDividedThreeFilte"))
	for {
		select {
		case data := <-prevChan:
			if data%3 == 0 {
				nextChan <- data
			}
		case <-done:
			slog.Info("Завершение 2-го уровня", slog.String("methd", "notDividedThreeFilte"))
			return
		}
	}
}

func bufferFunc(prevChan <-chan int, nextChan chan<- int, done chan bool, size int, interval time.Duration) {
	slog.Info("Запущен уровень буферизации, 3-й уровень", slog.String("methd", "bufferFunc"))
	buffer := NewRingBufferInt(size)

	for {
		select {
		case data := <-prevChan:
			buffer.Push(data)
		case <-time.After(interval):
			bufferdata := buffer.Get()
			for _, data := range bufferdata {
				nextChan <- data
			}

		case <-done:
			slog.Info("Завершение 3-го уровня", slog.String("methd", "bufferFunc"))
			return
		}
	}
}

func main() {
	done := make(chan bool)
	input := make(chan int)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	child := logger.With(
		slog.String("method", "main"),
	)

	child.Info("Pipeline запущен....")

	go read(input, done)

	notnegativeChannel := make(chan int)
	go notnegativeFilter(input, notnegativeChannel, done)

	notDividedThreeChannel := make(chan int)
	go notDividedThreeFilter(notnegativeChannel, notDividedThreeChannel, done)

	bufferChannel := make(chan int)
	bufferSize := 5
	timeIntervalClear := 3 * time.Second
	go bufferFunc(notDividedThreeChannel, bufferChannel, done, bufferSize, timeIntervalClear)

	for {
		select {
		case data := <-bufferChannel:
			child.Info("Получены данные:", slog.Int("data", data))
			fmt.Println("Получены данные: ", data)
		case <-done:
			child.Info("Выход из функции main")
			return
		}
	}
}
