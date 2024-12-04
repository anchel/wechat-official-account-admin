package logger

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type KafkaZapCore[T any] struct {
	buffer      []T
	bufferLock  sync.Mutex
	triggerChan chan struct{}

	maxBatchSize  int
	flushInterval time.Duration

	writer *kafka.Writer
}

func (core *KafkaZapCore[T]) Write(bs []byte) (int, error) {

	var logInfo T
	err := json.Unmarshal(bs, &logInfo)
	if err != nil {
		log.Println("KafkaZapCore Write json.Unmarshal error", err)
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

func (core *KafkaZapCore[T]) flushBufferPeriodically() {
	for {
		select {
		case <-core.triggerChan:
			core.flushBuffer()
		case <-time.After(core.flushInterval):
			core.flushBuffer()
		}
	}
}

func (core *KafkaZapCore[T]) flushBuffer() error {
	core.bufferLock.Lock()
	defer core.bufferLock.Unlock()

	// log.Println("flushBuffer", len(core.buffer))

	if len(core.buffer) == 0 {
		return nil
	}

	// if core.GetCollection == nil {
	// 	log.Println("KafkaZapCore GetCollection is nil")
	// 	return errors.New("GetCollection is nil")
	// }

	// collection, err := core.GetCollection()
	// if err != nil {
	// 	log.Println("KafkaZapCore GetCollection error", err)
	// 	return err
	// }

	// docs := make([]interface{}, len(core.buffer))
	// for i, v := range core.buffer {
	// 	docs[i] = v
	// }
	// _, err = collection.InsertMany(context.Background(), docs)
	// if err != nil {
	// 	log.Println("KafkaZapCore InsertMany error", err)
	// 	return err
	// }

	messages := make([]kafka.Message, len(core.buffer))
	for i, v := range core.buffer {
		bs, err := json.Marshal(v)
		if err != nil {
			log.Println("KafkaZapCore json.Marshal error", err)
			continue
		}
		messages[i] = kafka.Message{
			Key:   []byte("Key-A"),
			Value: bs,
		}
	}

	err := core.writer.WriteMessages(context.Background(),
		messages...,
	)
	if err != nil {
		log.Println("KafkaZapCore WriteMessages error", err)
		return err
	}

	core.buffer = core.buffer[:0] // 复用 buffer
	return nil
}

func (core *KafkaZapCore[T]) Sync() error {
	log.Println("KafkaZapCore Sync")
	return core.flushBuffer()
}

func (core *KafkaZapCore[T]) triggerFlush() {
	// log.Println("triggerFlush")
	select {
	case core.triggerChan <- struct{}{}:
	default:
		log.Println("triggerChan is full")
	}
}

func NewKafkaZapCore[T any](level zapcore.Level, writer *kafka.Writer) zapcore.Core {
	mzc := &KafkaZapCore[T]{
		buffer:        make([]T, 0, 5),
		bufferLock:    sync.Mutex{},
		triggerChan:   make(chan struct{}, 1),
		maxBatchSize:  5,
		flushInterval: 5 * time.Second,
		writer:        writer,
	}

	go mzc.flushBufferPeriodically()

	c := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(mzc), level)

	return c
}
