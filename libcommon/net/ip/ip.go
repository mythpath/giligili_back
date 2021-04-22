package ip

import "net"

// LocalAddrs is used to return all ipv4 addr of local host
func LocalAddrs() ([]net.IP, error) {

	localAddrs := []net.IP{}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				localAddrs = append(localAddrs, ipNet.IP)
			}
		}
	}

	return localAddrs, nil
}
