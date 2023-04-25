package cache

import (
	"encoding/json"
	"fmt"
	"wallpaper/cosmosdb"
)

var (
	Webp *Group
)

func init() {
	Webp = NewGroup("webp", GetterFunc(func(key string) ([]byte, error) {

		partitionKey := key[:6]
		query := fmt.Sprintf("SELECT * FROM c WHERE c.id = '%s'", key)

		result, err := cosmosdb.QueryItem(cosmosdb.Client, "bingWallpaper", "US", partitionKey, query)
		if err != nil {
			return nil, err
		}

		wallpaperItem := &cosmosdb.WallpaperItem{}
		err = json.Unmarshal(result, wallpaperItem)
		if err != nil {
			return nil, err
		}

		return wallpaperItem.Webp, nil

	}))

}
