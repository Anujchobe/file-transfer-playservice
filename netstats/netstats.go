// Package netstats logs the kernel's UDP counters so experiments can see datagrams the
// kernel drops, which the application never sees (for example a full socket receive buffer).
package netstats

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// Log prints the Udp counters from /proc/net/snmp to stderr every interval, tagged with name.
// It returns at once where /proc does not exist (macOS). Run it in its own goroutine.
func Log(name string, interval time.Duration) {
	if _, err := os.Stat("/proc/net/snmp"); err != nil {
		return
	}
	for {
		line, err := udpCounters()
		if err != nil {
			fmt.Fprintf(os.Stderr, "NETSTAT %v error: %v\n", name, err)
		} else {
			fmt.Fprintf(os.Stderr, "NETSTAT %v %v\n", name, line)
		}
		time.Sleep(interval)
	}
}

// udpCounters returns the Udp counters as "Name=value" pairs separated by spaces.
func udpCounters() (string, error) {
	file, err := os.Open("/proc/net/snmp")
	if err != nil {
		return "", err
	}
	defer file.Close()
	var rows [][]string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) > 0 && fields[0] == "Udp:" {
			rows = append(rows, fields[1:])
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if len(rows) < 2 || len(rows[0]) != len(rows[1]) {
		return "", fmt.Errorf("unexpected /proc/net/snmp format")
	}
	pairs := make([]string, len(rows[0]))
	for i := range rows[0] {
		pairs[i] = rows[0][i] + "=" + rows[1][i]
	}
	return strings.Join(pairs, " "), nil
}
