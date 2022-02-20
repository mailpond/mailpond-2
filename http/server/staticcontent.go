package server

import (
	"log"
	"os"

	httpservefile "github.com/yinyin/go-util-http-serve-file"
)

const staticContentURLPathPrefix = "/"
const staticContentDefaultFileName = "index.html"

func makeStaticContentServer(contentStorePath string) (contentServer httpservefile.HTTPFileServer) {
	fileInfo, err := os.Stat(contentStorePath)
	if nil != err {
		log.Printf("WARN: cannot stat static content storage [%s]: %v", contentStorePath, err)
		return
	}
	if !fileInfo.IsDir() {
		log.Printf("WARN: static content storage is not a folder [%s]", contentStorePath)
		return
	}
	c, err := httpservefile.NewServeFileSystemWithPrefix(staticContentURLPathPrefix, contentStorePath)
	if nil != err {
		log.Printf("WARN: cannot open static content storage [%s]: %v", contentStorePath, err)
		return
	}
	return c
}
