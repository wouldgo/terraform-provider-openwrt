package api

import (
	"fmt"
	"testing"
)

func TestLuciAuth(t *testing.T) {
	c := NewClient("http://192.168.8.1:8080")
	err := c.Auth("root", "admin")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := c.UCIGetAll("system")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(*raw))
}

func TestSystem(t *testing.T) {
	c := NewClient("http://192.168.8.1:8080")
	err := c.Auth("root", "admin")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := c.UCIGetSystem()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(raw)
}

func TestNTP(t *testing.T) {
	c := NewClient("http://192.168.8.1:8080")
	err := c.Auth("root", "admin")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := c.UCIGetAll("system", "ntp")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(raw)
}

func TestReadfile(t *testing.T) {
	c := NewClient("http://192.168.8.1:8080")
	err := c.Auth("root", "admin")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := c.ReadFile("/etc/config/system")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(raw))
}
