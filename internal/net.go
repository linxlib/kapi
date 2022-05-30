package internal

import (
	"net"
	"os"
)

func GetIntranetIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		Log.Errorln(err)
		os.Exit(1)
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}

		}
	}
	return "localhost"
}
