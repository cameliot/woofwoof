package main

import (
	"fmt"
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
	fmt.Println("Decoding TimestampMS:", payload.TimestampMs)

	sec := payload.TimestampMs / 1000
	nsec := 1000000 * (payload.TimestampMs % 1000)

	return time.Unix(sec, nsec)
}

func DecodePongPayload(action alpaca.Action) PongPayload {
	payload := PongPayload{}
	err := action.DecodePayload(&payload)
	fmt.Println("err:", err)

	return payload
}

func Ping(handle string) alpaca.Action {
	return alpaca.Action{
		Type:    PING,
		Payload: handle,
	}
}
