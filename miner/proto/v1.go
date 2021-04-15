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

	MsgWindowPost         = "WindowPost"
	MsgWindowPostResponse = "WindowPostResp"

	MsgMinerHeartBeat = "MinerHeartBeat"
)

const (
	StatusOK             = 0
	StatusFetchFileError = 1
)

type (
	Message struct {
		Type  string
		Nonce string
		Data  interface{}
	}
	FetchFileReq struct {
		Cid cid.Cid
	}
	FetchFileResp struct {
		Cid    cid.Cid
		Status int
	}
	WindowPostReqItem struct {
		FileCid   cid.Cid
		Positions []int64
	}
	WindowPostReq struct {
		Items []WindowPostReqItem
	}

	WindowPostRespItem struct {
		FileCid   cid.Cid
		Positions []int64
		Data      []byte
	}
	WindowPostResp struct {
		Items []WindowPostRespItem
	}

	MinerHartBeat struct {
		WalletAddress string
		Role          int
	}
)

func init() {
	gob.Register(Message{})
	gob.Register(FetchFileReq{})
	gob.Register(FetchFileResp{})
	gob.Register(WindowPostReq{})
	gob.Register(WindowPostResp{})
	gob.Register(MinerHartBeat{})
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
	dec := gob.NewDecoder(bytes.NewReader(data))
	var v Message
	err := dec.Decode(&v)
	if err != nil {
		return Message{}, err
	}
	return v, nil
}

func V1InternalTopic(id string) string {
	return V1 + "/internal/" + id
}

func V1ExternalTopic(id string) string {
	return V1 + "/external/" + id
}

func V1MinerHeartBeatTopic() string {
	return V1 + "/miner/heartbeat"
}
