package fetcher

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/hi20160616/exhtml"
	"github.com/hi20160616/ms-trt/configs"
	"github.com/pkg/errors"
)

// pass test
func TestFetchArticle(t *testing.T) {
	tests := []struct {
		url string
		err error
	}{
		{"https://www.trt.com/realtime/china/story20210627-1161676", ErrTimeOverDays},
		{"https://www.trt.com/realtime/china/story20210623-1159750", ErrTimeOverDays},
		{"https://www.trt.com/realtime/china/story20210617-1157196", ErrTimeOverDays},
		{"https://www.trt.com/realtime/china/story20210621-1159132", ErrTimeOverDays},
	}
	for _, tc := range tests {
		a := NewArticle()
		a, err := a.fetchArticle(tc.url)
		if err != nil {
			if !errors.Is(err, ErrTimeOverDays) {
				t.Error(err)
			} else {
				fmt.Println("ignore old news pass test: ", tc.url)
			}
		} else {
			fmt.Println("pass test: ", a.Content)
		}
	}
}

func TestFetchTitle(t *testing.T) {
	tests := []struct {
		url   string
		title string
	}{
		{"https://www.trt.com/realtime/china/story20210627-1161676", "贺一诚将出席中共百年党庆"},
		{"https://www.trt.com/realtime/china/story20210623-1159750", "加拿大代表逾40国要求中国允许联合国官员进入新疆"},
		{"https://www.trt.com/realtime/world/story20210602-1151196", "马国男子腰缠巨蟒骑摩托车送往放生引热议"},
		{"https://www.trt.com/realtime/world/story20210607-1153241", "以色列将于14日前投票批准新政府"},
	}
	for _, tc := range tests {
		a := NewArticle()
		u, err := url.Parse(tc.url)
		if err != nil {
			t.Error(err)
		}
		a.U = u
		// Dail
		a.raw, a.doc, err = exhtml.GetRawAndDoc(a.U, timeout)
		if err != nil {
			t.Error(err)
		}
		got, err := a.fetchTitle()
		if err != nil {
			if !errors.Is(err, ErrTimeOverDays) {
				t.Error(err)
			} else {
				fmt.Println("ignore pass test: ", tc.url)
			}
		} else {
			if tc.title != got {
				t.Errorf("\nwant: %s\n got: %s", tc.title, got)
			}
		}
	}

}

func TestFetchUpdateTime(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{
			"https://www.trt.com/realtime/china/story20210621-1159132",
			"2021-06-21 19:15:15 +0800 UTC",
		},
		{
			"https://www.trt.com/realtime/china/story20210615-1156563",
			"2021-06-15 19:04:20 +0800 UTC",
		},
	}
	var err error
	if err := configs.Reset("../../"); err != nil {
		t.Error(err)
	}

	for _, tc := range tests {
		a := NewArticle()
		a.U, err = url.Parse(tc.url)
		if err != nil {
			t.Error(err)
		}
		// Dail
		a.raw, a.doc, err = exhtml.GetRawAndDoc(a.U, timeout)
		if err != nil {
			t.Error(err)
		}
		tt, err := a.fetchUpdateTime()
		if err != nil {
			t.Error(err)
		} else {
			ttt := tt.AsTime()
			got := shanghai(ttt)
			if got.String() != tc.want {
				t.Errorf("\nwant: %s\n got: %s", tc.want, got.String())
			}
		}
	}
}

func TestFetchContent(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{
			"https://www.trt.com/realtime/world/story20210602-1151196",
			"2021-06-02 15:44:33 +0800 UTC",
		},
		{
			"https://www.trt.com/realtime/world/story20210607-1153241",
			"2021-06-07 21:38:53 +0800 UTC",
		},
	}
	var err error
	if err := configs.Reset("../../"); err != nil {
		t.Error(err)
	}

	for _, tc := range tests {
		a := NewArticle()
		a.U, err = url.Parse(tc.url)
		if err != nil {
			t.Error(err)
		}
		// Dail
		a.raw, a.doc, err = exhtml.GetRawAndDoc(a.U, timeout)
		if err != nil {
			t.Error(err)
		}
		c, err := a.fetchContent()
		if err != nil {
			t.Error(err)
		}
		fmt.Println(c)
	}
}
