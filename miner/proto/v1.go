package proto

import (
	"bytes"
	"encoding/gob"
	"github.com/ipfs/go-cid"
)

const (
	V1                   = "v1"
	MsgFetchFile         = "FetchFile"
	MsgFetchFileResponse = "FetchFileResp"

	MsgFileStat         = "FileStat"
	MsgFileStatResponse = "FileStatResp"
)

const (
	StatusOK             = 0
	StatusFetchFileError = 1
)

type (
	Message struct {
		Type  string
		Nonce uint64
		Data  interface{}
	}
	FetchFile struct {
		Cid cid.Cid
	}
	FetchFileResp struct {
		Cid    cid.Cid
		Status int
	}
	QueryFileState struct {
		Cids []cid.Cid
	}
	CidStat struct {
		Cid   cid.Cid
		Exist bool
	}
	QueryFileStateResp struct {
		Cids []CidStat
	}
)

func init() {
	gob.Register(Message{})
	gob.Register(FetchFile{})
	gob.Register(FetchFileResp{})
	gob.Register(QueryFileState{})
	gob.Register(CidStat{})
	gob.Register(QueryFileStateResp{})
}

func (m Message) EncodeMessage() ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(m)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func DecodeMessage(data []byte) (Message, error) {
	var buffer bytes.Buffer
	buffer.Write(data)
	dec := gob.NewDecoder(&buffer)
	var v Message
	err := dec.Decode(&v)
	if err != nil {
		return Message{}, err
	}
	return v, nil
}

func V1Topic(id string) string {
	return V1 + "/" + id
}
