package commands

import (
	"blue/api"
	"blue/cache"
	"blue/models"
)

type CommandFunc func(*api.Bot, *models.Message, []string, *cache.StickerCache)

var registry = make(map[string]CommandFunc)

func Register(name string, handler CommandFunc) {
	registry[name] = handler
}

func Get(name string) (CommandFunc, bool) {
	handler, exists := registry[name]
	return handler, exists
}
