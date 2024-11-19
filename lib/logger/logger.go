package logger

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type GetCollectionFunc func() (*mongo.Collection, error)

type MongoZapCore[T any] struct {
	// level   zapcore.Level
	// encoder zapcore.Encoder

	buffer     []T
	bufferLock sync.Mutex
	bufferChan chan struct{}

	maxBatchSize  int
	flushInterval time.Duration

	GetCollection GetCollectionFunc
}

// func (core *MongoZapCore[T]) With(fields []zap.Field) zapcore.Core {
// 	return core
// }

// func (core *MongoZapCore[T]) Enabled(level zapcore.Level) bool {
// 	return level >= core.level
// }

// func (core *MongoZapCore[T]) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
// 	if core.Enabled(entry.Level) {
// 		return checked.AddCore(entry, core)
// 	}
// 	return checked
// }

// func (core *MongoZapCore[T]) Write(entry zapcore.Entry, fields []zapcore.Field) error {
// 	// log.Println("MongoZapCore Write", entry, fields)

// 	buf, err := core.encoder.EncodeEntry(entry, fields)
// 	if err != nil {
// 		return err
// 	}

// 	var logInfo T
// 	err = json.Unmarshal(buf.Bytes(), &logInfo)
// 	if err != nil {
// 		log.Println("MongoZapCore Write json.Unmarshal error", err)
// 		return err
// 	}

// 	core.bufferLock.Lock()
// 	core.buffer = append(core.buffer, logInfo)
// 	if len(core.buffer) >= core.maxBatchSize {
// 		core.triggerFlush()
// 	}
// 	core.bufferLock.Unlock()

//		return nil
//	}
func (core *MongoZapCore[T]) Write(bs []byte) (int, error) {

	var logInfo T
	err := json.Unmarshal(bs, &logInfo)
	if err != nil {
		log.Println("MongoZapCore Write json.Unmarshal error", err)
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

func (core *MongoZapCore[T]) flushBufferPeriodically() {
	for {
		select {
		case <-core.bufferChan:
			core.flushBuffer()
		case <-time.After(core.flushInterval):
			core.flushBuffer()
		}
	}
}

func (core *MongoZapCore[T]) flushBuffer() error {
	core.bufferLock.Lock()
	defer core.bufferLock.Unlock()

	// log.Println("flushBuffer")

	if len(core.buffer) == 0 {
		return nil
	}

	if core.GetCollection == nil {
		log.Println("MongoZapCore GetCollection is nil")
		return errors.New("GetCollection is nil")
	}

	collection, err := core.GetCollection()
	if err != nil {
		log.Println("MongoZapCore GetCollection error", err)
		return err
	}

	docs := make([]interface{}, len(core.buffer))
	for i, v := range core.buffer {
		docs[i] = v
	}
	_, err = collection.InsertMany(context.Background(), docs)
	if err != nil {
		log.Println("MongoZapCore InsertMany error", err)
		return err
	}

	core.buffer = core.buffer[:0] // 复用 buffer
	return nil
}

func (core *MongoZapCore[T]) Sync() error {
	log.Println("MongoZapCore Sync")
	return core.flushBuffer()
}

func (core *MongoZapCore[T]) triggerFlush() {
	// log.Println("triggerFlush")
	select {
	case core.bufferChan <- struct{}{}:
	default:
		log.Println("bufferChan is full")
	}
}

func NewMongoZapCore[T any](level zapcore.Level, gcf GetCollectionFunc) zapcore.Core {
	mzc := &MongoZapCore[T]{
		buffer:        make([]T, 0, 5),
		bufferLock:    sync.Mutex{},
		bufferChan:    make(chan struct{}, 1),
		maxBatchSize:  5,
		flushInterval: 5 * time.Second,
		GetCollection: gcf,
	}

	go mzc.flushBufferPeriodically()

	c := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(mzc), level)

	return c
}
