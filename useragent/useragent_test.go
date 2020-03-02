package useragent

import "testing"

func TestGetUserAgent(t *testing.T) {
	agent := GetUserAgent()

	expected := "nancy-client/build"
	if agent != expected {
		t.Errorf("User Agent not retrieved successfully, got %s, expected %s", agent, expected)
	}
}
