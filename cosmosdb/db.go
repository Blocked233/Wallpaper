package cosmosdb

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

var (
	Endpoint string
	Key      string
)

func init() {

	// Initialize Azure Cosmos DB
	Endpoint = ""
	Key = ""

}

// this is a helper function that swallows 409 errors
func ErrorIs409(err error) bool {
	var responseErr *azcore.ResponseError
	return err != nil && errors.As(err, &responseErr) && responseErr.StatusCode == 409
}

func NewClient() (*azcosmos.Client, error) {
	// Create a credential object and Start to create a client
	cred, err := azcosmos.NewKeyCredential(Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create a credential: %w", err)
	}

	client, err := azcosmos.NewClientWithKey(Endpoint, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Cosmos DB db client: %w", err)
	}

	return client, nil
}

func CreateDatabase(client *azcosmos.Client, databaseName string) error {
	//	databaseName := "adventureworks"

	databaseProperties := azcosmos.DatabaseProperties{ID: databaseName}

	// This is a helper function that swallows 409 errors
	errorIs409 := func(err error) bool {
		var responseErr *azcore.ResponseError
		return err != nil && errors.As(err, &responseErr) && responseErr.StatusCode == 409
	}
	ctx := context.TODO()
	databaseResp, err := client.CreateDatabase(ctx, databaseProperties, nil)

	switch {
	case errorIs409(err):
		log.Printf("Database [%s] already exists\n", databaseName)
	case err != nil:
		return err
	default:
		log.Printf("Database [%v] created. ActivityId %s\n", databaseName, databaseResp.ActivityID)
	}
	return nil
}

func CreateContainer(client *azcosmos.Client, databaseName, containerName, partitionKey string) error {

	databaseClient, err := client.NewDatabase(databaseName)
	if err != nil {
		return err
	}

	// creating a container
	containerProperties := azcosmos.ContainerProperties{
		ID: containerName,
		PartitionKeyDefinition: azcosmos.PartitionKeyDefinition{
			Paths: []string{partitionKey},
		},
	}

	// this is a helper function that swallows 409 errors
	errorIs409 := func(err error) bool {
		var responseErr *azcore.ResponseError
		return err != nil && errors.As(err, &responseErr) && responseErr.StatusCode == 409
	}

	// setting options upon container creation
	throughputProperties := azcosmos.NewManualThroughputProperties(400) //defaults to 400 if not set
	options := &azcosmos.CreateContainerOptions{
		ThroughputProperties: &throughputProperties,
	}
	ctx := context.TODO()
	containerResponse, err := databaseClient.CreateContainer(ctx, containerProperties, options)

	switch {
	case errorIs409(err):
		log.Printf("Container [%s] already exists\n", containerName)
	case err != nil:
		return err
	default:
		log.Printf("Container [%s] created. ActivityId %s\n", containerName, containerResponse.ActivityID)
	}
	return nil
}

// partitionKey is item.xx
func CreateItem(client *azcosmos.Client, databaseName, containerName, partitionKey string, Item any) error {

	// create container client
	containerClient, err := client.NewContainer(databaseName, containerName)
	if err != nil {
		return fmt.Errorf("failed to create a container client: %s", err)
	}

	// specifies the value of the partiton key
	pk := azcosmos.NewPartitionKeyString(partitionKey)

	b, err := json.Marshal(Item)
	if err != nil {
		return err
	}
	// setting the Item options upon creating ie. consistency level
	ItemOptions := azcosmos.ItemOptions{
		ConsistencyLevel: azcosmos.ConsistencyLevelSession.ToPtr(),
	}

	ctx := context.TODO()
	ItemResponse, err := containerClient.CreateItem(ctx, pk, b, &ItemOptions)

	switch {
	case ErrorIs409(err):
		log.Printf("Item with partitionkey value %s already exists\n", pk)
		return err
	case err != nil:
		return err
	default:
		log.Printf("Status %d. Item %v created. ActivityId %s. Consuming %v Request Units.\n", ItemResponse.RawResponse.StatusCode, pk, ItemResponse.ActivityID, ItemResponse.RequestCharge)
	}

	return nil
}

