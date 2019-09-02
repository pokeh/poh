package main

import (
	"testing"
)

func TestExtractMessage(t *testing.T) {
	var tests = []struct {
		subject string
		text    string
		want    string
	}{
		{"With space", "<@UVW6ABCDE> ping", "ping"},
		{"Without space", "<@UVW6ABCDE>ping", "ping"},
		{"Upper cases and trimmable spaces", "<@UVW6ABCDE>  Ping ", "ping"},
		{"Full-width romaji and zenkaku spaces", "<@UVW6ABCDE>　ハローＷＯＲＬＤ", "ハローｗｏｒｌｄ"},
	}

	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			got := extractMessage(tt.text)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRespond(t *testing.T) {
	var tests = []struct {
		text string
		want string
	}{
		{"ping", "ぽん"},
		{"pingした", "pingしてえらい〜！"},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := respond(tt.text)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
