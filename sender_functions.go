package main

import (
	"context"
	"time"
)

func worker(snp *Snapshot, sit *SummaryInfo, ctx context.Context) {
	heartbeat := time.Tick(1 * time.Second)
L:
	for {
		select {
		case <-heartbeat:
			done := make(chan any)
			go getCPUUsage(sit, done)
			go getLoadAverage(sit, done)
			go getDisksUsage(sit, done)
			go getDisksCapacity(sit, done)
			go getTopTalkersByProtocol(sit, done)
			go getTopTalkersByIP(sit, done)
			go getTCPnUDPListeners(sit, done)
			go getTCPStates(sit, done)
			for i := 0; i < 8; i++ {
				<-done
			}
			snp.addToSnapshot(*sit)
		case <-ctx.Done():
			break L
		}
	}
}
