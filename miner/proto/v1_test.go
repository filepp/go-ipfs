package proto

import (
	"encoding/hex"
	"github.com/ipfs/go-cid"
	"testing"
)

func TestDecodeMessage(t *testing.T) {
	c, _ := cid.Decode("bafybeickencdqw37dpz3ha36ewrh4undfjt2do52chtcky4rxkj447qhdm")
	msg := Message{
		Type: MsgAddFile,
		Data: FetchFile{
			Cid: c,
		},
	}
	data, err := msg.EncodeMessage()
	if err != nil {
		t.Errorf("%v", err.Error())
	}
	t.Logf("data: %s", hex.EncodeToString(data))

	msg2, _ := DecodeMessage(data)

	t.Logf("msg2: %+v", msg2)
	ff, ok := msg2.Data.(FetchFile)
	if !ok {
		t.Error("data type error")
	}
	t.Logf("ff: %+v", ff)
}
