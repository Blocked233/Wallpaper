package cache

import (
	"encoding/json"
	"fmt"
	"time"
	"wallpaper/cosmosdb"
)

var (
	Wallpaper *Group
)

func NewWallPaperGroup() *Group {
	Wallpaper = NewGroup("wallpaper", GetterFunc(func(key string) ([]byte, error) {

		if len(key) != 8 {
			return nil, fmt.Errorf("invalid key")
		}

		partitionKey := key[:6]

		query := fmt.Sprintf("SELECT * FROM c WHERE c.Month = '%s'", partitionKey)
		results, err := cosmosdb.QueryItems(cosmosdb.Client, "bingWallpaper", "US", partitionKey, query)
		if err != nil {
			return nil, err
		}

		var b []byte
		wallpaperItem := &cosmosdb.WallpaperItem{}
		for _, result := range results {

			json.Unmarshal(result, wallpaperItem)

			RedisClient.Set(wallpaperItem.ID, result, time.Hour)
			if wallpaperItem.ID == key {
				b = result
			}
		}

		if len(b) == 0 {
			return nil, fmt.Errorf("not found")
		}

		return b, nil

	}))
	return Wallpaper
}
