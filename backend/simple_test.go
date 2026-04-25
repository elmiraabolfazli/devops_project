package main

import (
	"testing"
)

func TestSimple(t *testing.T) {

	t.Log("✅ Test passed successfully!")
}

func TestAnotherSimple(t *testing.T) {

	expected := 1
	actual := 1
	if expected != actual {
		t.Errorf("Expected %d, got %d", expected, actual)
	}
}
