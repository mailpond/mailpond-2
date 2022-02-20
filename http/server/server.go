package server

import (
	"log"
	"net/http"
	"sync"

	"github.com/mailpond/mailpond-2"
)

func Serve(
	waitGroup *sync.WaitGroup,
	listenAddress, tlsCertFile, tlsKeyFile string,
	storageEngine *mailpond.Storage,
	staticContentFolderPath string) (httpServer *http.Server, err error) {
	hnd := &mailPondHTTPHandler{
		storageEngine:           storageEngine,
		staticContentFolderPath: staticContentFolderPath,
	}
	hnd.prepareStaticContentServer()
	httpServer = &http.Server{
		Addr:    listenAddress,
		Handler: hnd,
	}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		var err error
		if (tlsCertFile != "") && (tlsKeyFile != "") {
			err = httpServer.ListenAndServeTLS(tlsCertFile, tlsKeyFile)
		} else {
			err = httpServer.ListenAndServe()
		}
		log.Printf("INFO: stopped MailPond-2 HTTP service (err = %v)", err)
	}()
	return httpServer, nil
}
