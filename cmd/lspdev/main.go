package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
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
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	mu     sync.Mutex
	nextID int64
}

func newLSPClient(cmd *exec.Cmd) (*LSPClient, error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &LSPClient{
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		nextID: 1,
	}, nil
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

	if _, err := fmt.Fprintf(c.stdin, "Content-Length: %d\r\n\r\n", len(payload)); err != nil {
		return nil, err
	}
	if _, err := c.stdin.Write(payload); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(c.stdout)

	contentLength := 0
	for {
		line, err := reader.ReadString('\n')
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
	if _, err := io.ReadFull(reader, body); err != nil {
		return nil, err
	}

	var resp lspResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func runLsp() *LSPClient {
	// Replace this with the actual server executable entrypoint you want to launch.
	cmd := exec.Command(os.Args[0])

	client, err := newLSPClient(cmd)
	if err != nil {
		panic(err)
	}

	return client
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

	fmt.Println(string(resp.Result))
}

func main() {
	client := runLsp()
	defer func() {
		_ = client.stdin.Close()
		_ = client.stdout.Close()
		_ = client.stderr.Close()
	}()

	sendInit(client)
}
