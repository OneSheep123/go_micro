package net

import (
	"fmt"
	"testing"
	"time"
)

func TestClient_Send(t *testing.T) {
	go func() {
		s := &Serve{}
		err := s.Start("tcp", ":8081")
		if err != nil {
			t.Log(err)
		}
	}()
	time.Sleep(3 * time.Second)
	client := &Client{
		address: "localhost:8081",
		network: "tcp",
	}
	res, err := client.Send("nihaoya")
	if err != nil {
		t.Log(err)
	}
	fmt.Println(res)
}
