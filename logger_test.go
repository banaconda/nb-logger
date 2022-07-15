package nblogger

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"golang.org/x/exp/constraints"
)

var LogFilePath = "test.log"
var bufferSize = 100000

func init() {
	runtime.GOMAXPROCS(16)
}

func logging(logger Logger){
	logger.Trace("")
	logger.Debug("")
	logger.Info("")
	logger.Warn("")
	logger.Error("")
}

func TestLogLevelDefault(t *testing.T) {
	fmt.Printf("Info Level Test\n")
	logger, err := NewLogger(LogFilePath, Info, bufferSize, Lshortfile|LstdFlags|Lmicroseconds|Lstdout)
	if err != nil{
		t.Fatalf("%v", err)
	}
	defer logger.Close()

	logging(logger)

	os.Remove(LogFilePath)
}

func TestLogLevelTrace(t *testing.T) {
	fmt.Printf("Trace Level Test\n")
	logger, err := NewLogger(LogFilePath, Trace, bufferSize, Lshortfile|LstdFlags|Lmicroseconds|Lstdout)
	if err != nil{
		t.Fatalf("%v", err)
	}
	defer logger.Close()

	logging(logger)

	os.Remove(LogFilePath)
}

func TestLogLevelError(t *testing.T) {
	fmt.Printf("Error Level Test\n")
	logger, err := NewLogger(LogFilePath, Error, bufferSize, Lshortfile|LstdFlags|Lmicroseconds|Lstdout)
	if err != nil{
		t.Fatalf("%v", err)
	}
	defer logger.Close()

	logging(logger)

	os.Remove(LogFilePath)
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
	for i := 0; i < 1000; i++ {
		buf = append(buf, "#"...)
	}

	s := string(buf)

	start := time.Now()
	for i := 0; i < bufferSize; i++ {
		logger.Info(s)
	}

	return time.Since(start).Milliseconds()
}

func TestAsyncBenchmark(t *testing.T) {
	logger, err := NewLogger(LogFilePath, Info, bufferSize, Lshortfile|LstdFlags|Lmicroseconds)
	defer func() {
		logger.Close()
	}()

	if err != nil {
		t.Fatalf("%v", err)
	}

	os.Remove(LogFilePath)
	fmt.Printf("Async result: %d milis\n", benchmark(logger))
}

func TestSyncBenchmark(t *testing.T) {
	logger, err := NewLogger(LogFilePath, Info, bufferSize, Lshortfile|LstdFlags|Lmicroseconds|Lblocking)
	defer func() {
		logger.Close()
	}()

	if err != nil {
		t.Fatalf("%v", err)
	}

	os.Remove(LogFilePath)
	fmt.Printf("Sync result: %d milis\n", benchmark(logger))
}

func TestAsyncThreadedBenchmark(t *testing.T) {
	num := 20
	wg := sync.WaitGroup{}
	wg.Add(num)

	result := make([]int64, num)
	for i := 0; i < num; i++ {
		logger, err := NewLogger(fmt.Sprintf("%s.%d", LogFilePath, i), Info, bufferSize, LstdFlags|Lmicroseconds)
		defer func() {
			logger.Close()
		}()

		if err != nil {
			t.Fatalf("%v", err)
		}
		go func(index int) {
			result[index] = benchmark(logger)
			wg.Done()
		}(i)
	}
	wg.Wait()

	for i := 0; i < num; i++ {
		os.Remove(fmt.Sprintf("%s.%d", LogFilePath, i))
	}

	min, max, avg := minMaxAvg(result)
	fmt.Printf("AsyncThreaded min=%d, max=%d, avg=%d\n", min, max, avg)
}

func TestSyncThreadedBenchmark(t *testing.T) {
	num := 20
	wg := sync.WaitGroup{}
	wg.Add(num)

	result := make([]int64, num)
	for i := 0; i < num; i++ {
		logger, err := NewLogger(fmt.Sprintf("%s.%d", LogFilePath, i), Info, bufferSize, LstdFlags|Lmicroseconds|Lblocking)
		defer func() {
			logger.Close()
		}()

		if err != nil {
			t.Fatalf("%v", err)
		}
		go func(index int) {
			result[index] = benchmark(logger)
			wg.Done()
		}(i)
	}
	wg.Wait()

	for i := 0; i < num; i++ {
		os.Remove(fmt.Sprintf("%s.%d", LogFilePath, i))
	}

	min, max, avg := minMaxAvg(result)
	fmt.Printf("SyncThreaded min=%d, max=%d, avg=%d\n", min, max, avg)
}

func TestAsyncThreadedWithRuntimeBenchmark(t *testing.T) {
	num := 20
	wg := sync.WaitGroup{}
	wg.Add(num)

	result := make([]int64, num)
	for i := 0; i < num; i++ {
		logger, err := NewLogger(fmt.Sprintf("%s.%d", LogFilePath, i), Info, bufferSize, Lshortfile|LstdFlags|Lmicroseconds)
		defer func() {
			logger.Close()
		}()

		if err != nil {
			t.Fatalf("%v", err)
		}
		go func(index int) {
			result[index] = benchmark(logger)
			wg.Done()
		}(i)
	}
	wg.Wait()

	for i := 0; i < num; i++ {
		os.Remove(fmt.Sprintf("%s.%d", LogFilePath, i))
	}

	min, max, avg := minMaxAvg(result)
	fmt.Printf("AsyncThreadedWithRuntime min=%d, max=%d, avg=%d\n", min, max, avg)
}

func TestSyncThreadedWithRuntimeBenchmark(t *testing.T) {
	num := 20
	wg := sync.WaitGroup{}
	wg.Add(num)

	result := make([]int64, num)
	for i := 0; i < num; i++ {
		logger, err := NewLogger(fmt.Sprintf("%s.%d", LogFilePath, i), Info, bufferSize, Lshortfile|LstdFlags|Lmicroseconds|Lblocking)
		defer func() {
			logger.Close()
		}()

		if err != nil {
			t.Fatalf("%v", err)
		}
		go func(index int) {
			result[index] = benchmark(logger)
			wg.Done()
		}(i)
	}
	wg.Wait()

	for i := 0; i < num; i++ {
		os.Remove(fmt.Sprintf("%s.%d", LogFilePath, i))
	}

	min, max, avg := minMaxAvg(result)
	fmt.Printf("SyncThreadedWithRuntime min=%d, max=%d, avg=%d\n", min, max, avg)
}