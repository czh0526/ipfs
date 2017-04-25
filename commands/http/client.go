package http

import (
	"errors"
	"net/http"

	cmds "github.com/czh0526/ipfs/commands"
)

type Client interface {
	Send(req cmds.Request) (cmds.Response, error)
}

type client struct {
	serverAddress string
	httpClient    *http.Client
}

func NewClient(address string) Client {
	return &client{
		serverAddress: address,
		httpClient:    http.DefaultClient,
	}
}

func (c *client) Send(req cmds.Request) (cmds.Response, error) {
	return nil, errors.New("Client.Send() has not be implemented !")
}
