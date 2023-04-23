package main

import (
	"testing"
	"wallpaper/cosmosdb"
)

func TestAccountRegister(t *testing.T) {

	err := accountRegister("reg", "reg", "reg")
	if err != nil {
		t.Error(err)
	}

	err = accountRegister("reg", "reg", "reg")
	if cosmosdb.ErrorIs409(err) == false {
		t.Error(err)
	}

	err = accountDelete("reg", "reg")
	if err != nil {
		t.Error(err)
	}
}

func TestAccountLogin(t *testing.T) {

	err := accountRegister("log", "log", "log")
	if err != nil {
		t.Error(err)
	}

	err = accountLogin("log", "log")
	if err != nil {
		t.Error(err)
	}

	err = accountDelete("log", "log")
	if err != nil {
		t.Error(err)
	}
}

func TestAccountLoginFail(t *testing.T) {

	err := accountLogin("fail", "fail")
	if err == nil {
		t.Error(err)
	}
}

func TestAccountUpdate(t *testing.T) {

	err := accountRegister("upa", "upa", "upa")
	if err != nil {
		t.Error(err)
	}

	err = accountUpdate("upa", "upa", "new")
	if err != nil {
		t.Error(err)
	}

	err = accountLogin("upa", "new")
	if err != nil {
		t.Error(err)
	}

	err = accountDelete("upa", "new")
	if err != nil {
		t.Error(err)
	}

}

func TestAccountDelete(t *testing.T) {

	err := accountRegister("delete", "delete", "delete")
	if err != nil {
		t.Error(err)
	}

	err = accountDelete("delete", "delete")
	if err != nil {
		t.Error(err)
	}
}
