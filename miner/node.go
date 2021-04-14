package miner

import (
	"context"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/go-ipfs/miner/proto"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"runtime/debug"
	"time"
)

var log = logging.Logger("miner")

func Run(ctx context.Context, node *core.IpfsNode, walletAddress string) {
	miner := &Miner{
		node:          node,
		walletAddress: walletAddress,
	}
	api, err := coreapi.NewCoreAPI(node, options.Api.FetchBlocks(true))
	if err != nil {
		log.Errorf("")
		return
	}
	miner.handler = NewV1Handler(api, miner)

	go miner.Run(ctx)
}

type Miner struct {
	node          *core.IpfsNode
	walletAddress string
	handler       MessageHandler
}

func (m *Miner) Run(ctx context.Context) {
	m.subscribe()
	m.heartbeat(ctx)

	ticker := time.NewTicker(time.Minute * 10)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.heartbeat(ctx)
		}
	}
}

func (m *Miner) subscribe() error {
	topic, err := m.node.PubSub.Join(proto.V1InternalTopic(m.node.Identity.String()))
	if err != nil {
		log.Errorf("failed to create sub topic: %v", err)
		return err
	}
	sub, err := topic.Subscribe()
	if err != nil {
		log.Errorf("failed to subscribe: %v", err)
		return err
	}
	log.Infof("subscribe: %v", proto.V1InternalTopic(m.node.Identity.String()))

	go func() {
		for {
			pmsg, err := sub.Next(context.Background())
			if err != nil {
				log.Errorf("failed get message: %v", err)
				time.Sleep(time.Second)
				continue
			}
			log.Infof("received message from %v: %v", pmsg.ReceivedFrom.String(), pmsg.String())
			go func() {
				defer func() {
					if err := recover(); err != nil {
						log.Errorf("%v", string(debug.Stack()))
					}
				}()
				msg, err := proto.DecodeMessage(pmsg.Data)
				if err != nil {
					log.Errorf("failed to decode message: %v", err)
					return
				}
				err = m.handler.Handle(context.TODO(), pmsg.ReceivedFrom, &msg)
				if err != nil {
					log.Errorf("failed to handler message: %v", err)
				}
			}()
		}
	}()
	return nil
}

func (m *Miner) PublishMessage(ctx context.Context, topic string, msg *proto.Message) error {
	data, err := msg.EncodeMessage()
	if err != nil {
		log.Errorf("failed to encode message: %v", err)
		return err
	}
	receiverTopic, err := m.node.PubSub.Join(topic)
	if err != nil {
		log.Errorf("failed to create pub message: %v", err)
		return err
	}
	defer receiverTopic.Close()

	err = receiverTopic.Publish(ctx, data)
	if err != nil {
		log.Errorf("failed publish message: %v", err)
	}
	return nil
}

func (m *Miner) heartbeat(ctx context.Context) error {
	msgResp := proto.Message{
		Type: proto.MsgMinerHeartBeat,
		Data: proto.MinerHartBeat{
			WalletAddress: m.walletAddress,
		},
	}
	err := m.PublishMessage(ctx, proto.V1MinerHeartBeatTopic(), &msgResp)
	if err != nil {
		log.Errorf("failed to publish:%v", err.Error())
		return err
	}
	return nil
}
