package auth

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVerifyPassword_success(t *testing.T) {
	pw, s := "Hardware", "my_little_salty"
	hashed := "ee9a23000af19a22acd0d9a22dfe9558580771dc"
	assert.Equal(t, hashed, hash(pw, s))

	stored := "my_little_salty:" + hashed
	vp, err := VerifyPassword(pw, stored)
	require.NoError(t, err)
	assert.True(t, vp)

	vp, err = VerifyPassword("wrong password", stored)
	require.NoError(t, err)
	assert.False(t, vp)
}

func TestVerifyPassword_badStoredValue(t *testing.T) {
	noColonInThisString := "abcdefg"
	_, err := VerifyPassword("some_password", noColonInThisString)
	require.Error(t, err)
	require.Contains(t, "invalid hashed password", err.Error())
}
