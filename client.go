package main

import (
	"context"

	"github.com/benhenryhunter/grpc-server-client-testing/proto"
)

type Client struct {
	proto.StreamerClient
	ID               int
	DisconnectModulo int
}

func (c *Client) StreamStuff(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	stream, err := c.StreamerClient.StreamStuff(ctx, &proto.StreamStuffRequest{Id: int32(c.ID)})
	if err != nil {
		panic(err)
	}

	count := 0

	for {
		response, err := stream.Recv()
		if err != nil {
			panic(err)
		}
		count++
		println("Got response: ", response.Id, " for client: ", c.ID)

		if count%c.DisconnectModulo == 0 {
			if err := stream.CloseSend(); err != nil {
				println("error closing stream: ", err)
			}
			cancel()
			c.ID++
			go c.StreamStuff(context.Background())
			return
		}

	}
}
