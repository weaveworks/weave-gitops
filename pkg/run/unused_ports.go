package run

import "net"

// GetUnusedPorts returns a slice of unused ports.
func GetUnusedPorts(count int) (ports []int32, retErr error) {
	for i := 0; i < count; i++ {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			return nil, err
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return nil, err
		}

		defer func(l *net.TCPListener) {
			err := l.Close()
			if err != nil {
				retErr = err
			}
		}(l)

		port := int32(l.Addr().(*net.TCPAddr).Port)
		ports = append(ports, port)
	}

	return ports, nil
}
