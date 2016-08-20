package proto

import (
	"fmt"
	"os"

	"github.com/pborman/uuid"
)

func hostname() string {
	host, err := os.Hostname()
	if err != nil {
		host = "Host-Unknown"
	}
	return host
}

func pid() int {
	return os.Getpid()
}

func gid() string {
	return uuid.New()
}

func Ident() string {
	return fmt.Sprintf("%v-pid-%v-uuid-%v", hostname(), pid(), gid())
}

func SenderId() string {
	return gid()
}
