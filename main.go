package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"log"

	"github.com/dustin/go-humanize"
	"runtime/pprof"
)

var (
	flState     = flag.String("state", "", "Set the state (a.k.a. subdivision) to filter on")
	flGeoIpFile = flag.String("geoip", "", "Set the path to the GeoIP2 City file")
	flHelp      = flag.Bool("help", false, "Print usage")
	flInputFile = flag.String("input", "", "Set the input file of IP addresses")
	flCpuProfile = flag.Bool("cpuprof", false, "Write the CPU profile to a file")
)

func init() {
	flag.Usage = func() {
		fmt.Printf("Usage: sauron [OPTIONS]\n\nA concurrent utility to resolve IPs to states.\n\nOptions:\n")
		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if *flHelp {
		flag.Usage()
		return
	}

	if *flState == "" {
		fmt.Println("Please specify a state to filter on.")
		flag.Usage()
		os.Exit(1)
	}
	state := strings.ToUpper(strings.TrimSpace(*flState))

	if *flGeoIpFile == "" {
		fmt.Println("Please specify a GeoIP2 City file.")
		flag.Usage()
		os.Exit(1)
	}

	inputFile := os.Stdin
	if *flInputFile != "" {
		file, err := os.Open(*flInputFile)
		if err != nil {
			log.Fatal(err)
		}
		inputFile = file
	}
	defer inputFile.Close()

	if *flCpuProfile {
		f, err := os.Create("sauron.prof")
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
	}

	result, err := Run(*flGeoIpFile, state, inputFile)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("# of Lines: %s\n", humanize.Comma(result.TotalLines))
	fmt.Printf("# of IPs from %s: %s\n", state, humanize.Comma(result.Matches))
	fmt.Printf("# of Unparseable IPs: %s\n", humanize.Comma(result.ParseErrors))
	fmt.Printf("# of GeoIP2 Lookup Errors: %s\n", humanize.Comma(result.LookupErrors))
	fmt.Printf("# of GeoIP2 No State Found Errors: %s\n", humanize.Comma(result.NoStateErrors))
}
