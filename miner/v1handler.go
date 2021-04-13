package miner

import (
	"context"
	"errors"
	"github.com/ipfs/go-cid"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/go-ipfs/miner/proto"
	"github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/libp2p/go-libp2p-core/peer"
	"io"
	"os"
)

type V1Handler struct {
	api        iface.CoreAPI
	handleFunc map[string]HandleFunc
	publisher  MessagePublisher
}

func NewV1Handler(api iface.CoreAPI, publisher MessagePublisher) *V1Handler {
	h := &V1Handler{
		api:        api,
		handleFunc: make(map[string]HandleFunc),
		publisher:  publisher,
	}
	h.handleFunc[proto.MsgFetchFile] = h.FetchFile
	h.handleFunc[proto.MsgWindowPost] = h.WindowPost
	return h
}

func (h *V1Handler) Handle(ctx context.Context, receivedFrom peer.ID, msg *proto.Message) error {
	if f, ok := h.handleFunc[msg.Type]; ok {
		return f(ctx, receivedFrom, msg)
	}
	log.Warnf("message type not register: %v", msg.Type)
	return nil
}

func (h *V1Handler) FetchFile(ctx context.Context, receivedFrom peer.ID, msg *proto.Message) error {
	fmsg, _ := msg.Data.(proto.FetchFileReq)
	resp := proto.FetchFileResp{
		Cid:    fmsg.Cid,
		Status: proto.StatusOK,
	}
	err := h.doFetchFile(ctx, fmsg)
	if err != nil {
		resp.Status = proto.StatusFetchFileError
	}

	msgResp := proto.Message{
		Type:  proto.MsgFetchFileResponse,
		Nonce: msg.Nonce,
		Data:  resp,
	}
	err = h.publisher.PublishMessage(ctx, proto.V1Topic(receivedFrom.String()), &msgResp)
	if err != nil {
		log.Errorf("failed to publish:%v", err.Error())
		return err
	}
	return nil
}

func (h *V1Handler) doFetchFile(ctx context.Context, fmsg proto.FetchFileReq) error {
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
	return nil
}

func (h *V1Handler) WindowPost(ctx context.Context, receivedFrom peer.ID, msg *proto.Message) error {
	respItems := h.doWindowPost(ctx, msg)
	resp := proto.WindowPostResp{
		Items: respItems,
	}
	msgResp := proto.Message{
		Type:  proto.MsgWindowPostResponse,
		Nonce: msg.Nonce,
		Data:  resp,
	}
	err := h.publisher.PublishMessage(ctx, proto.V1Topic(receivedFrom.String()), &msgResp)
	if err != nil {
		log.Errorf("failed to publish:%v", err.Error())
		return err
	}
	return nil
}

func (h *V1Handler) doWindowPost(ctx context.Context, msg *proto.Message) []proto.WindowPostRespItem {
	req, _ := msg.Data.(proto.WindowPostReq)
	items := make([]proto.WindowPostRespItem, len(req.Items))

	pins, err := accPins(h.api.Pin().Ls(ctx))
	if err != nil {
		log.Errorf("failed to get pins:%v", err.Error())
		return items
	}
	for i, item := range req.Items {
		respItem := proto.WindowPostRespItem{
			FileCid:   item.FileCid,
			Positions: item.Positions,
		}
		if _, exist := pins[item.FileCid.String()]; !exist {
			log.Warnf("file not exist: %v", item.FileCid)
		} else {
			data, err := h.getFileDataAtFixedPosition(ctx, item.FileCid, item.Positions)
			if err != nil {
				log.Warnf("failed to get file data: %v", err)
			} else {
				respItem.Data = data
			}
		}
		items[i] = respItem
	}
	return items
}

func accPins(pins <-chan iface.Pin, err error) (map[string]iface.Pin, error) {
	if err != nil {
		return nil, err
	}
	result := make(map[string]iface.Pin)
	for pin := range pins {
		if pin.Err() != nil {
			return nil, pin.Err()
		}
		result[pin.Path().Cid().String()] = pin
	}
	return result, nil
}

func (h *V1Handler) getFileDataAtFixedPosition(ctx context.Context, fileCid cid.Cid, positions []int64) ([]byte, error) {
	cidPath := path.New("/ipfs/" + fileCid.String())

	fileNode, err := h.api.Unixfs().Get(ctx, cidPath)
	if err != nil {
		log.Errorf("failed to get:%v", err.Error())
		return nil, err
	}
	defer fileNode.Close()
	file, ok := fileNode.(files.File)
	if !ok {
		log.Warnf("file type error")
		return nil, errors.New("file type error")
	}

	size, _ := file.Size()
	data := make([]byte, len(positions))
	for i, pos := range positions {
		if pos < 0 || pos >= size {
			log.Warnf("out of range: %v %v %v", fileCid.String(), size, pos)
			return data, errors.New("out off range")
		}
		file.Seek(pos, io.SeekStart)
		var buf [1]byte
		file.Read(buf[:])
		data[i] = buf[0]
	}
	return data, nil
}
