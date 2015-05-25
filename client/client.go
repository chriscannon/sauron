package client

import (
	"bufio"
	"net"
	"runtime"

	"os"

	"github.com/oschwald/geoip2-golang"
)

type Result struct {
	Matches       int64
	ParseErrors   int64
	LookupErrors  int64
	NoStateErrors int64
	TotalLines    int64
}

const ipChannelBufferSize = 10000

func Run(geoIpFilePath string, state string, inputFile *os.File) (*Result, error) {
	workers := runtime.NumCPU()
	runtime.GOMAXPROCS(workers)

	ips := make(chan string, ipChannelBufferSize)
	results := make(chan Result, workers)

	db, err := geoip2.Open(geoIpFilePath)
	defer db.Close()

	if err != nil {
		return nil, err
	}

	go readIps(ips, inputFile)

	for i := 0; i < workers; i++ {
		go lookupIps(state, db, ips, results)
	}

	result := Result{0, 0, 0, 0, 0}
	for i := 0; i < workers; i++ {
		workerResult := <-results
		result.TotalLines += workerResult.TotalLines
		result.Matches += workerResult.Matches
		result.ParseErrors += workerResult.ParseErrors
		result.NoStateErrors += workerResult.NoStateErrors
		result.LookupErrors += workerResult.LookupErrors
	}

	return &result, nil
}

func lookupIps(state string, db *geoip2.Reader, ips chan string, results chan Result) {
	result := Result{0, 0, 0, 0, 0}
	for ipString := range ips {
		result.TotalLines++

		ip := net.ParseIP(ipString)
		if ip == nil {
			result.ParseErrors++
			continue
		}

		record, err := db.City(ip)

		if err != nil {
			result.LookupErrors++
		}

		if len(record.Subdivisions) == 0 {
			result.NoStateErrors++
			continue
		}

		if record.Subdivisions[0].IsoCode == state {
			result.Matches++
		}
	}
	results <- result
}

func readIps(ips chan string, file *os.File) {
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		ips <- scanner.Text()
	}

	close(ips)
}
