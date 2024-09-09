package shutdown

import (
	"os"
	"os/signal"
	"syscall"
)

func On(do func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	do()
}
