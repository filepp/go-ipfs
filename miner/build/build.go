package build

import (
	rice "github.com/GeertJohan/go.rice"
	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("build")

func SwarmKey() []byte {
	builtinKeys, err := rice.FindBox("keys")
	if err != nil {
		log.Warnf("loading built-in genesis: %s", err)
		return nil
	}
	keyBytes, err := builtinKeys.Bytes("swarm.key")
	if err != nil {
		log.Warnf("loading built-in swarm key: %s", err)
	}

	return keyBytes
}
