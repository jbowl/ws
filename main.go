package main

import (
	"fmt"
	"github/jbowl/ws/server"
	"os"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

// All - return true iff len of all envVars > 0
func All(envVars ...string) bool {
	for _, envVar := range envVars {
		if len(envVar) < 1 {
			return false
		}
	}
	return true
}

var Healthy int64

func run() error {

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel) // set logging to how

	log.WithFields(log.Fields{}).Info("starting up with these settings")
	//	log.SetOutput(os.Stderr) // reset default output

	//read env variables``
	port := os.Getenv("PORT")
	brewery := os.Getenv("BREWERY_ADDR")

	log.WithFields(log.Fields{
		"PORT":         port,    // default port
		"BREWERY_ADDR": brewery, // grpc brewery api     ip:addr
	}).Info("starting up with these settings")
	log.SetOutput(os.Stderr) // reset default output

	log.Printf("logging")
	log.Println("PORT=", port)
	if !All(port, brewery) {
		log.Printf("env arg missing")
		log.Fatal("env arg missing")
	}

	atomic.StoreInt64(&Healthy, time.Now().UnixNano())

	svr := &server.Server{}

	svr.Healthy = &Healthy

	shutdownSig := svr.Start(port, brewery)

	<-shutdownSig

	atomic.StoreInt64(&Healthy, 0)

	//sigChannel := make(chan os.Signal, 1)
	//signal.Notify(sigChannel, os.Interrupt)
	//<-sigChannel // kill signal  ,  // force kill fuser -k apiPort/tcp
	return nil
}

func main() {

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

}
