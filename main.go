package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/dustin/go-humanize"
)

var (
	flState      = flag.String("state", "", "Set the state (a.k.a. subdivision) to filter on")
	flCountry    = flag.String("country", "US", "Set the country to filter on")
	flGeoIpFile  = flag.String("geoip", "", "Set the path to the GeoIP2 City file")
	flHelp       = flag.Bool("help", false, "Print usage")
	flInputFile  = flag.String("input", "", "Set the input file of IP addresses")
	flCpuProfile = flag.Bool("cpuprof", false, "Write the CPU profile to a file")
)

func init() {
	flag.Usage = func() {
		fmt.Printf("Usage: sauron [OPTIONS]\n\nA concurrent utility to resolve IPs to countries and states.\n\nOptions:\n")
		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()
		fmt.Printf("\nAs input sauron expects a list of IPv4 IP addresses each on its own line. E.g.,\n10.0.0.1\n10.0.0.2\n10.0.0.3\n")
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

	fmt.Println("Processing...")

	result, err := Run(*flGeoIpFile, *flState, *flCountry, inputFile)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("# of Lines: %s\n", humanize.Comma(result.TotalLines))
	fmt.Printf("# of IPs from %s: %s\n", CleanIso(*flCountry), humanize.Comma(result.CountryMatches))
	fmt.Printf("# of IPs from %s: %s\n", CleanIso(*flState), humanize.Comma(result.StateMatches))

	if result.ParseErrors > 0 {
		fmt.Printf("# of Unparseable IPs: %s\n", humanize.Comma(result.ParseErrors))
	}

	if result.LookupErrors > 0 {
		fmt.Printf("# of GeoIP2 Lookup Errors: %s\n", humanize.Comma(result.LookupErrors))
	}

	if result.NoStateErrors > 0 {
		fmt.Printf("# of GeoIP2 No State Found Errors: %s\n", humanize.Comma(result.NoStateErrors))
	}
}
