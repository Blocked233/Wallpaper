package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	// Redis
	RedisPassword string
	// Database
	Endpoint string
	Key      string
	// Qiniu
	Domin     string
	AccessKey string
	SecretKey string
	Bucket    string
}

func GetConfig() *Config {

	config := &Config{}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetDefault("RedisPassword", "")
	viper.SetDefault("Endpoint", "http://localhost:6379")
	viper.SetDefault("Key", "mykey")
	viper.SetDefault("Domin", "mydomin")
	viper.SetDefault("AccessKey", "accesskey")
	viper.SetDefault("SecretKey", "secretkey")
	viper.SetDefault("Bucket", "mybucket")
	err := viper.ReadInConfig()
	if err != nil {

		fmt.Println("File Not Exist, Producing...")
		err = viper.SafeWriteConfig()
		if err != nil {
			fmt.Println("Config Fail:", err)
			return nil
		}
		fmt.Println("New Config, Please Edit it and Restart ")
		return nil
	}

	config.RedisPassword = viper.GetString("RedisPassword")
	config.Endpoint = viper.GetString("Endpoint")
	config.Key = viper.GetString("Key")
	config.Domin = viper.GetString("Domin")
	config.AccessKey = viper.GetString("AccessKey")
	config.SecretKey = viper.GetString("SecretKey")
	config.Bucket = viper.GetString("Bucket")

	fmt.Println("Information:")
	fmt.Printf("RedisPassword: %s\n", config.RedisPassword)
	fmt.Printf("Endpoint: %s\n", config.Endpoint)
	fmt.Printf("Key: %s\n", config.Key)
	fmt.Printf("Domin: %s\n", config.Domin)
	fmt.Printf("AccessKey: %s\n", config.AccessKey)
	fmt.Printf("SecretKey: %s\n", config.SecretKey)
	fmt.Printf("Bucket: %s\n", config.Bucket)
	return config
}
