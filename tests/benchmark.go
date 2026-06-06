package main

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
}

func NewClient(addr string) (*Client, error) {
	start := time.Now()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Connection established in %v\n", time.Since(start))

	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) SendCommand(args ...string) error {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "*%d\r\n", len(args))

	for _, arg := range args {
		fmt.Fprintf(&buf, "$%d\r\n%s\r\n", len(arg), arg)
	}

	_, err := c.conn.Write(buf.Bytes())
	if err != nil {
		return err
	}

	_, err = c.readResponse()
	return err
}

func (c *Client) readResponse() (interface{}, error) {
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimSpace(line)

	if len(line) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	switch line[0] {

	case '+':
		return line[1:], nil

	case '-':
		return nil, fmt.Errorf(line[1:])

	case ':':
		return strconv.Atoi(line[1:])

	case '$':
		n, _ := strconv.Atoi(line[1:])
		if n < 0 {
			return nil, nil
		}

		data := make([]byte, n)
		_, err := c.reader.Read(data)
		if err != nil {
			return nil, err
		}

		c.reader.Discard(2)
		return string(data), nil
	}

	return nil, fmt.Errorf("unknown response")
}

type Metrics struct {
	latencies []float64
	mu        sync.Mutex
}

func (m *Metrics) Add(d time.Duration) {
	m.mu.Lock()
	m.latencies = append(m.latencies, float64(d.Microseconds()))
	m.mu.Unlock()
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	idx := int((p / 100.0) * float64(len(sorted)-1))
	return sorted[idx]
}

func main() {

	addr := "localhost:6379"

	concurrency := 100
	opsPerWorker := 10000

	fmt.Printf("Target: %s\n", addr)
	fmt.Printf("Workers: %d\n", concurrency)
	fmt.Printf("Ops/worker: %d\n\n", opsPerWorker)

	var totalOps uint64
	var errors uint64

	metrics := &Metrics{}

	startBench := time.Now()

	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {

		wg.Add(1)

		go func(worker int) {
			defer wg.Done()

			client, err := NewClient(addr)
			if err != nil {
				atomic.AddUint64(&errors, 1)
				return
			}
			defer client.Close()

			for j := 0; j < opsPerWorker; j++ {

				key := fmt.Sprintf("key:%d:%d", worker, j)

				value := fmt.Sprintf("value-%d", rand.Int())

				start := time.Now()

				err := client.SendCommand("SET", key, value)

				latency := time.Since(start)
				metrics.Add(latency)

				if err != nil {
					atomic.AddUint64(&errors, 1)
				}

				atomic.AddUint64(&totalOps, 1)
			}

		}(i)
	}

	wg.Wait()

	elapsed := time.Since(startBench)

	sort.Float64s(metrics.latencies)

	total := atomic.LoadUint64(&totalOps)

	var avg float64

	for _, v := range metrics.latencies {
		avg += v
	}

	if len(metrics.latencies) > 0 {
		avg /= float64(len(metrics.latencies))
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	fmt.Println("========== RESULTS ==========")

	fmt.Printf("Operations:      %d\n", total)
	fmt.Printf("Errors:          %d\n", errors)
	fmt.Printf("Duration:        %v\n", elapsed)

	fmt.Printf("Ops/sec:         %.2f\n",
		float64(total)/elapsed.Seconds())

	fmt.Printf("\nLatency (µs)\n")
	fmt.Printf("Average:         %.2f\n", avg)
	fmt.Printf("Min:             %.2f\n", metrics.latencies[0])
	fmt.Printf("P50:             %.2f\n", percentile(metrics.latencies, 50))
	fmt.Printf("P90:             %.2f\n", percentile(metrics.latencies, 90))
	fmt.Printf("P95:             %.2f\n", percentile(metrics.latencies, 95))
	fmt.Printf("P99:             %.2f\n", percentile(metrics.latencies, 99))
	fmt.Printf("P99.9:           %.2f\n", percentile(metrics.latencies, 99.9))
	fmt.Printf("Max:             %.2f\n",
		metrics.latencies[len(metrics.latencies)-1])

	fmt.Printf("\nRuntime\n")
	fmt.Printf("Goroutines:      %d\n", runtime.NumGoroutine())
	fmt.Printf("Heap Alloc:      %.2f MB\n",
		float64(mem.Alloc)/1024/1024)
	fmt.Printf("Heap Sys:        %.2f MB\n",
		float64(mem.HeapSys)/1024/1024)
	fmt.Printf("GC Runs:         %d\n", mem.NumGC)

	fmt.Println("=============================")

	os.Exit(0)
}