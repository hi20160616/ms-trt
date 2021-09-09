package fetcher

import (
	"testing"

	"github.com/hi20160616/ms-trt/configs"
)

func TestFetch(t *testing.T) {
	if err := configs.Reset("../../"); err != nil {
		t.Error(err)
	}

	if err := Fetch(); err != nil {
		t.Error(err)
	}
}
