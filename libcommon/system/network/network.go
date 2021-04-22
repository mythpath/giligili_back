package network


import (
"net"
"strings"
"log"
"errors"
"math/rand"
)

func LocalAddrs() ([]string, error) {

	localAddrs := []string{}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return []string{}, err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				localAddrs = append(localAddrs, ipNet.IP.String())
			}
		}
	}

	return localAddrs, nil
}

func LocalAddr(addrs []string) (string, error) {

	localAddr := ""
	size := len(addrs)
	invalid := 0

	for _, i := range rand.Perm(size) {
		conn, err := net.Dial("tcp", addrs[i])
		if err != nil {
			log.Printf("failed to dial target addr<%s>", addrs[i])
			invalid += 1
			continue
		}

		localAddr = strings.Split(conn.LocalAddr().String(), ":")[0]
		log.Printf("successfully get local address: %s", localAddr)
		conn.Close()
		break
	}

	if invalid >= size {
		return localAddr, errors.New("failed to get local addr")
	}

	return localAddr, nil
}

