package cache

import (
	"blue/api"
	"blue/models"
	"log"
	"sync"
)

type StickerCache struct {
	sets map[string]*models.StickerSet
	mu   sync.RWMutex
	bot  *api.Bot
}

func NewStickerCache(bot *api.Bot) *StickerCache {
	return &StickerCache{
		sets: make(map[string]*models.StickerSet),
		bot:  bot,
	}
}

func (c *StickerCache) Load(setName string) error {
	stickerSet, err := c.bot.GetStickerSet(setName)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.sets[setName] = stickerSet
	c.mu.Unlock()

	log.Printf("Loaded sticker set: %s (%d stickers)", setName, len(stickerSet.Stickers))
	return nil
}

func (c *StickerCache) Get(setName string) (*models.StickerSet, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	set, exists := c.sets[setName]
	return set, exists
}
