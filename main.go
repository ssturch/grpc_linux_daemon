package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

type Snapshot struct {
	info []SummaryInfo
}

func (snp *Snapshot) addToSnapshot(si SummaryInfo) {
	snp.info = append(snp.info, si)
}

type SummaryInfo struct {
	cpuQua                int
	loadAverage           *LoadAverageInfo
	cpuUsage              *CPUUsageInfo
	diskUsage             *DiskUsageInfoSummary
	dickCapac             *DiskCapacInfoSummary
	talkersByProt_1s      *TopTalkersByProtInfoSumm
	talkersByIPAverage_1s *TopTalkersByIPInfoSumm
	listeners             *ListenersInfoSummary
	states                *StatesInfoSummary
}

func AddInfo(s interface{}, info interface{}) {
	switch v := s.(type) {
	case *DiskUsageInfoSummary:
		v.summary = append(v.summary, info.(DiskUsageInfo))
	case *DiskCapacInfoSummary:
		v.summary = append(v.summary, info.(DiskCapacInfo))
	case *TopTalkersByProtInfoSumm:
		v.summary = append(v.summary, info.(TopTalkersByProtInfo))
	case *TopTalkersByIPInfoSumm:
		v.summary = append(v.summary, info.(TopTalkersByIPInfo))
	case *ListenersInfoSummary:
		v.summary = append(v.summary, info.(ListenersInfo))
	}
}

type LoadAverageInfo struct {
	loadAverage1  float64
	loadAverage5  float64
	loadAverage15 float64
}
type CPUUsageInfo struct {
	byUser   float64
	bySystem float64
	idle     float64
}

type DiskUsageInfo struct {
	name string
	tps  float64
	rps  float64
}
type DiskUsageInfoSummary struct {
	summary []DiskUsageInfo
}

type DiskCapacInfo struct {
	filesystem string
	mount      string
	use        int
	iuse       int
}
type DiskCapacInfoSummary struct {
	summary []DiskCapacInfo
}

type TopTalkersByProtInfo struct {
	prot string
	val  float64
}
type TopTalkersByProtInfoSumm struct {
	summary []TopTalkersByProtInfo
}

type TopTalkersByIPInfo struct {
	source      string
	destination string
	protocol    string
	bps         float64
}
type TopTalkersByIPInfoSumm struct {
	summary []TopTalkersByIPInfo
}

type ListenersInfo struct {
	command  string
	pid      string
	user     string
	protocol string
	port     string
}
type ListenersInfoSummary struct {
	summary []ListenersInfo
}

type StatesInfoSummary struct {
	info map[string]int
}

func main() {
	cpuQuaByte, _ := exec.Command("nproc").Output()
	cpuQuaInt, _ := strconv.Atoi(string(cpuQuaByte[0]))
	snp := new(Snapshot)
	siTemp := new(SummaryInfo)
	siTemp.cpuQua = cpuQuaInt

	ctx, cancel := context.WithCancel(context.Background())

	go worker(snp, siTemp, ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	exitChan := make(chan int)
	defer close(exitChan)
	go func() {
	L:
		for {
			select {
			case s := <-sigChan:
				switch s {
				case syscall.SIGINT:
					fmt.Println("Catch: SIGNAL INTERRUPT | Server stopped")
					exitChan <- 0
				case syscall.SIGTERM:
					fmt.Println("Catch: SIGNAL TERMINATE | Server stopped")
					exitChan <- 0
				case syscall.SIGKILL:
					fmt.Println("Catch: SIGNAL KILL | Server stopped")
					exitChan <- 0
				}
			case <-ctx.Done():
				break L
			}
		}
	}()

	exitCode := <-exitChan
	cancel()

	for _, val := range snp.info {
		fmt.Println(*val.loadAverage)
		fmt.Println(*val.cpuUsage)
		fmt.Println(*val.diskUsage)
		fmt.Println(*val.dickCapac)
		fmt.Println(*val.listeners)
		fmt.Println(*val.talkersByProt_1s)
		fmt.Println(*val.talkersByIPAverage_1s)
		fmt.Println(*val.states)
	}

	os.Exit(exitCode)
}
