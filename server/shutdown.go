/******************************************************************************
 *
 *  Description :
 *
 *  Graceful shutdown of the server
 *
 *****************************************************************************/

package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func signalHandler() <-chan bool {
	stop := make(chan bool)

	signchan := make(chan os.Signal, 1)
	signal.Notify(signchan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		// Wait for a signal. Don't care which signal it is
		sig := <-signchan
		log.Printf("Signal received: '%s', shutting down", sig)
		stop <- true
	}()

	return stop
}

func listenAndServe(addr string, stop <-chan bool) error {
	shuttingDown := false

	httpdone := make(chan bool)

	server := &http.Server{Addr: addr}
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}

	go func() {
		err = server.Serve(tcpGracefulListener{ln.(*net.TCPListener)})
		if shuttingDown {
			// Clear the error because this is not a failure
			err = nil
			log.Printf("HTTP server stopped")
		}
		httpdone <- true
	}()

	// Wait for either a termination signal or an error
loop:
	for {
		select {
		case <-stop:
			// Flip the flag that we are terminating and close the Accept-ing socket, so no new connections are possible
			shuttingDown = true
			ln.Close()

			// Wait for http server to stop Accept()-ing connections
			<-httpdone

			// Terminate all sessions
			globals.sessionStore.Shutdown()

			// Shutdown local cluster node, if it's a part of a cluster.
			globals.cluster.shutdown()

			// Shutdown the hub. The hub will shutdown topics
			hubdone := make(chan bool)
			globals.hub.shutdown <- hubdone

			// wait for the hub to finish
			<-hubdone

			break loop

		case <-httpdone:
			break loop
		}
	}
	return err
}

// tcpGracefulListener is a copy of tcpKeepAliveListener from https://golang.org/src/net/http/server.go)
// Code copied to gain access to TCPListener.Close()
type tcpGracefulListener struct {
	*net.TCPListener
}

func (ln tcpGracefulListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
