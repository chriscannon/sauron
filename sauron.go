package main

import (
	"bufio"
	"net"
	"os"
	"runtime"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

type Result struct {
	Matches       int64
	ParseErrors   int64
	LookupErrors  int64
	NoStateErrors int64
	TotalLines    int64
}

func Run(geoIpFilePath string, state string, inputFile *os.File) (*Result, error) {
	numWorkers := runtime.NumCPU()
	runtime.GOMAXPROCS(numWorkers)

	db, err := geoip2.Open(geoIpFilePath)
	defer db.Close()

	if err != nil {
		return nil, err
	}

	results := makeWorkers(numWorkers, db, state, inputFile)

	result := Result{0, 0, 0, 0, 0}
	for r := range results {
		result.TotalLines += 1
		result.Matches += r.Matches
		result.ParseErrors += r.ParseErrors
		result.NoStateErrors += r.NoStateErrors
		result.LookupErrors += r.LookupErrors
	}

	return &result, nil
}

func makeWorkers(numWorkers int, db *geoip2.Reader, state string, file *os.File) <-chan Result {
	inputChan := make(chan string, numWorkers)
	outputChan := make(chan Result, numWorkers)
	wg := new(sync.WaitGroup)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			for ip := range inputChan {
				outputChan <- parseIp(ip, db, state)
			}
			wg.Done()
		}()
	}

	go func() {
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			inputChan <- scanner.Text()
		}

		close(inputChan)
		wg.Wait()
		close(outputChan)
	}()

	return outputChan
}

func parseIp(ipString string, db *geoip2.Reader, state string) Result {
	ip := net.ParseIP(ipString)
	if ip == nil {
		return Result{ParseErrors: 1}
	}

	record, err := db.City(ip)

	if err != nil {
		return Result{LookupErrors: 1}
	}

	if len(record.Subdivisions) == 0 {
		return Result{NoStateErrors: 1}
	}

	if record.Subdivisions[0].IsoCode == state {
		return Result{Matches: 1}
	}

	return Result{}
}
