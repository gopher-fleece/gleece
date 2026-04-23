package ipc

import "fmt"

type TcpOptions struct {
	Address string
}

func (TcpOptions) IpcType() LangServerIpcType { return LangServerIpcTcp }

func (o TcpOptions) Validate() error {
	if o.Address == "" {
		return fmt.Errorf("address is required for tcp ipc type")
	}
	return nil
}
