package client

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

const baseUrl = "http://127.0.0.1:9800/"

func TestB(T *testing.T) {
	T.Parallel()
	for p := 0; p < 10; p++ {
		go func() {
			for {
				uid := rand.Intn(500)
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%sspin?uid=%d&appid=%d", baseUrl, uid, 1), nil)
				if err != nil {
					continue
				}
				_, resError := http.DefaultClient.Do(req)
				if resError != nil {
					continue
				}
				time.Sleep(time.Millisecond * 10)
			}
		}()
	}
	time.Sleep(time.Second * 30)
}
