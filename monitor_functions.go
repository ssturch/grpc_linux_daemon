package main

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func getLoadAverage(sI *SummaryInfo, done chan<- any) {
	la := new(LoadAverageInfo)
	laByte, _ := exec.Command("uptime").Output()
	laStr := string(laByte)
	laSlice := strings.Fields(laStr)
	for i, _ := range laSlice {
		laSlice[i] = strings.TrimSuffix(laSlice[i], ",")
		laSlice[i] = strings.Replace(laSlice[i], ",", ".", -1)
	}
	La1, _ := strconv.ParseFloat(laSlice[len(laSlice)-3], 64)
	La5, _ := strconv.ParseFloat(laSlice[len(laSlice)-2], 64)
	La15, _ := strconv.ParseFloat(laSlice[len(laSlice)-1], 64)
	la.loadAverage1 = La1
	la.loadAverage5 = La5
	la.loadAverage15 = La15
	sI.loadAverage = la
	done <- struct{}{}
}
func getCPUUsage(sI *SummaryInfo, done chan<- any) {
	ca := new(CPUUsageInfo)
	cpuAByte, _ := exec.Command("bash", "-c", "top -b -n1 | grep '%Cpu'").Output()
	cpuAStr := string(cpuAByte)
	cpuAStr = strings.Replace(cpuAStr, ",", ".", -1)
	cpuASlice := strings.Fields(cpuAStr)
	usParam, _ := strconv.ParseFloat(cpuASlice[1], 64)
	syParam, _ := strconv.ParseFloat(cpuASlice[3], 64)
	niParam, _ := strconv.ParseFloat(cpuASlice[5], 64)
	hiParam, _ := strconv.ParseFloat(cpuASlice[11], 64)
	siParam, _ := strconv.ParseFloat(cpuASlice[13], 64)

	userCPUUsage := usParam + niParam
	systemCPUUsage := syParam + hiParam + siParam
	idleCPU := 100 - userCPUUsage - systemCPUUsage

	ca.byUser = userCPUUsage
	ca.bySystem = systemCPUUsage
	ca.idle = idleCPU

	sI.cpuUsage = ca
	done <- struct{}{}
}
func getDisksUsage(sI *SummaryInfo, done chan<- any) {
	duis := new(DiskUsageInfoSummary)
	du := new(DiskUsageInfo)
	disksUsageByte, err := exec.Command("bash", "-c", "iostat -d -k | sed '1,3d' | awk '{print $1,$2,$3}' | tr '/n' ' '").Output()
	if err != nil {
		fmt.Println(err)
		return
	}
	disksUsageStr := string(disksUsageByte)
	disksUsageStr = strings.Replace(disksUsageStr, ",", ".", -1)
	disksUsageSlice := strings.Fields(disksUsageStr)
	for i := 0; i < len(disksUsageSlice); i += 3 {
		tpsTemp, _ := strconv.ParseFloat(disksUsageSlice[i+1], 64)
		rpsTemp, _ := strconv.ParseFloat(disksUsageSlice[i+2], 64)
		du.name = disksUsageSlice[i]
		du.tps = tpsTemp
		du.rps = rpsTemp
		AddInfo(duis, *du)
	}
	sI.diskUsage = duis
	done <- struct{}{}
}
func getDisksCapacity(sI *SummaryInfo, done chan<- any) {
	dcis := new(DiskCapacInfoSummary)
	dc := new(DiskCapacInfo)
	disksCapacityByte, _ := exec.Command("bash", "-c", "df -k | sed '1d' | awk '{print $1,$6,$5}' | tr '\n' ' '").Output()
	disksCapacityInodeByte, _ := exec.Command("bash", "-c", "df -i | sed '1d' | awk '{print $1,$6,$5}' | tr '\n' ' '").Output()
	disksCapacityStr := string(disksCapacityByte)
	disksCapacityInodeStr := string(disksCapacityInodeByte)
	disksCapacitySlice := strings.Fields(disksCapacityStr)
	disksCapacityInodeSlice := strings.Fields(disksCapacityInodeStr)
	for i := 0; i < len(disksCapacitySlice); i += 3 {
		useStr := strings.Replace(disksCapacitySlice[i+2], "%", "", -1)
		iuseStr := strings.Replace(disksCapacityInodeSlice[i+2], "%", "", -1)
		useTemp, _ := strconv.Atoi(useStr)
		iuseTemp, _ := strconv.Atoi(iuseStr)
		dc.filesystem = disksCapacitySlice[i]
		dc.mount = disksCapacitySlice[i+1]
		dc.use = useTemp
		dc.iuse = iuseTemp
		AddInfo(dcis, *dc)
	}
	sI.dickCapac = dcis
	done <- struct{}{}
}
func getTopTalkersByProtocol(sI *SummaryInfo, done chan<- any) {
	ttpis := new(TopTalkersByProtInfoSumm)
	err := exec.Command("bash", "-c", "sudo iptables -F").Run()
	if err != nil {
		fmt.Println(err)
	}
	err = exec.Command("bash", "-c", "sudo iptables -A OUTPUT -p tcp").Run()
	if err != nil {
		fmt.Println(err)
	}
	err = exec.Command("bash", "-c", "sudo iptables -A OUTPUT -p udp").Run()
	if err != nil {
		fmt.Println(err)
	}
	err = exec.Command("bash", "-c", "sudo iptables -A OUTPUT -p icmp").Run()
	if err != nil {
		fmt.Println(err)
	}
	err = exec.Command("bash", "-c", "sudo iptables -A OUTPUT -p all").Run()
	if err != nil {
		fmt.Println(err)
	}
	time.Sleep(1 * time.Second)

	cmd := exec.Command("bash", "-c", "sudo iptables -nvxL OUTPUT | sed '1,2d' | awk '{print $2,$3}' | tr '\n' ' '")

	portTalkersByte, _ := cmd.Output()
	portTalkersStr := string(portTalkersByte)
	portTalkersSlice := strings.Fields(portTalkersStr)

	tcpByteQuantity, _ := strconv.ParseFloat(portTalkersSlice[0], 64)
	udpByteQuantity, _ := strconv.ParseFloat(portTalkersSlice[2], 64)
	icmpByteQuantity, _ := strconv.ParseFloat(portTalkersSlice[4], 64)
	allByteQuantity, _ := strconv.ParseFloat(portTalkersSlice[6], 64)
	tempval := 100 / allByteQuantity
	tcpRes := tempval * tcpByteQuantity
	udpRes := tempval * udpByteQuantity
	icmpRes := tempval * icmpByteQuantity
	AddInfo(ttpis, TopTalkersByProtInfo{prot: "TCP", val: tcpRes})
	AddInfo(ttpis, TopTalkersByProtInfo{prot: "UDP", val: udpRes})
	AddInfo(ttpis, TopTalkersByProtInfo{prot: "ICMP", val: icmpRes})
	sI.talkersByProt_1s = ttpis
	done <- struct{}{}
}
func getTopTalkersByIP(sI *SummaryInfo, done chan<- any) {
	//CHECK IT!!!!
	ttipis := new(TopTalkersByIPInfoSumm)
	ttipi := new(TopTalkersByIPInfo)
	tempCommand := fmt.Sprintf("tcpdump -ntq -i any & sleep %v; kill $!", 1)
	cmd := exec.Command("bash", "-c", tempCommand)
	stdout, err := cmd.StdoutPipe()
	buf := new(strings.Builder)

	if err != nil {
		log.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(buf, stdout)

	if err != nil {
		fmt.Println("err buf")
		fmt.Println(err)
	}
	if err = cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	ipTalkersStr := buf.String()
	ipTalkersSlice := strings.Split(ipTalkersStr, "\n")

	for _, str := range ipTalkersSlice {
		strSliced := strings.Fields(str)
		if len(strSliced) == 0 {
			continue
		}
		if strSliced[1] == "In" {
			continue
		}

		ttipi.source = strSliced[3]
		ttipi.destination = strSliced[5]
		ttipi.protocol = strSliced[6]

		bites, _ := strconv.ParseFloat(strSliced[len(strSliced)-1], 64)
		ttipi.bps += bites
		AddInfo(ttipis, *ttipi)
	}
	sI.talkersByIPAverage_1s = ttipis
	done <- struct{}{}
}
func getTCPnUDPListeners(sI *SummaryInfo, done chan<- any) {
	lis := new(ListenersInfoSummary)
	li := new(ListenersInfo)
	cmd := exec.Command("bash", "-c", "sudo netstat -ltupe --numeric-hosts --numeric-ports | sed '1,2d' | grep 'LISTEN'")
	TCPnUDPListenerByte, _ := cmd.Output()
	TCPnUDPListenerStr := string(TCPnUDPListenerByte)

	TCPnUDPListenerSlice := strings.Split(TCPnUDPListenerStr, "\n")

	for _, str := range TCPnUDPListenerSlice {
		str = strings.Replace(str, ": ", ":", -1)
		tempSlice := strings.Fields(str)
		if len(tempSlice) == 0 {
			continue
		}
		pidNcomm := strings.Split(tempSlice[8], "/")
		li.command = pidNcomm[1]
		li.pid = pidNcomm[0]
		li.user = tempSlice[6]
		li.protocol = tempSlice[0]
		li.port = tempSlice[3]
		AddInfo(lis, *li)
	}
	sI.listeners = lis
	done <- struct{}{}
}
func getTCPStates(sI *SummaryInfo, done chan<- any) {
	tcpss := new(StatesInfoSummary)
	cmd := exec.Command("bash", "-c", "sudo ss -ta | awk '{print $1}' | sed '1d'| tr '\n' ' '")
	TCPStatesByte, _ := cmd.Output()
	TCPStatesStr := string(TCPStatesByte)

	valueMap := make(map[string]int)
	TCPStatesSlice := strings.Fields(TCPStatesStr)

	for _, val := range TCPStatesSlice {
		valueMap[val]++
	}
	tcpss.info = valueMap
	sI.states = tcpss
	done <- struct{}{}

}
