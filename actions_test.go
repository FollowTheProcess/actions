package actions_test

import (
	"testing"

	"github.com/FollowTheProcess/actions"
)

func TestHello(t *testing.T) {
	got := actions.Hello()
	want := "Hello actions"

	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}
