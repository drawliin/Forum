package tests

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"forum/internal/app"
	"forum/internal/config"
	"forum/internal/handlers"
)

func TestMain(t *testing.T) {
	cfg := &config.Config{
		Port:         "",
		DBPath:       ":memory:", // create db in RAM instead of disk
		CookieSecure: false,
	}
	err := app.New(cfg)
	if err != nil {
		t.Fatal("Failed to init")
	}

	cookies, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal("Failed to create client")
	}
	client := &http.Client{
		Jar: cookies,
	}

	testServer := httptest.NewServer(handlers.SetupRoutes())
	defer testServer.Close()

	// home
	if err := get(client, testServer.URL+"/"); err != nil {
		t.Fatal(err.Error())
	}

	// bad register
	badRegisterCases := []map[string][]string{
		{"email": {"badEmail"}, "username": {"a"}, "password": {"a"}},
		{"email": {"a@a"}, "username": {"a b"}, "password": {"a b"}},
		{"email": {"badEmail@"}, "username": {"a"}, "password": {"a"}},
		{"email": {"a@a"}, "username": {""}, "password": {"a"}},
		{"email": {"a@a"}, "username": {"a"}, "password": {""}},
	}
	for _, badCase := range badRegisterCases {
		if err := post(client, testServer.URL+"/register", badCase); err == nil {
			t.Fatal("")
		}
	}

	// bad login
	if err := post(client, testServer.URL+"/login", map[string][]string{
		"email":    {"aaa"},
		"password": {"bbb"},
	}); err == nil {
		t.Fatal("")
	}

	// good register
	if err := post(client, testServer.URL+"/register", map[string][]string{
		"email":    {"a@a"},
		"username": {"a"},
		"password": {"aaaaaa"},
	}); err != nil {
		t.Fatal(err)
	}

	// good login
	if err := post(client, testServer.URL+"/login", map[string][]string{
		"email":    {"a@a"},
		"password": {"aaaaaa"},
	}); err != nil {
		t.Fatal(err)
	}

	// bad filter/category
	badCatCases := []string{
		"filter=abcd",
		"category=12",
		"filter=aaaaaaa&category=0",
	}
	for _, badCase := range badCatCases {
		if err := get(client, testServer.URL+"/?"+badCase); err == nil {
			t.Fatal("")
		}
	}

	// good filter/category
	goodCatCases := []string{
		"filter=mine&category=2",
		"filter=liked&category=3",
	}
	for _, goodCase := range goodCatCases {
		if err := get(client, testServer.URL+"/?"+goodCase); err != nil {
			t.Fatal(err)
		}
	}

	// bad post
	badPostCases := []map[string][]string{
		{"title": {"no category"}, "content": {""}, "categories": {}},
		{"title": {""}, "content": {"no title"}, "categories": {"1"}},
		{"title": {strings.Repeat("a", 1000)}, "content": {"too long"}, "categories": {"1"}},
		{"title": {"bad category"}, "content": {"a"}, "categories": {"1", "2", "999999999999999"}},
		{"title": {"bad category"}, "content": {"a"}, "categories": {"1", "2", "-5"}},
		{"title": {"bad category"}, "content": {"a"}, "categories": {"1", "2", "abc"}},
	}
	for _, badCase := range badPostCases {
		if err := post(client, testServer.URL+"/post/new", badCase); err == nil {
			t.Fatal("")
		}
	}

	// good post
	if err := post(client, testServer.URL+"/post/new", map[string][]string{
		"title":      {"hello"},
		"content":    {"hello"},
		"categories": {"1", "2", "3", "4"},
	}); err != nil {
		t.Fatal(err)
	}
}

// get is a small test helper for requests that should return 200.
func get(client *http.Client, url string) error {
	resp, err := client.Get(url)
	if err != nil {
		return errors.New("Failed to get " + url)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Wrong status code %d", resp.StatusCode)
	}

	return nil
}

// post is a small test helper for form posts that should return 200.
func post(client *http.Client, url_ string, data map[string][]string) error {
	form := url.Values{}
	for k, values := range data {
		for _, v := range values {
			form.Add(k, v)
		}
	}

	resp, err := client.PostForm(url_, form)
	if err != nil {
		return errors.New("Failed to post " + url_)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Wrong status code %d", resp.StatusCode)
	}

	return nil
}
