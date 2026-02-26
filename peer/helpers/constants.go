package helpers

// The protocol used for connections.
const PROTOCOL = "tcp"

// The size of the message header (in bytes).
const MESSAGE_HEADER_SIZE = 4

// Send this message to a connection to let them know who you are when first establishing a connection with them.
const CONNECT_MESSAGE_TYPE = "Wave"

// Ping message to check whether a connection is active.
const PING_MESSAGE_TYPE = "Ping"

// Reply to ping message to acknowledge connection.
const PONG_MESSAGE_TYPE = "Pong"
