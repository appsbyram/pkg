// Copyright ©  2019 AppsByRam authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//Server base http server
type Server interface {
	Start()
}

type server struct {
	http.Server
	wait              time.Duration
	tls               bool
	certFile, keyFile string
}

//NewServer creates a new instance of http server
func NewServer(addr string, tls bool, certFile, keyFile string, routes Routes) Server {
	handler := newRouter(routes)

	return &server{
		Server: http.Server{
			Addr:    addr,
			Handler: handler,
		},
		tls:      tls,
		certFile: certFile,
		keyFile:  keyFile,
	}
}

//Start starts http server
func (s *server) Start() {

	done := make(chan bool)
	go func() {
		var err error

		if s.tls {
			err = s.ListenAndServeTLS(s.certFile, s.keyFile)
		} else {
			err = s.ListenAndServe()
		}
		if err != nil {
			fmt.Printf("Listen and serve: %v \n", err)
		}
		done <- true
	}()

	//wait shutdown
	s.waitShutdown()

	<-done
	fmt.Printf("DONE!")

}

func (s *server) waitShutdown() {
	irqSig := make(chan os.Signal, 1)
	signal.Notify(irqSig, syscall.SIGINT, syscall.SIGTERM)

	//Wait interrupt or shutdown request through /shutdown
	select {
	case sig := <-irqSig:
		fmt.Printf("\n Shutdown request (signal: %v)\n", sig)
	}

	fmt.Printf("Stopping server ...\n")

	//Create shutdown context with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//shutdown the server
	err := s.Shutdown(ctx)
	if err != nil {
		fmt.Printf("Shutdown request error: %v", err)
	}
}
