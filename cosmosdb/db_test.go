package cosmosdb

import "testing"

func TestCreateItem(t *testing.T) {
	// init CosmosDB client
	client, err := NewClient()
	if err != nil {
		t.Error(err)
	}

	err = CreateItem(client, "bingWallpaper", "US", "test", "test")
	if err == nil {
		t.Error("CreateItem should return error")
	}

	item := WallpaperItem{
		ID:        "test",
		Month:     "test",
		Copyright: "",
		URL:       "",
	}
	err = CreateItem(client, "bingWallpaper", "US", "test", item)
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteItem(t *testing.T) {

	client, err := NewClient()
	if err != nil {
		t.Error(err)
	}

	err = DeleteItem(client, "bingWallpaper", "US", "test", "test123")
	if err == nil {
		t.Error("DeleteItem should return error")
	}

	item := WallpaperItem{
		ID:        "test",
		Month:     "test",
		Copyright: "",
		URL:       "",
	}
	err = DeleteItem(client, "bingWallpaper", "US", "test", item.ID)
	if err != nil {
		t.Error(err)
	}

}
