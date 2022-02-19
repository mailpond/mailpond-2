package mail

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	emailnormalize "github.com/yinyin/go-email-address-normalize"
)

func digestMailAddress(mailAddr string) (digestAddr string) {
	addrDigest := sha256.Sum256([]byte(mailAddr))
	digestAddr = base64.RawURLEncoding.EncodeToString(addrDigest[:])
	return
}

type Storage struct {
	lck        sync.Mutex
	baseFolder string
}

func NewStorage(baseFolder string) (s *Storage, err error) {
	if baseFolder, err = filepath.Abs(baseFolder); nil != err {
		return
	}
	fileInfo, err := os.Stat(baseFolder)
	if nil != err {
		return
	}
	if !fileInfo.IsDir() {
		err = fmt.Errorf("storage base folder is not directory: [%s]", baseFolder)
		return
	}
	s = &Storage{
		baseFolder: baseFolder,
	}
	return
}

func (s *Storage) writeMailForReceipt(rcptAddr, tickFolderPart string, encodedContent []byte) (err error) {
	_, normalizedAddr, err := emailnormalize.NormalizeEmailAddress(rcptAddr, nil)
	if nil != err {
		log.Printf("WARN: cannot normalize receipt address for writing mail [%s]: %v", rcptAddr, err)
		return
	}
	rcptFolderPart := digestMailAddress(normalizedAddr)
	destFolder := filepath.Join(s.baseFolder, tickFolderPart, rcptFolderPart)
	if err = os.MkdirAll(destFolder, 0600); nil != err {
		log.Printf("WARN: cannot create mail folder [%s]: %v", rcptAddr, err)
		return
	}
	fileNamePart := strconv.FormatInt(time.Now().UnixNano(), 10) + ".json"
	mailFilePath := filepath.Join(rcptFolderPart, fileNamePart)
	fp, err := os.Create(mailFilePath)
	if nil != err {
		log.Printf("WARN: cannot create mail file [%s]: %v", mailFilePath, err)
		return
	}
	defer fp.Close()
	if _, err = fp.Write(encodedContent); nil != err {
		log.Printf("WARN: cannot write mail file [%s]: %v", mailFilePath, err)
	}
	return
}

func (s *Storage) AddMail(ctx context.Context, m *Mail) (err error) {
	s.lck.Lock()
	defer s.lck.Unlock()
	tickFolderPart := strconv.FormatInt(time.Time(m.ReceiveAt).Unix()/3600, 10)
	contentBuf, err := json.Marshal(m)
	if nil != err {
		log.Printf("ERROR: cannot marshal mail content: %v", err)
		return
	}
	for _, rcptAddr := range m.ReceiptAddresses {
		if err = ctx.Err(); nil != err {
			return
		}
		s.writeMailForReceipt(rcptAddr, tickFolderPart, contentBuf)
	}
	return
}

func (s *Storage) loadMail(targetPath string) (m *Mail, err error) {
	fp, err := os.Open(targetPath)
	if nil != err {
		log.Printf("ERROR: cannot open mail content [%s]: %v", targetPath, err)
		return
	}
	defer fp.Close()
	var b Mail
	dec := json.NewDecoder(fp)
	if err = dec.Decode(&b); nil != err {
		log.Printf("ERROR: cannot unpack mail content [%s]: %v", targetPath, err)
		return
	}
	m = &b
	return
}

func (s *Storage) collectMails(existedResult []*Mail, folderPath string) (result []*Mail, err error) {
	result = existedResult
	dirEntries, err := os.ReadDir(folderPath)
	if nil != err {
		return
	}
	for _, dirEnt := range dirEntries {
		if dirEnt.IsDir() {
			continue
		}
		targetPath := filepath.Join(folderPath, dirEnt.Name())
		m, err := s.loadMail(targetPath)
		if nil != err {
			continue
		}
		result = append(result, m)
	}
	return
}

func (s *Storage) ListMails(ctx context.Context, receiptAddr string) (result []*Mail, err error) {
	s.lck.Lock()
	defer s.lck.Unlock()
	_, normalizedAddr, err := emailnormalize.NormalizeEmailAddress(receiptAddr, nil)
	if nil != err {
		log.Printf("WARN: cannot normalize receipt address for listing mail [%s]: %v", receiptAddr, err)
		return
	}
	rcptFolderPart := digestMailAddress(normalizedAddr)
	dirEntries, err := os.ReadDir(s.baseFolder)
	if nil != err {
		log.Printf("ERROR: cannot read folder for list mail: %v", err)
		return
	}
	for _, dirEnt := range dirEntries {
		if err = ctx.Err(); nil != err {
			return
		}
		if !dirEnt.IsDir() {
			continue
		}
		destFolder := filepath.Join(s.baseFolder, dirEnt.Name(), rcptFolderPart)
		fileInfo, err := os.Stat(destFolder)
		if nil != err {
			continue
		}
		if !fileInfo.IsDir() {
			continue
		}
		result, _ = s.collectMails(result, destFolder)
	}
	return

}

func (s *Storage) Purge(ctx context.Context, retain time.Duration) (err error) {
	s.lck.Lock()
	defer s.lck.Unlock()
	targetTick := (time.Now().Add(-retain).Unix() / 3600) - 1
	dirEntries, err := os.ReadDir(s.baseFolder)
	if nil != err {
		log.Printf("ERROR: cannot read folder for purge: %v", err)
		return
	}
	for _, dirEnt := range dirEntries {
		if err = ctx.Err(); nil != err {
			return
		}
		if !dirEnt.IsDir() {
			continue
		}
		t, err := strconv.ParseInt(dirEnt.Name(), 10, 64)
		if nil != err {
			continue
		}
		if t > targetTick {
			continue
		}
		expiredPath := filepath.Join(s.baseFolder, dirEnt.Name())
		if err = os.RemoveAll(expiredPath); nil != err {
			log.Printf("WARN: cannot remove expired folder [%s]: %v", expiredPath, err)
		}
	}
	return
}
