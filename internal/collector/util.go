package collector

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// readFileInt reads a single integer from a file.
func readFileInt(path string) (int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
}

// NetSNMPStats holds TCP and UDP statistics.
type NetSNMPStats struct {
	TCP map[string]int64
	UDP map[string]int64
}

// readNetSNMP reads /proc/net/snmp and parses it.
func readNetSNMP(procPath string) (*NetSNMPStats, error) {
	file, err := os.Open(procPath + "/net/snmp")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseSNMP(file)
}

func parseSNMP(file *os.File) (*NetSNMPStats, error) {
	stats := &NetSNMPStats{
		TCP: make(map[string]int64),
		UDP: make(map[string]int64),
	}

	scanner := bufio.NewScanner(file)
	var lastHeaders []string
	var lastProto string

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		proto := strings.TrimSuffix(parts[0], ":")

		// Check if header or value
		// value line usually has numbers.
		// header line has strings.
		isHeader := false
		if _, err := strconv.ParseInt(parts[1], 10, 64); err != nil {
			isHeader = true
		}

		if isHeader {
			lastHeaders = parts[1:]
			lastProto = proto
		} else {
			// Ensure this value line matches the last header we saw
			if proto != lastProto {
				continue // Skip or error?
			}
			// Parse values
			for i, valStr := range parts[1:] {
				if i >= len(lastHeaders) {
					break
				}
				val, err := strconv.ParseInt(valStr, 10, 64)
				if err == nil {
					if proto == "Tcp" {
						stats.TCP[lastHeaders[i]] = val
					} else if proto == "Udp" {
						stats.UDP[lastHeaders[i]] = val
					}
				}
			}
		}
	}
	return stats, scanner.Err()
}
