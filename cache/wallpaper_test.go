package cache

import (
	"errors"
	"testing"
	"wallpaper/config"
	"wallpaper/cosmosdb"
)

func TestGetFromDatabase(t *testing.T) {

	config := config.GetConfig()
	cosmosdb.NewClient(config.Endpoint, config.Key)

	w := NewWallPaperGroup()

	b, err := w.getter.Get("20021221")

	if err == nil {
		t.Error(errors.New("should be error"), b, err)
	}

	b, err = w.getter.Get("12312231312")
	if err == nil {
		t.Error(errors.New("should be error"), b, err)
	}

	b, err = w.getter.Get("20230501")
	if err != nil {
		t.Error(err)
	}

	if len(b) == 0 {
		t.Error(errors.New("should not be empty"))
	}

}
