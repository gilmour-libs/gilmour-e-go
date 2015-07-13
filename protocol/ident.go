package protocol

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"os"
)

func getHostname() string {
	host, err := os.Hostname()
	if err != nil {
		host = "Host-Unknown"
	}
	return host
}

func getPid() int {
	return os.Getpid()
}

func getUUID() string {
	return uuid.New()
}

func MakeIdent() string {
	hostname := getHostname()
	pid := getPid()
	uuid := getUUID()

	return fmt.Sprintf("%v-pid-%v-uuid-%v", hostname, pid, uuid)
}

func MakeSenderId() string {
	return getUUID()
}
