package miner

import (
	"context"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/go-ipfs/miner/proto"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/libp2p/go-libp2p-core/peer"
	"os"
)

type V1Handler struct {
	api         coreiface.CoreAPI
	handlerFunc map[string]HandleFunc
	publisher   MessagePublisher
}

func NewV1Handler(api coreiface.CoreAPI, publisher MessagePublisher) *V1Handler {
	h := &V1Handler{
		api:         api,
		handlerFunc: make(map[string]HandleFunc),
		publisher:   publisher,
	}
	h.handlerFunc[proto.MsgFetchFile] = h.HandleFetchFile
	h.handlerFunc[proto.MsgFileStat] = h.HandleFileStat
	return h
}

func (h *V1Handler) Handle(ctx context.Context, receivedFrom peer.ID, msg *proto.Message) error {
	if f, ok := h.handlerFunc[msg.Type]; ok {
		return f(ctx, receivedFrom, msg)
	}
	log.Warnf("message type not register: %v", msg.Type)
	return nil
}

func (h *V1Handler) HandleFetchFile(ctx context.Context, receivedFrom peer.ID, msg *proto.Message) error {
	fmsg, _ := msg.Data.(proto.FetchFile)
	cidPath := path.New("/ipfs/" + fmsg.Cid.String())

	fileNode, err := h.api.Unixfs().Get(ctx, cidPath)
	if err != nil {
		log.Errorf("failed to get:%v", err.Error())
		return err
	}
	defer fileNode.Close()
	err = files.WriteTo(fileNode, os.TempDir()+"/"+fmsg.Cid.String())
	if err != nil {
		log.Errorf("failed to write file:%v", err.Error())
		return err
	}
	err = h.api.Pin().Add(ctx, cidPath)
	if err != nil {
		log.Errorf("failed to add pin:%v", err.Error())
		return err
	}
	resp := proto.FetchFileResp{
		Cid: fmsg.Cid,
	}
	msgResp := proto.Message{
		Type: proto.MsgAddFileResponse,
		Data: resp,
	}
	err = h.publisher.PublishMessage(ctx, proto.V1Topic(receivedFrom.String()), &msgResp)
	if err != nil {
		log.Errorf("failed to publish:%v", err.Error())
		return err
	}
	return nil
}

func (h *V1Handler) HandleFileStat(ctx context.Context, receivedFrom peer.ID, msg *proto.Message) error {
	//TODO:
	return nil
}
