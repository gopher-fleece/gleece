package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gopher-fleece/gleece/v2/internal/lsp/server"
	"github.com/gopher-fleece/gleece/v2/internal/lsp/server/ipc"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type lspRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type lspResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   json.RawMessage `json:"error,omitempty"`
}

type LSPClient struct {
	conn   io.ReadWriteCloser
	reader *bufio.Reader
	mu     sync.Mutex
	nextID int64
}

func newLSPClient(conn io.ReadWriteCloser) *LSPClient {
	return &LSPClient{
		conn:   conn,
		reader: bufio.NewReader(conn),
		nextID: 1,
	}
}

func (c *LSPClient) Send(method string, params any) (*lspResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := c.nextID
	c.nextID++

	req := lspRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	if _, err := fmt.Fprintf(c.conn, "Content-Length: %d\r\n\r\n", len(payload)); err != nil {
		return nil, err
	}
	if _, err := c.conn.Write(payload); err != nil {
		return nil, err
	}

	contentLength := 0
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		if strings.HasPrefix(strings.ToLower(line), "content-length:") {
			value := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			contentLength, err = strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
		}
	}

	body := make([]byte, contentLength)
	if _, err := io.ReadFull(c.reader, body); err != nil {
		return nil, err
	}

	var resp lspResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *LSPClient) Close() error {
	return c.conn.Close()
}

func runLsp() *LSPClient {
	const address = "127.0.0.1:43891"

	langServer, err := server.NewLanguageServer(ipc.TcpOptions{
		Address: address,
	})
	if err != nil {
		panic(err)
	}

	go func() {
		if err := langServer.Run(); err != nil {
			panic(err)
		}
	}()

	var conn net.Conn
	for range 20 {
		conn, err = net.Dial("tcp", address)
		if err == nil {
			return newLSPClient(conn)
		}
		time.Sleep(50 * time.Millisecond)
	}

	panic(fmt.Errorf("failed to connect to lsp server at %s: %w", address, err))
}

func sendInit(client *LSPClient) {
	resp, err := client.Send("initialize", map[string]any{
		"processId":    os.Getpid(),
		"rootUri":      nil,
		"capabilities": map[string]any{},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Current PID: %v, Resp: %v", os.Getpid(), string(resp.Result))
}

func sendDidChangeTextDocument(client *LSPClient, uri string) {
	body := protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: uri,
			},
		},
	}

	resp, err := client.Send("textDocument/didChange", body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Current PID: %v, Resp: %v", os.Getpid(), string(resp.Result))
}

func main() {
	client := runLsp()
	defer func() {
		_ = client.Close()
	}()

	sendInit(client)
	sendDidChangeTextDocument(
		client,
		"/mnt/7e91759c-6dd7-4c99-8d38-e6422452a469/git/gleece/test/sanity/sanity.controller.go",
	)
}
