package main

import (
	"github.com/cameliot/alpaca"
)

const PING = "@meta/PING"
const PONG = "@meta/PONG"

type PongPayload struct {
	Timestamp int64  `json"timestamp"`
	Handle    string `json:"handle"`
}

func Ping(handle string) alpaca.Action {
	return alpaca.Action{
		Type:    PING,
		Payload: handle,
	}
}
