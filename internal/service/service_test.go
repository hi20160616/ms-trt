package service

import (
	"context"
	"fmt"
	"testing"

	v1 "github.com/hi20160616/fetchnews-api/proto/v1"
)

func TestListArticles(t *testing.T) {
	s := &Server{}
	ss, err := s.ListArticles(context.Background(), &v1.ListArticlesRequest{})
	if err != nil {
		t.Error(err)
	}
	for _, e := range ss.Articles {
		fmt.Println(e)
	}
}
