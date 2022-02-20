package server

import (
	"net/http"
	"strings"

	httpservefile "github.com/yinyin/go-util-http-serve-file"

	"github.com/mailpond/mailpond-2"
)

type mailPondHTTPHandler struct {
	storageEngine *mailpond.Storage

	staticContentFolderPath string
	staticContentServer     httpservefile.HTTPFileServer
}

func (hnd *mailPondHTTPHandler) prepareStaticContentServer() {
	if hnd.staticContentFolderPath == "" {
		return
	}
	hnd.staticContentServer = makeStaticContentServer(hnd.staticContentFolderPath)
}

func (hnd *mailPondHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if reqPath := strings.TrimPrefix(r.URL.Path, "/endpoints/get-mails-of/"); len(reqPath) < len(r.URL.Path) {
		hnd.getMailsOfReceiptAddr(w, r, reqPath)
		return
	}
	if nil != hnd.staticContentServer {
		hnd.staticContentServer.ServeHTTP(w, r, staticContentDefaultFileName, "")
		return
	}
	http.NotFound(w, r)
}
