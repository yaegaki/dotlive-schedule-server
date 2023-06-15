package handler

import (
	"testing"

	"firebase.google.com/go/messaging"
)

type notifyVideoTestClient struct {
	m *messaging.Message
}

func TestIsDotLiveScheduleText(t *testing.T) {
	tests := []struct {
		text     string
		expected bool
	}{
		{
			"どっとライブ予定表",
			true,
		},
		{
			"どっとライブ・ぶいぱい予定表",
			true,
		},
		{
			"どっとライブ・ほげ・ふが予定表",
			true,
		},
		{
			"どっとライブ",
			false,
		},
		{
			"予定表",
			false,
		},
		{
			"ほげほげ",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := isDotLiveScheduleText(tt.text)
			if got != tt.expected {
				t.Errorf("got: %v expect: %v", got, tt.expected)
			}
		})
	}
}
