package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// код писать тут

func SearchServer(w http.ResponseWriter, request *http.Request) {
	switch {
	case request.FormValue("query") == "timeout":
		time.Sleep(time.Second * 2)
	case request.Headerequest.Get("AccessToken") == "invalid_token":
		w.WriteHeader(http.StatusUnauthorized)
		return
	case request.FormValue("query") == "invalid_query":
		w.WriteHeader(http.StatusInternalServerError)
		return
	case request.FormValue("query") == "bad_request":
		w.WriteHeader(http.StatusBadRequest)
		return
	case request.FormValue("query") == "bad_request_invalid_json":
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid_json"))
		return
	case request.FormValue("query") == "invalid_json":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid_json"))
		return
	case request.FormValue("query") == "unknown_bad_request":
		resp, _ := json.Marshal(SearchErrorResponse{"UnknownBadReqeust"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	case request.FormValue("query") == "bad_order_field":
		resp, _ := json.Marshal(SearchErrorResponse{"ErrorBadOrderField"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	default:
		Users := []User{
			User{
				Id:     1,
				Name:   "Brooks Aguilar",
				Age:    1,
				About:  "test1",
				Gender: "male",
			},
			User{
				Id:     2,
				Name:   "Jennings Mays",
				Age:    2,
				About:  "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.",
				Gender: "male",
			},
		}
		resp, err := json.Marshal(Users)

		if err != nil {
			fmt.Println("cant pack result json:", err)
			return
		}

		w.Write(resp)
	}
}

func TestSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	client := SearchClient{
		URL: server.URL,
	}

	response, err := client.FindUsers(SearchRequest{Limit: 25, Offset: 1})

	if response == nil || err != nil {
		t.Errorf("Unexpected response, got error %v", err)
	}
}

func TestSuccess2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	client := SearchClient{
		URL: server.URL,
	}

	response, err := client.FindUsers(SearchRequest{Limit: 1, Offset: 1})

	if response == nil || err != nil {
		t.Errorf("Unexpected response, got error %v", err)
	}
}

func TestUnkownError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	client := SearchClient{
		URL: "http://unknown_error",
	}

	_, err := client.FindUsers(SearchRequest{Limit: 30})

	expectError("unknown error", err, t)
}

func TestInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	client := SearchClient{
		URL: server.URL,
	}

	_, err := client.FindUsers(SearchRequest{Query: "invalid_json"})

	expectError("cant unpack result json", err, t)
}

func TestBadRequestInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	client := SearchClient{
		URL: server.URL,
	}

	_, err := client.FindUsers(SearchRequest{Query: "bad_request_invalid_json"})

	expectError("cant unpack error json", err, t)
}

func TestBadOrderField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	client := SearchClient{
		URL: server.URL,
	}

	_, err := client.FindUsers(SearchRequest{Query: "bad_order_field"})

	expectError("OrderFeld", err, t)
}

func TestUnknownBadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	client := SearchClient{
		URL: server.URL,
	}

	_, err := client.FindUsers(SearchRequest{Query: "unknown_bad_request"})

	expectError("unknown bad request error", err, t)
}

func TestInternalServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	client := SearchClient{
		URL: server.URL,
	}

	_, err := client.FindUsers(SearchRequest{Query: "invalid_query"})

	expectError("SearchServer fatal error", err, t)
}

func TestIncorrectOffset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	client := SearchClient{
		URL: server.URL,
	}

	_, err := client.FindUsers(SearchRequest{Offset: -1})

	expectError("offset must be > 0", err, t)
}

func TestIncorrectLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	client := SearchClient{
		URL: server.URL,
	}

	_, err := client.FindUsers(SearchRequest{Limit: -1})

	expectError("limit must be > 0", err, t)
}

func TestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	client := SearchClient{
		URL: server.URL,
	}

	_, err := client.FindUsers(SearchRequest{Query: "timeout"})

	expectError("timeout for", err, t)
}

func TestInvalidToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	client := SearchClient{
		URL:         server.URL,
		AccessToken: "invalid_token",
	}

	_, err := client.FindUsers(SearchRequest{Limit: 1})

	expectError("Bad AccessToken", err, t)
}

func expectError(message string, err error, t *testing.T) {
	if err != nil && !strings.Contains(err.Error(), message) {
		t.Errorf("Unexpected error: %#v", err)
	}
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
