package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultMailDomainAddr = "example.net"
	defaultSMTPListenAddr = ":1025"
	defaultHTTPListenAddr = ":8080"
)

const (
	defaultSMTPMaxRecipients = 16
	minSMTPMaxRecipients     = 1
	maxSMTPMaxRecipients     = 100

	defaultSMTPMaxMessageBytes = 1024 * 1024 * 20
	minSMTPMaxMessageBytes     = 128
	maxSMTPMaxMessageBytes     = 1024 * 1024 * 1024 * 1

	defaultSMTPReadTimeout = time.Second * 30
	minSMTPReadTimeout     = time.Second * 1
	maxSMTPReadTimeout     = time.Minute * 30

	defaultSMTPWriteTimeout = time.Second * 30
	minSMTPWriteTimeout     = time.Second * 1
	maxSMTPWriteTimeout     = time.Minute * 30

	defaultMailRetainDuration = time.Hour * 24 * 7
	minMailRetainDuration     = time.Hour * 2
)

type commandOptions struct {
	mailDomainAddress       string
	smtpListenAddr          string
	smtpTLSCert             string
	smtpTLSKey              string
	smtpMaxRecipients       int
	smtpMaxMessageBytes     int
	smtpReadTimeout         time.Duration
	smtpWriteTimeout        time.Duration
	httpListenAddr          string
	httpTLSCert             string
	httpTLSKey              string
	httpStaticContentFolder string
	storagePath             string
	mailRetainDuration      time.Duration
}

func (opts *commandOptions) normalize() (err error) {
	if opts.mailDomainAddress == "" {
		opts.mailDomainAddress = os.Getenv("MAILPOND_DOMAIN")
	}
	if opts.mailDomainAddress == "" {
		opts.mailDomainAddress = defaultMailDomainAddr
	}
	if opts.smtpListenAddr == "" {
		opts.smtpListenAddr = defaultSMTPListenAddr
	}
	if opts.smtpMaxRecipients == 0 {
		opts.smtpMaxRecipients = defaultSMTPMaxRecipients
	} else if (opts.smtpMaxRecipients < minSMTPMaxRecipients) || (opts.smtpMaxRecipients > maxSMTPMaxRecipients) {
		err = fmt.Errorf("option smtpMaxRecipients out of acceptable range: %v", opts.smtpMaxRecipients)
		return
	}
	if opts.smtpMaxMessageBytes == 0 {
		opts.smtpMaxMessageBytes = defaultSMTPMaxMessageBytes
	} else if (opts.smtpMaxMessageBytes < minSMTPMaxMessageBytes) || (opts.smtpMaxMessageBytes > maxSMTPMaxMessageBytes) {
		err = fmt.Errorf("option smtpMaxMessageBytes out of acceptable range: %v", opts.smtpMaxMessageBytes)
		return
	}
	if opts.smtpReadTimeout == 0 {
		opts.smtpReadTimeout = defaultSMTPReadTimeout
	} else if (opts.smtpReadTimeout < minSMTPReadTimeout) || (opts.smtpReadTimeout > maxSMTPReadTimeout) {
		err = fmt.Errorf("option smtpReadTimeout out of acceptable range: %v", opts.smtpReadTimeout)
		return
	}
	if opts.smtpWriteTimeout == 0 {
		opts.smtpReadTimeout = defaultSMTPWriteTimeout
	} else if (opts.smtpWriteTimeout < minSMTPWriteTimeout) || (opts.smtpWriteTimeout > maxSMTPWriteTimeout) {
		err = fmt.Errorf("option smtpWriteTimeout out of acceptable range: %v", opts.smtpWriteTimeout)
		return
	}
	if opts.httpStaticContentFolder != "" {
		if opts.httpStaticContentFolder, err = filepath.Abs(opts.httpStaticContentFolder); nil != err {
			return
		}
	}
	if opts.storagePath == "" {
		opts.storagePath = os.Getenv("MAILPOND_STORAGE")
	}
	if opts.storagePath == "" {
		err = errors.New("option storagePath is required")
		return
	}
	if opts.storagePath, err = filepath.Abs(opts.storagePath); nil != err {
		return
	}
	if opts.mailRetainDuration == 0 {
		opts.mailRetainDuration = defaultMailRetainDuration
	} else if opts.mailRetainDuration < minMailRetainDuration {
		err = fmt.Errorf("option mailRetainDuration out of acceptable range: %v", opts.mailRetainDuration)
		return
	}
	return
}

func parseCommandFlags() (cmdOpts *commandOptions, err error) {
	cmdOpts = &commandOptions{}
	flag.StringVar(&cmdOpts.mailDomainAddress, "mailDomain", "", "mail domain address")
	flag.StringVar(&cmdOpts.smtpListenAddr, "smtpListen", defaultSMTPListenAddr, "listen address of SMTP service")
	flag.StringVar(&cmdOpts.smtpTLSCert, "smtpTLSCert", "", "path of TLS certification file for SMTP service")
	flag.StringVar(&cmdOpts.smtpTLSKey, "smtpTLSKey", "", "path of TLS private key for SMTP service")
	flag.IntVar(&cmdOpts.smtpMaxRecipients, "smtpMaxRecipients", defaultSMTPMaxRecipients, "max number of recipients per mail for SMTP service")
	flag.IntVar(&cmdOpts.smtpMaxMessageBytes, "smtpMaxMessageBytes", defaultSMTPMaxMessageBytes, "max mail size in bytes for SMTP service")
	flag.DurationVar(&cmdOpts.smtpReadTimeout, "smtpReadTimeout", defaultSMTPReadTimeout, "read timeout for SMTP service")
	flag.DurationVar(&cmdOpts.smtpWriteTimeout, "smtpWriteTimeout", defaultSMTPWriteTimeout, "write timeout for SMTP service")
	flag.StringVar(&cmdOpts.httpListenAddr, "httpListenAddr", defaultHTTPListenAddr, "listen address of HTTP service")
	flag.StringVar(&cmdOpts.httpTLSCert, "httpTLSCert", "", "path of TLS certification file for HTTP service")
	flag.StringVar(&cmdOpts.httpTLSKey, "httpTLSKey", "", "path of TLS private key for HTTP service")
	flag.StringVar(&cmdOpts.httpStaticContentFolder, "httpStaticContent", "", "path to static content folder for HTTP service")
	flag.StringVar(&cmdOpts.storagePath, "storagePath", "", "path to mail storage folder")
	flag.DurationVar(&cmdOpts.mailRetainDuration, "mailRetainDuration", defaultMailRetainDuration, "duration to kept received mail")
	flag.Parse()
	if err = cmdOpts.normalize(); nil != err {
		cmdOpts = nil
	}
	return
}
