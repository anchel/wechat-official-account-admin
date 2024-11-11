package lru

import (
	"container/list"
	"context"
	"log"
	"sync"
)

type CacheLRU[T any] struct {
	maxCount int
	itemsMap sync.Map
	list     *list.List
	Lock     sync.Mutex

	creator func(ctx context.Context, key string) (*T, error)
}

// 多一层这个结构，是为了从list获得的元素取得key，然后再去map里面操作key
type CacheLRUListItem[T any] struct {
	Key         string
	BusinessObj *T
}

func NewCacheLRU[T any](maxCount int, creator func(ctx context.Context, key string) (*T, error)) *CacheLRU[T] {
	return &CacheLRU[T]{
		maxCount: maxCount,
		itemsMap: sync.Map{},
		list:     list.New(),
		Lock:     sync.Mutex{},
		creator:  creator,
	}
}

func (c *CacheLRU[T]) Get(ctx context.Context, key string) (*T, error) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	elementPtr, ok := c.itemsMap.Load(key)
	if ok {
		// log.Println("found in map", key)
		// log.Println("--------------------------------------------------")
		element := elementPtr.(*list.Element)
		c.list.Remove(element)

		newElement := c.list.PushFront(element.Value)
		c.itemsMap.Store(key, newElement)

		lruItem, _ := element.Value.(*CacheLRUListItem[T])
		return lruItem.BusinessObj, nil
	}

	bo, err := c.creator(ctx, key)
	if err != nil {
		return nil, err
	}

	// log.Println("not found in map", key)
	// 如果没有找到，新建一个，然后插入进去
	// 插入之前，判断元素个数是否已经达到最大限制
	if c.list.Len() >= c.maxCount {
		log.Println("lru list full, remove tail key", key)
		// 删除尾部最后那个
		element := c.list.Back()
		if element != nil {
			lruItem, _ := element.Value.(*CacheLRUListItem[T])
			log.Println("list full, remove tail key", lruItem.Key)

			c.list.Remove(element)
			c.itemsMap.Delete(lruItem.Key)
		}
	}

	lruItem := &CacheLRUListItem[T]{
		Key:         key,
		BusinessObj: bo,
	}
	newElement := c.list.PushFront(lruItem)
	c.itemsMap.Store(key, newElement)
	log.Println("lru create new one and insert front", key)
	// log.Println("--------------------------------------------------")

	return bo, nil
}
