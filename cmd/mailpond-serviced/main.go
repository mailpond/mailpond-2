package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/mailpond/mailpond-2"
	httpserver "github.com/mailpond/mailpond-2/http/server"
	smtpserver "github.com/mailpond/mailpond-2/smtp/server"
)

func runServices(waitGroup *sync.WaitGroup, cmdOpts *commandOptions) (err error) {
	storageEngine, err := mailpond.NewStorage(cmdOpts.storagePath)
	if nil != err {
		log.Fatalf("ERROR: cannot init storage at [%s]: %v", cmdOpts.storagePath, err)
		return
	}
	smtpSrv, err := smtpserver.Serve(
		waitGroup,
		cmdOpts.smtpListenAddr,
		cmdOpts.smtpTLSCert, cmdOpts.smtpTLSKey,
		storageEngine,
		cmdOpts.mailDomainAddress,
		cmdOpts.smtpMaxRecipients, cmdOpts.smtpMaxMessageBytes,
		cmdOpts.smtpReadTimeout, cmdOpts.smtpWriteTimeout)
	if nil != err {
		log.Fatalf("ERROR: cannot start SMTP service: %v", err)
		return
	}
	defer smtpSrv.Close()
	httpSrv, err := httpserver.Serve(
		waitGroup,
		cmdOpts.httpListenAddr,
		cmdOpts.httpTLSCert, cmdOpts.httpTLSKey,
		storageEngine,
		cmdOpts.httpStaticContentFolder)
	if nil != err {
		log.Fatalf("ERROR: cannot start HTTP service: %v", err)
		return
	}
	defer httpSrv.Close()
	ticker := time.NewTicker(time.Hour * 2)
	defer ticker.Stop()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			log.Print("INFO: stopping MailPond-2 service.")
			return
		case <-ticker.C:
			storageEngine.Purge(ctx, cmdOpts.mailRetainDuration)
		}
	}
}

func main() {
	cmdOpts, err := parseCommandFlags()
	if nil != err {
		log.Fatalf("ERROR: missing required options: %v", err)
		return
	}
	log.Printf("INFO: domain = [%s]", cmdOpts.mailDomainAddress)
	log.Printf("INFO: SMTP listen = [%s]", cmdOpts.smtpListenAddr)
	log.Printf("INFO: HTTP listen = [%s]", cmdOpts.httpListenAddr)
	log.Printf("INFO: storage = [%s]", cmdOpts.storagePath)
	var wg sync.WaitGroup
	runServices(&wg, cmdOpts)
	wg.Wait()
	log.Print("INFO: stopped MailPond-2 service.")
}
