package server

type LangServerIpcType string

const (
	LangServerIpcStdIo     LangServerIpcType = "stdio"
	LangServerIpcTcp       LangServerIpcType = "tcp"
	LangServerIpcWebsocket LangServerIpcType = "websocket"
	LangServerIpcNodeJs    LangServerIpcType = "nodejs"
)
