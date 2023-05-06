package main

import (
	"fmt"
	"testing"
	"wallpaper/cosmosdb"
)

func TestCreateItem(t *testing.T) {

	err := cosmosdb.CreateItem(client, "bingWallpaper", "US", "test", "test")
	if err == nil {
		t.Error("CreateItem should return error")
	}

	item := cosmosdb.WallpaperItem{
		ID:        "test",
		Month:     "test",
		Copyright: "",
		URL:       "",
	}
	err = cosmosdb.CreateItem(client, "bingWallpaper", "US", "test", item)
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteItem(t *testing.T) {

	err := cosmosdb.DeleteItem(client, "bingWallpaper", "US", "test", "test123")
	if err == nil {
		t.Error("DeleteItem should return error")
	}

	item := cosmosdb.WallpaperItem{
		ID:        "test",
		Month:     "test",
		Copyright: "",
		URL:       "",
	}
	err = cosmosdb.DeleteItem(client, "bingWallpaper", "US", "test", item.ID)
	if err != nil {
		t.Error(err)
	}

}

func TestQueryItems(t *testing.T) {
	pk := "202304"
	query := fmt.Sprintf("SELECT * FROM c WHERE c.Month = '%s'", "202304")
	results, err := cosmosdb.QueryItems(cosmosdb.Client, "bingWallpaper", "US", pk, query)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(len(results))
}
