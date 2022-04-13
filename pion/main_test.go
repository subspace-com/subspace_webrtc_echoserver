package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	port = "8888"
)

func TestPostOfferEmptyBody(t *testing.T) {
	req, err := http.NewRequest("POST", buildUrl("/offer"), nil)
	if err != nil {
		t.Fatal(err)
	}

	res := httptest.NewRecorder()

	offer(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("Response code was %v; want 200", res.Code)
	}
}

func TestPostOffer(t *testing.T) {

	body := []byte(`{"type":"offer","sdp":"v=0\r\no=- 5832113965204996707 2 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\na=group:BUNDLE 0\r\na=extmap-allow-mixed\r\na=msid-semantic: WMS\r\nm=application 9 UDP/DTLS/SCTP webrtc-datachannel\r\nc=IN IP4 0.0.0.0\r\na=ice-ufrag:UzE8\r\na=ice-pwd:XKdclC95XNQ5R0HXQuin2lSg\r\na=ice-options:trickle\r\na=fingerprint:sha-256 1F:45:64:D1:CB:E0:75:E5:99:50:14:FA:EA:52:21:33:A6:08:78:90:59:50:B6:E9:AB:2E:20:5A:67:19:A9:02\r\na=setup:actpass\r\na=mid:0\r\na=sctp-port:5000\r\na=max-message-size:262144\r\n"}`)

	req, err := http.NewRequest("POST", buildUrl("/offer"), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		t.Fatal(err)
	}

	res := httptest.NewRecorder()

	offer(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("Response code was %v; want 200", res.Code)
	}
}

func TestGetOffer(t *testing.T) {
	req, err := http.NewRequest("GET", buildUrl("/offer"), nil)
	if err != nil {
		t.Fatal(err)
	}

	res := httptest.NewRecorder()

	offer(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("Response code was %v; want 200", res.Code)
	}
}

func buildUrl(path string) string {
	return urlFor("http", port, path)
}

func urlFor(scheme string, serverPort string, path string) string {
	return scheme + "://localhost:" + serverPort + path
}
