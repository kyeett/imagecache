package imagecache

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"

	"github.com/hajimehoshi/ebiten"
	"github.com/peterhellberg/gfx"
)

func NewEbiten() EbitenImageCache {
	return EbitenImageCache{
		cache: map[string]*ebiten.Image{},
	}
}

type EbitenImageCache struct {
	cache             map[string]*ebiten.Image
	monitoringUpdates bool
	fileWatcher       *fsnotify.Watcher
}

// MonitorUpdates is a bit of a mess. Haven't really thought this out, but it works for now.
func (c *EbitenImageCache) MonitorUpdates() {
	if c.monitoringUpdates {
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	c.fileWatcher = watcher
	c.monitoringUpdates = true

	for {
		select {
		case event, ok := <-c.fileWatcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				c.loadImage(event.Name)
			}
		case err, ok := <-c.fileWatcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func (c *EbitenImageCache) loadImage(path string) {
	img, err := gfx.OpenImage(path)
	if err != nil {
		log.Fatal(err)
	}
	tmp, err := ebiten.NewImageFromImage(img, ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}
	c.cache[path] = tmp
}

func (c *EbitenImageCache) addToCache(path string) {
	c.loadImage(path)
	fmt.Println("addToCache", path, c.monitoringUpdates)
	if c.monitoringUpdates {
		// Add to fileWatcher
		err := c.fileWatcher.Add(path)
		fmt.Printf("Added %q\n", path)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (c *EbitenImageCache) CachedImage(path string) *ebiten.Image {
	if _, ok := c.cache[path]; !ok {
		c.addToCache(path)
	}
	return c.cache[path]
}
