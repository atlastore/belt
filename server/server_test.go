package server

import (
	"context"
	"testing"

	"github.com/atlastore/belt/logx"
	"github.com/google/uuid"
)

func TestServer(t *testing.T) {
	ctx := context.Background()
	log := logx.New(ctx, logx.NewConfig(logx.Development, "master", uuid.NewString()), &logx.ConsoleTransport{})
	server := NewServer(HTTP, log)
	err := server.Start(ctx, "localhost:4000")

	if err != nil {
		panic(err)
	}
}