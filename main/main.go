package main

import (
	"bufio"
	"fmt"
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
	scanner := bufio.NewScanner(os.Stdin)
	var data string
	for scanner.Scan() {
		data = scanner.Text()
		if strings.EqualFold(data, "exit") {
			fmt.Println("Программа завершила работу")
			close(done)
			return
		}
		number, err := strconv.Atoi(data)
		if err != nil {
			fmt.Println("Программа обрабатывает только целые числа!")
			continue
		}
		nextstep <- number
	}
}

func notnegativeFilter(prevChan <-chan int, nextChan chan<- int, done chan bool) {
	for {
		select {
		case data := <-prevChan:
			if data >= 0 {
				nextChan <- data
			}
		case <-done:
			return
		}
	}
}

func notDividedThreeFilter(prevChan <-chan int, nextChan chan<- int, done chan bool) {
	for {
		select {
		case data := <-prevChan:
			if data%3 == 0 {
				nextChan <- data
			}
		case <-done:
			return
		}
	}
}

func bufferFunc(prevChan <-chan int, nextChan chan<- int, done chan bool, size int, interval time.Duration) {
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
			return
		}
	}
}

func main() {
	done := make(chan bool)
	input := make(chan int)
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
			fmt.Println("Получены данные: ", data)
		case <-done:
			return
		}
	}
}
