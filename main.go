package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/randsw/validationwebhook/cmd"
	"github.com/randsw/validationwebhook/pkg/logger"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

type application struct {
	client kubernetes.Interface
}

func main() {
	logger.InitLogger()
	defer logger.CloseLogger()
	//Create channel for signal
	cancelChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	done := make(chan bool, 1)
	go func() {
		sig := <-cancelChan
		logger.Info("Caught signal", zap.String("Signal", sig.String()))
		logger.Info("Wait for 1 second to finish processing")
		time.Sleep(1 * time.Second)
		logger.Info("Exiting.....")
		// shutdown other goroutines gracefully
		// close other resources
		done <- true
		os.Exit(0)

	}()
	cmd.Execute()
}
