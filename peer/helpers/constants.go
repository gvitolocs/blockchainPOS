package helpers

// The protocol used for connections.
const PROTOCOL = "tcp"

// Send this message to a connection to let them know who you are when first establishing a connection with them.
const CONNECT_MESSAGE_TYPE = "Wave"
const PING_MESSAGE_TYPE = "Ping"
const PONG_MESSAGE_TYPE = "Pong"