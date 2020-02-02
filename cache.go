package main

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/patrickmn/go-cache"
)

type ObjectCache interface {
	Set(snowflake.ID, interface{})
	Get(snowflake.ID) (interface{}, bool)
	Del(snowflake.ID)
	Limit(l int) int
}

type LocalObjectCache struct {
	cache    *cache.Cache
	maxItems int
}

var gCache = ObjectCache(&LocalObjectCache{cache.New(2*time.Minute, 3*time.Minute), 4096})

var sessCache = ObjectCache(&LocalObjectCache{cache.New(2*time.Minute, 10*time.Minute), 65536})

func (oc *LocalObjectCache) Set(id snowflake.ID, obj interface{}) {
	if oc.cache.ItemCount() > oc.maxItems {
		oc.cache.DeleteExpired()
		return
	}
	oc.cache.SetDefault(id.String(), obj)
}

func (oc *LocalObjectCache) Get(id snowflake.ID) (interface{}, bool) {
	obj, ok := oc.cache.Get(id.String())
	return obj, ok
}

func (oc *LocalObjectCache) Del(id snowflake.ID) {
	oc.cache.Delete(id.String())
}

func (oc *LocalObjectCache) Limit(i int) int {
	if i != -1 {
		oc.maxItems = i
	}
	return oc.maxItems
}
