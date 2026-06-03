package web

import "fmt"

func ListenAddr(port int) string {
	return fmt.Sprintf("127.0.0.1:%d", port)
}
