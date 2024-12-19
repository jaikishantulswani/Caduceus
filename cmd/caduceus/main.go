package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/g0ldencybersec/Caduceus/pkg/scrape"
	"github.com/g0ldencybersec/Caduceus/pkg/types"
)

func main() {
	args := types.ScrapeArgs{}
	scrapeUsage := "-i <IPs/CIDRs or File> "

	flag.IntVar(&args.Concurrency, "c", 100, "How many goroutines running concurrently")
	flag.StringVar(&args.PortList, "p", "443", "TLS ports to check for certificates")
	flag.IntVar(&args.Timeout, "t", 4, "Timeout for TLS handshake")
	flag.StringVar(&args.Input, "i", "NONE", "Either IPs & CIDRs separated by commas, or a file with IPs/CIDRs on each line")
	flag.BoolVar(&args.Debug, "debug", false, "Add this flag if you want to see failures/timeouts")
	flag.BoolVar(&args.Help, "h", false, "Show the program usage message")
	flag.BoolVar(&args.JsonOutput, "j", false, "print cert data as jsonl")
	flag.BoolVar(&args.PrintWildcards, "wc", false, "print wildcards to stdout")
	continuousScan := flag.Bool("noEnd", false, "Continue scanning in cycles (default: exit after one scan)")

	flag.Parse()

	if args.Concurrency < 1 {
		args.Concurrency = 100
	}

	if args.Help {
		fmt.Println(scrapeUsage)
		flag.PrintDefaults()
		return
	}

	originalInput := args.Input
	cycleCount := 1
	
	for {
		fmt.Fprintf(os.Stderr, "\nStarting scan cycle #%d\n", cycleCount)

		args.Input = originalInput

		if args.Input == "NONE" {
			var stdinIPs []string
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := scanner.Text()
				if line != "" {
					stdinIPs = append(stdinIPs, line)
				}
			}

			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
				continue
			}

			args.Input = strings.Join(stdinIPs, ",")
		}

		args.Ports = strings.Split(args.PortList, ",")

		scrape.RunScrape(args)

		fmt.Fprintf(os.Stderr, "\nScan cycle #%d completed\n", cycleCount)

		if !*continuousScan {
			fmt.Fprintf(os.Stderr, "Scan completed. Exiting...\n")
			return
		}

		fmt.Fprintf(os.Stderr, "Waiting before starting next cycle...\n")
		time.Sleep(1 * time.Second)
		cycleCount++
	}
}