func ReadItemWithID(client *azcosmos.Client, databaseName, containerName, partitionKey, ItemId string) ([]byte, error) {

	// Create container client
	containerClient, err := client.NewContainer(databaseName, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create a container client: %s", err)
	}

	// Specifies the value of the partiton key
	pk := azcosmos.NewPartitionKeyString(partitionKey)

	// Read an Item
	ctx := context.TODO()
	ItemResponse, err := containerClient.ReadItem(ctx, pk, ItemId, nil)
	if err != nil {
		return nil, err
	}

	return ItemResponse.Value, nil
}

func DeleteItem(client *azcosmos.Client, databaseName, containerName, partitionKey, ItemId string) error {

	// Create container client
	containerClient, err := client.NewContainer(databaseName, containerName)
	if err != nil {
		return fmt.Errorf("failed to create a container client:: %s", err)
	}
	// Specifies the value of the partiton key
	pk := azcosmos.NewPartitionKeyString(partitionKey)

	// Delete an Item
	ctx := context.TODO()

	res, err := containerClient.DeleteItem(ctx, pk, ItemId, nil)
	if err != nil {
		return err
	}

	log.Printf("Status %d. Item %v deleted. ActivityId %s. Consuming %v Request Units.\n", res.RawResponse.StatusCode, pk, res.ActivityID, res.RequestCharge)

	return nil
}

func UpdateItem(client *azcosmos.Client, databaseName, containerName, partitionKey, ItemId string, newItem any) error {

	// Create container client
	containerClient, err := client.NewContainer(databaseName, containerName)
	if err != nil {
		return fmt.Errorf("failed to create a container client:: %s", err)
	}
	// Specifies the value of the partiton key
	pk := azcosmos.NewPartitionKeyString(partitionKey)

	// Create a new context
	ctx := context.TODO()

	// Marshal the Item to JSON
	b, err := json.Marshal(newItem)
	if err != nil {
		return err
	}

	// Replace the Item in Cosmos DB
	_, err = containerClient.ReplaceItem(ctx, pk, ItemId, b, nil)

	if err != nil {
		return fmt.Errorf("failed to replace Item: %v", err)
	}

	log.Printf("Item %s updated", ItemId)
	return nil
}

func HashPartitionKey(partitionKey string) string {
	hasher := md5.New()
	hasher.Write([]byte(partitionKey))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return hash
}

type WallpaperItem struct {
	ID        string `json:"id"`
	Month     string `json:"Month"`
	Copyright string
	URL       string
	Bytes     []byte
}

func QueryWallpaperItems(client *azcosmos.Client, databaseName, containerName string, partitionKey string, query string) (results []WallpaperItem, err error) {

	// create container client
	containerClient, err := client.NewContainer(databaseName, containerName)
	if err != nil {
		return results, fmt.Errorf("failed to create a container client: %s", err)
	}

	// specifies the value of the partiton key
	pk := azcosmos.NewPartitionKeyString(partitionKey)

	// read multiple Items
	//ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	ctx := context.TODO()

	ItemPager := containerClient.NewQueryItemsPager(query, pk, nil)

	for ItemPager.More() {

		ItemList, err := ItemPager.NextPage(ctx)
		if err != nil {
			return results, err
		}
		for _, ItemBytes := range ItemList.Items {

			ItemResponseBody := &WallpaperItem{}

			err = json.Unmarshal(ItemBytes, ItemResponseBody)
			if err != nil {
				return results, err
			}

			results = append(results, *ItemResponseBody)
		}

	}

	return results, nil
}

// QueryItem queries an Item in the container and returns the first result
func QueryItem(client *azcosmos.Client, databaseName, containerName string, partitionKey string, query string) ([]byte, error) {

	// create container client
	containerClient, err := client.NewContainer(databaseName, containerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create a container client: %s", err)
	}

	// specifies the value of the partiton key
	pk := azcosmos.NewPartitionKeyString(partitionKey)

	// read multiple Items
	//ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	ctx := context.TODO()

	ItemPager := containerClient.NewQueryItemsPager(query, pk, nil)

	for ItemPager.More() {

		ItemList, err := ItemPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, ItemBytes := range ItemList.Items {

			return ItemBytes, nil
		}

	}

	return nil, nil
}
