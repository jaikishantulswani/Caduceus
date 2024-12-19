package scrape

import (
	"fmt"
	"net"

	"time"

	"github.com/g0ldencybersec/Caduceus/pkg/types"
	"github.com/g0ldencybersec/Caduceus/pkg/utils"
	"github.com/g0ldencybersec/Caduceus/pkg/workers"
)

func RunScrape(args types.ScrapeArgs) {
    dialer := &net.Dialer{
        Timeout: time.Duration(args.Timeout) * time.Second,
    }

    inputChannel := make(chan string)
    resultChannel := make(chan types.Result)
    outputChannel := make(chan string, args.Concurrency/10)
    retryChannel := make(chan string, args.Concurrency)

    // Create and start the worker pool
    workerPool := workers.NewWorkerPool(args.Concurrency, dialer, inputChannel, resultChannel)
    workerPool.Start()

    // Create and start the results worker pool
    resultsWorkerPool := workers.NewResultWorkerPool(max(1, args.Concurrency/10), resultChannel, outputChannel)
    resultsWorkerPool.Start(args)

    // Handle input feeding
    go func() {
        utils.IntakeFunction(inputChannel, args.Ports, args.Input)
        close(inputChannel)
    }()

    // Handle outputs and retries
    go func() {
        retryCount := make(map[string]int)
        maxRetries := 1

        for output := range outputChannel {
            fmt.Println(output)
        }

        for result := range resultChannel {
            if result.Retry && retryCount[result.IP] < maxRetries {
                retryCount[result.IP]++
                retryChannel <- result.IP
            } else if result.Error != nil && args.Debug {
                fmt.Printf("Persistent error for IP %s: %v\n", result.IP, result.Error)
            }
        }

        close(retryChannel)
    }()

    // Retry mechanism
    go func() {
        for ip := range retryChannel {
            inputChannel <- ip
        }
    }()

    workerPool.Stop()
    resultsWorkerPool.Stop()
}
