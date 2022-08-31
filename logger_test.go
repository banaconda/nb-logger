package nblogger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"golang.org/x/exp/constraints"
)

var logFilePath = "test.log"
var bufferSize = 10000
var threadNum = 8

func init() {
	runtime.GOMAXPROCS(threadNum)
}

func logging(logger Logger) {
	logger.Trace("")
	logger.Debug("")
	logger.Info("")
	logger.Warn("")
	logger.Error("")
}

func TestLogLevelDefault(t *testing.T) {
	fmt.Printf("Info Level Test\n")
	logger, err := NewLogger(logFilePath, Info, bufferSize, Lshortfile|LstdFlags|Lmicroseconds|Lstdout)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer logger.Close()

	logging(logger)

	os.Remove(logFilePath)
}

func TestLogLevelTrace(t *testing.T) {
	fmt.Printf("Trace Level Test\n")
	logger, err := NewLogger(logFilePath, Trace, bufferSize, Lshortfile|LstdFlags|Lmicroseconds|Lstdout)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer logger.Close()

	logging(logger)

	os.Remove(logFilePath)
}

func TestLogLevelError(t *testing.T) {
	fmt.Printf("Error Level Test\n")
	logger, err := NewLogger(logFilePath, Error, bufferSize, Lshortfile|LstdFlags|Lmicroseconds|Lstdout)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer logger.Close()

	logging(logger)

	os.Remove(logFilePath)
}

func minMaxAvg[T constraints.Integer](array []T) (T, T, T) {
	var max T = array[0]
	var min T = array[0]
	var avg T = 0
	for _, value := range array {
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}

		avg += value
	}

	avg /= T(len(array))
	return min, max, avg
}

func benchmark(logger Logger) int64 {
	var buf []byte
	var list []any
	for i := 0; i < 3000; i++ {
		buf = append(buf, "%d"...)
		list = append(list, i)
	}

	s := string(buf)

	start := time.Now()
	for i := 0; i < bufferSize; i++ {
		logger.Info(s, list...)
	}

	return time.Since(start).Milliseconds()
}

func TestAsyncBenchmark(t *testing.T) {
	logger, err := NewLogger(logFilePath, Info, bufferSize, Lshortfile|LstdFlags|Lmicroseconds)
	defer func() {
		logger.Close()
	}()

	if err != nil {
		t.Fatalf("%v", err)
	}

	os.Remove(logFilePath)
	result := benchmark(logger)
	logger.Close()
	log.Printf("Async result: %d milis\n", result)
}

func TestSyncBenchmark(t *testing.T) {
	logger, err := NewLogger(logFilePath, Info, bufferSize, Lshortfile|LstdFlags|Lmicroseconds|Lblocking)
	defer func() {
		logger.Close()
	}()

	if err != nil {
		t.Fatalf("%v", err)
	}

	os.Remove(logFilePath)
	result := benchmark(logger)
	logger.Close()
	log.Printf("Sync result: %d milis\n", result)
}

func TestAsyncThreadedBenchmark(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(threadNum)

	result := make([]int64, threadNum)
	for i := 0; i < threadNum; i++ {
		logger, err := NewLogger(fmt.Sprintf("%s.%d", logFilePath, i), Info, bufferSize, LstdFlags|Lmicroseconds)
		if err != nil {
			t.Fatalf("%v", err)
		}

		go func(index int) {
			result[index] = benchmark(logger)
			logger.Close()
			wg.Done()
		}(i)
	}
	wg.Wait()

	for i := 0; i < threadNum; i++ {
		os.Remove(fmt.Sprintf("%s.%d", logFilePath, i))
	}

	min, max, avg := minMaxAvg(result)
	log.Printf("AsyncThreaded min=%d, max=%d, avg=%d\n", min, max, avg)
}

func TestSyncThreadedBenchmark(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(threadNum)

	result := make([]int64, threadNum)
	for i := 0; i < threadNum; i++ {
		logger, err := NewLogger(fmt.Sprintf("%s.%d", logFilePath, i), Info, bufferSize, LstdFlags|Lmicroseconds|Lblocking)
		if err != nil {
			t.Fatalf("%v", err)
		}

		go func(index int) {
			result[index] = benchmark(logger)
			logger.Close()
			wg.Done()
		}(i)
	}
	wg.Wait()

	for i := 0; i < threadNum; i++ {
		os.Remove(fmt.Sprintf("%s.%d", logFilePath, i))
	}

	min, max, avg := minMaxAvg(result)
	log.Printf("SyncThreaded min=%d, max=%d, avg=%d\n", min, max, avg)
}

func TestAsyncThreadedWithRuntimeBenchmark(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(threadNum)

	result := make([]int64, threadNum)
	for i := 0; i < threadNum; i++ {
		logger, err := NewLogger(fmt.Sprintf("%s.%d", logFilePath, i), Info, bufferSize, Lshortfile|LstdFlags|Lmicroseconds)
		if err != nil {
			t.Fatalf("%v", err)
		}

		go func(index int) {
			result[index] = benchmark(logger)
			logger.Close()
			wg.Done()
		}(i)
	}
	wg.Wait()

	for i := 0; i < threadNum; i++ {
		os.Remove(fmt.Sprintf("%s.%d", logFilePath, i))
	}

	min, max, avg := minMaxAvg(result)
	log.Printf("AsyncThreadedWithRuntime min=%d, max=%d, avg=%d\n", min, max, avg)
}

func TestSyncThreadedWithRuntimeBenchmark(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(threadNum)

	result := make([]int64, threadNum)
	for i := 0; i < threadNum; i++ {
		logger, err := NewLogger(fmt.Sprintf("%s.%d", logFilePath, i), Info, bufferSize, Lshortfile|LstdFlags|Lmicroseconds|Lblocking)
		if err != nil {
			t.Fatalf("%v", err)
		}

		go func(index int) {
			result[index] = benchmark(logger)
			logger.Close()
			wg.Done()
		}(i)
	}
	wg.Wait()

	for i := 0; i < threadNum; i++ {
		os.Remove(fmt.Sprintf("%s.%d", logFilePath, i))
	}

	min, max, avg := minMaxAvg(result)
	log.Printf("SyncThreadedWithRuntime min=%d, max=%d, avg=%d\n", min, max, avg)
}
