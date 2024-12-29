package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPassword(t *testing.T) {
	// use not-so-secure params for faster test runs
	p := &params{
		memory:      16 * 1024,
		iterations:  1,
		parallelism: 1,
		saltLength:  16,
		keyLength:   32,
	}

	password := "passw0r#"
	encodedHash, err := generateFromPassword(password, p)
	require.NoError(t, err)

	match, err := ComparePasswordAndHash(password, encodedHash)
	require.NoError(t, err)
	require.True(t, match, "password should match")

	match, err = ComparePasswordAndHash("cats", encodedHash)
	require.NoError(t, err)
	require.False(t, match, "password should not match")
}
