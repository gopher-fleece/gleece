package ipc

type StdIoOptions struct{}

func (StdIoOptions) IpcType() LangServerIpcType { return LangServerIpcStdIo }
func (StdIoOptions) Validate() error            { return nil }
