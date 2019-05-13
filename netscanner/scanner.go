package netscanner

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/Ullaakut/nmap"
	"github.com/gammazero/deque"
)

func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

type Host struct {
	Address string
	// Ports   []uint16
	Ports *deque.Deque
}

// var Deque deque.Deque

func Scan(ctx context.Context, myIP string, myPort uint16) (*deque.Deque, error) {
	scanner, err := nmap.NewScanner(
		nmap.WithTargets("127.0.0.1", "192.168.1.0/24"),
		nmap.WithPorts("8001,8002,8003"),
		nmap.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}

	result, err := scanner.Run()
	if err != nil {
		return nil, err
	}

	// hosts := Hosts{}
	q := &deque.Deque{}
	// Use the results to print an example output
	for _, host := range result.Hosts {
		if len(host.Ports) == 0 || len(host.Addresses) == 0 {
			continue
		}

		// fmt.Printf("Host %q:\n", host.Addresses[0])

		hostAddress := fmt.Sprintf("%s", host.Addresses[0])
		// if (hostAddress == myIP) {
		// continue
		// }

		ports := &deque.Deque{}

		for _, port := range host.Ports {
			fmt.Printf(
				"\tPort %d/%s %s %s\n",
				port.ID,
				port.Protocol,
				port.State,
				port.Service.Name,
			)
			if port.Status() == nmap.Open {
				if (hostAddress == myIP) && (port.ID == myPort) {
				} else {
					// ports = append(ports, strconv.Itoa(int(port.ID)))
					ports.PushBack(strconv.Itoa(int(port.ID)))
				}
			}
		}

		if ports.Len() != 0 {
			// for _, port := range ports {
			// 	// hosts[hostAddress] = append(hosts[hostAddress], port)
			// 	// ports = append(ports, port)
			// }

			host := Host{
				Address: hostAddress,
				Ports:   ports,
			}

			q.PushBack(host)
		}
	}

	fmt.Printf(
		"Nmap done: %d hosts up scanned in %3f seconds\n",
		len(result.Hosts),
		result.Stats.Finished.Elapsed,
	)

	return q, nil
}

func ScanMyIP(ctx context.Context, myIP string) (string, uint16, error) {
	scanner, err := nmap.NewScanner(
		nmap.WithTargets(myIP),
		nmap.WithPorts("8000,8001,8002,8003"),
		nmap.WithContext(ctx),
	)
	if err != nil {
		return "", 0, err
	}

	result, err := scanner.Run()
	if err != nil {
		return "", 0, err
	}

	var hostAddress string
	var port uint16
	// Use the results to print an example output
	for _, host := range result.Hosts {
		if len(host.Ports) == 0 || len(host.Addresses) == 0 {
			continue
		}

		hostAddress = fmt.Sprintf("%s", host.Addresses[0])

		ports := []uint16{}
		for _, port := range host.Ports {
			if port.Status() == nmap.Open {
				ports = append(ports, port.ID)
			}
		}

		if len(ports) != 0 {
			port = ports[0]
		}
	}

	// fmt.Printf("%+v\n", hosts[hostAddress])
	// panic(len(hosts))
	// if len(hosts) == 1 {
	// 	return hosts, nil
	// }

	return hostAddress, port, nil
}
