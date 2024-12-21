package tests

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestSendHeartbeat(t *testing.T) {
	client := resty.New()

	resp, err := client.R().
		Get("8088")

	require.NoError(t, err)

	t.Log(resp)
}
