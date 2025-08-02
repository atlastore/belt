package logx

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestLoggerBuild(t *testing.T) {
	ctx := context.Background()

	cfg := NewConfig(Development, "master", uuid.NewString())

	logger := New(ctx, cfg, &ConsoleTransport{})
	// logger.Info("test", zap.String("foo", "bar"), zap.Bool("bool", true), zap.Strings("strings", []string{"hello", "world"}), zap.Binary("byte", []byte("hello world")), zap.Error(errors.New("yalla")))

	for i := range 2 {
		logger.Info(fmt.Sprintf("test-%v",i+1), zap.String("foo", "bar"))
	}

	time.Sleep(2*time.Second)

	child := logger.With(zap.String("child", "logger"))
	child.Info("test", zap.String("foo", "bar"), zap.Bool("bool", true), zap.Strings("strings", []string{"hello", "world"}), zap.Binary("byte", []byte("hello world")), zap.Error(errors.New("yalla")))

	for i := range 2 {
		child.Info(fmt.Sprintf("test-%v",i+1), zap.String("foo", "bar"))
	}

	time.Sleep(2*time.Second)

	logger.Close()
}