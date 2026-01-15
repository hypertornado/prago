package prago

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *App) initExitHandler() {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, os.Interrupt)
		<-sigChan
		app.exiting = true
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}()

}
