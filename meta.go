package main

import (
	"github.com/cameliot/alpaca"
	"time"
)

const PING = "@meta/PING"
const PONG = "@meta/PONG"

type PongPayload struct {
	TimestampMs int64  `json:"timestamp"`
	Handle      string `json:"handle"`
}

/*
Decode int64 millisecond timestamp
*/
func (payload PongPayload) Timestamp() time.Time {
	sec := payload.TimestampMs / 1000
	nsec := 1000000 * (payload.TimestampMs % 1000)

	return time.Unix(sec, nsec).UTC()
}

func DecodePongPayload(action alpaca.Action) PongPayload {
	payload := PongPayload{}
	action.DecodePayload(&payload)

	return payload
}

func Ping(handle string) alpaca.Action {
	return alpaca.Action{
		Type:    PING,
		Payload: handle,
	}
}
