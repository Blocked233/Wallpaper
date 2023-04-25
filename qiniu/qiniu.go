package qiniu

import (
	"fmt"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

var (
	Bucket        = ""
	BucketManager *storage.BucketManager
	Domain        = ""
	Mac           *qbox.Mac
	AccessKey     = ""
	SecretKey     = ""
)

func init() {

	Mac = qbox.NewMac(AccessKey, SecretKey)
	cfg := storage.Config{}
	BucketManager = storage.NewBucketManager(Mac, &cfg)

}

func Upload2Qiniu(key, url string) {

	// download image from url
	fsize, err := BucketManager.Fetch(url, Bucket, key)
	if err != nil {
		fmt.Println("fetch error:", err)
		return
	}
	fmt.Printf("fetch %s, file size: %s", key, fsize.String())

}

func Key2PrivateUrl(key string, deadline int64) string {

	privateAccessURL := storage.MakePrivateURL(Mac, Domain, key, deadline)
	return privateAccessURL
}

func Key2PublicUrl(key string) string {

	publicAccessURL := storage.MakePublicURL(Domain, key)
	return publicAccessURL
}
