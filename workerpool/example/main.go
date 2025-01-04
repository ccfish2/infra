package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/ccfish2/infra/workerpool"
)

func isPrime(n int64) bool {
	if n <= 2 {
		return false
	}
	for i := int64(2); i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func main() {
	wp := workerpool.New(runtime.NumCPU())
	for i, n := 0, int64(1_000_000_000_000_000_000); n < 1_000_000_000_000_000_100; i, n = i+1, n+1 {
		n := n
		id := fmt.Sprintf("task #%d", i)
		err := wp.Submit(id, func(_ context.Context) error {
			if isPrime(n) {
				fmt.Println(n, "isPrime !")
			}
			return nil
		})
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			return
		}
	}

	tasks, err := wp.Drain()
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		return
	}

	for _, t := range tasks {
		if t.Err() != nil {
			fmt.Fprintln(os.Stdout, t.String(), t.Err())
		}
	}
	if err := wp.Close(); err != nil {
		fmt.Fprintln(os.Stdout, err)
	}

}
