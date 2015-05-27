package main

import (
	"bufio"
	"net"
	"runtime"

	"os"

	"github.com/oschwald/geoip2-golang"
	"sync"
)

type Result struct {
	Matches       int64
	ParseErrors   int64
	LookupErrors  int64
	NoStateErrors int64
	TotalLines    int64
}

func Run(geoIpFilePath string, state string, inputFile *os.File) (*Result, error) {
	workers := runtime.NumCPU()
	runtime.GOMAXPROCS(workers)

	ips := make(chan string)
	results := make(chan Result, workers)
	closer := make(chan struct{})

	db, err := geoip2.Open(geoIpFilePath)
	defer db.Close()

	if err != nil {
		return nil, err
	}

	wg := new(sync.WaitGroup)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go lookupIps(state, db, ips, results, closer, wg)
	}

	go readIps(ips, results, closer, inputFile, wg)

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

func lookupIps(state string, db *geoip2.Reader, ipInput chan string, resultsOutput chan Result, closer chan struct{}, wg *sync.WaitGroup) {
	RUNLOOP:
	for {
		select {
		case ipString := <-ipInput:
			resultsOutput <- getIpResult(ipString, db, state)
		case <-closer:
			break RUNLOOP
		}
	}
	wg.Done()
}

func getIpResult(ipString string, db *geoip2.Reader, state string) (Result) {
	ip := net.ParseIP(ipString)
	if ip == nil {
		return Result{ParseErrors:1}
	}

	record, err := db.City(ip)

	if err != nil {
		return Result{LookupErrors:1}
	}

	if len(record.Subdivisions) == 0 {
		return Result{NoStateErrors:1}
	}

	if record.Subdivisions[0].IsoCode == state {
		return Result{Matches:1}
	}

	return Result{}
}

func readIps(input chan string, output chan Result, closer chan struct{}, file *os.File, wg *sync.WaitGroup) {
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		input <- scanner.Text()
	}
	close(closer)
	wg.Wait()
	close(output)
}
