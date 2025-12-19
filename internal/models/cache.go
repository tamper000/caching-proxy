package models

import (
	"net/http"

	"github.com/bytedance/sonic"
)

type CacheEntry struct {
	Status int         `json:"status"`
	Header http.Header `json:"header"`
	Body   []byte      `json:"body"`
}

func (e *CacheEntry) MarshalBinary() ([]byte, error) {
	return sonic.Marshal(e)
}

func (e *CacheEntry) UnmarshalBinary(data []byte) error {
	return sonic.Unmarshal(data, e)
}
