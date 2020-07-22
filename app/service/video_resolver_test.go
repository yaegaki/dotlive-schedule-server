package service

import "testing"

func TestVideoResolverExcept(t *testing.T) {
	v := VideoResolver{}
	if !v.Except("https://www.youtube.com/channel/UCP9ZgeIJ3Ri9En69R0kJc9Q") {
		t.Errorf("Channel URL")
	}

	if v.Except("https://www.youtube.com/watch?v=bVsei7pIrbk") {
		t.Errorf("Youtube URL")
	}
}
