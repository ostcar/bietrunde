package model_test

import (
	"testing"
	"time"

	"github.com/ostcar/bietrunde/model"
	"github.com/ostcar/sticky"
)

func TestBieterCreate(t *testing.T) {
	now := func() time.Time { return time.Time{} }
	dbContent := sticky.NewMemoryDB(`
	{"time":"2023-10-20 18:15:58","type":"bieter-create","payload":{"id":252350}}
	`)
	_, err := sticky.New(dbContent, model.New(), model.GetEvent, sticky.WithNow[model.Model](now))
	if err != nil {
		t.Fatalf("sticky.New: %v", err)
	}
}
