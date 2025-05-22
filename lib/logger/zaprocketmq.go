package logger

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"

	"github.com/apache/rocketmq-clients/golang/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type RocketZapCore[T any] struct {
	buffer      []T
	bufferLock  sync.Mutex
	triggerChan chan struct{}

	maxBatchSize  int
	flushInterval time.Duration

	writer golang.Producer
	topic  string
}

func (core *RocketZapCore[T]) Write(bs []byte) (int, error) {
	log.Info("RocketZapCore Write", "bs", string(bs))

	var logInfo T
	err := json.Unmarshal(bs, &logInfo)
	if err != nil {
		log.Error("RocketZapCore Write json.Unmarshal error", "err", err)
		return 0, err
	}

	core.bufferLock.Lock()
	core.buffer = append(core.buffer, logInfo)
	if len(core.buffer) >= core.maxBatchSize {
		core.triggerFlush()
	}
	core.bufferLock.Unlock()

	return len(bs), nil
}

func (core *RocketZapCore[T]) flushBufferPeriodically() {
	for {
		select {
		case <-core.triggerChan:
			core.flushBuffer()
		case <-time.After(core.flushInterval):
			core.flushBuffer()
		}
	}
}

func (core *RocketZapCore[T]) flushBuffer() error {
	core.bufferLock.Lock()
	defer core.bufferLock.Unlock()

	log.Info("flushBuffer", "len", len(core.buffer))

	if len(core.buffer) == 0 {
		return nil
	}

	// messages := make([]kafka.Message, len(core.buffer))
	for _, v := range core.buffer {
		bs, err := json.Marshal(v)
		if err != nil {
			log.Error("RocketZapCore json.Marshal error", "err", err)
			continue
		}

		msg := &golang.Message{
			Topic: core.topic,
			Body:  bs,
		}
		// set keys and tag
		// msg.SetKeys(fmt.Sprintf("%d", req.UserID), req.OrderID)

		msg.SetTag(os.Getenv("RMQ_MESSAGE_TAG"))

		// send message in sync
		resps, err := core.writer.Send(context.Background(), msg)
		if err != nil {
			return err
		}
		log.Info("send message success", "len(resps)", len(resps))
		for _, resp := range resps {
			log.Info("send message success", "msgId", resp.MessageID)
		}
	}

	// err := core.writer.WriteMessages(context.Background(),
	// 	messages...,
	// )
	// if err != nil {
	// 	log.Println("RocketZapCore WriteMessages error", err)
	// 	return err
	// }

	core.buffer = core.buffer[:0] // 复用 buffer
	return nil
}

func (core *RocketZapCore[T]) Sync() error {
	log.Info("RocketZapCore Sync")
	return core.flushBuffer()
}

func (core *RocketZapCore[T]) triggerFlush() {
	// log.Println("triggerFlush")
	select {
	case core.triggerChan <- struct{}{}:
	default:
		log.Info("triggerChan is full")
	}
}

func NewRocketMQZapCore[T any](level zapcore.Level, topic string, writer golang.Producer) zapcore.Core {
	mzc := &RocketZapCore[T]{
		buffer:        make([]T, 0, 5),
		bufferLock:    sync.Mutex{},
		triggerChan:   make(chan struct{}, 1),
		maxBatchSize:  5,
		flushInterval: 5 * time.Second,
		topic:         topic,
		writer:        writer,
	}

	go mzc.flushBufferPeriodically()

	c := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(mzc), level)

	return c
}
