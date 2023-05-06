package qiniu

import (
	"fmt"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

var (
	BucketManager *storage.BucketManager
	Domain        = ""
	Mac           *qbox.Mac
	AccessKey     = ""
	SecretKey     = ""
)

func NewQiniu(domain, ak, sk string) {
	Domain = domain
	AccessKey = ak
	SecretKey = sk
	Mac = qbox.NewMac(ak, sk)
}

func Upload2Qiniu(bucket, key, url string) {

	// download image from url
	cfg := storage.Config{}
	BucketManager = storage.NewBucketManager(Mac, &cfg)

	fsize, err := BucketManager.Fetch(url, bucket, key)
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
