package main

type Message struct {
	Type    string // The type is an indicator for receiver on how to respond.
	MsgID   string // For later (see exercise manual).
	From    string // Name of sender.
	Payload []byte // Any data passed with the message.
}
