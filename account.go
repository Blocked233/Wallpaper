package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"wallpaper/cosmosdb"
)

var (
	once sync.Once
)

type Account struct {
	ID       string `json:"id"`
	Username string `json:"Username"` // Partition Key
	Password string
	Email    string
}

func accountRegister(username, password, email string) error {

	once.Do(func() {
		err := cosmosdb.CreateContainer(client, databaseName, "Account", "/Username")
		if err != nil {
			panic(err)
		}
	})

	account := Account{
		ID:       cosmosdb.HashPartitionKey(username),
		Username: username,
		Password: password,
		Email:    email,
	}

	err := cosmosdb.CreateItem(client, databaseName, "Account", account.Username, account)
	if err != nil {
		return err
	}
	return nil
}

func accountLogin(username, password string) error {

	once.Do(func() {
		err := cosmosdb.CreateContainer(client, databaseName, "Account", "/Username")
		if err != nil {
			panic(err)
		}
	})

	partitionKey := username
	query := fmt.Sprintf("SELECT * FROM c WHERE c.Username = '%s'", partitionKey)
	itemBytes, err := cosmosdb.QueryItem(client, databaseName, "Account", partitionKey, query)
	if err != nil {
		return err
	}

	var account Account

	err = json.Unmarshal(itemBytes, &account)
	if err != nil {
		return err
	}

	if account.Password == password {
		return nil
	}
	return errors.New("password error")

}

func accountUpdate(username, oldPassword, newPassword string) error {

	once.Do(func() {
		err := cosmosdb.CreateContainer(client, databaseName, "Account", "/Username")
		if err != nil {
			panic(err)
		}
	})

	partitionKey := username
	query := fmt.Sprintf("SELECT * FROM c WHERE c.Username = '%s'", partitionKey)
	itemBytes, err := cosmosdb.QueryItem(client, databaseName, "Account", partitionKey, query)
	if err != nil {
		return err
	}

	var account Account

	err = json.Unmarshal(itemBytes, &account)
	if err != nil {
		return err
	}

	if account.Password != oldPassword {
		return errors.New("password error")
	}

	account.Password = newPassword

	err = cosmosdb.UpdateItem(client, databaseName, "Account", partitionKey, account.ID, account)
	if err != nil {
		return err
	}
	return nil
}

func accountDelete(username, password string) error {

	once.Do(func() {
		err := cosmosdb.CreateContainer(client, databaseName, "Account", "/Username")
		if err != nil {
			panic(err)
		}
	})

	partitionKey := username
	query := fmt.Sprintf("SELECT * FROM c WHERE c.Username = '%s'", partitionKey)
	itemBytes, err := cosmosdb.QueryItem(client, databaseName, "Account", partitionKey, query)
	if err != nil {
		return err
	}

	var account Account

	err = json.Unmarshal(itemBytes, &account)
	if err != nil {
		return err
	}

	if account.Password != password {
		return errors.New("password error")
	}

	err = cosmosdb.DeleteItem(client, databaseName, "Account", partitionKey, account.ID)
	if err != nil {
		return err
	}
	return nil
}
