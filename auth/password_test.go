package auth

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVerifyPassword_success(t *testing.T) {
	pw, s := "some_password", "some_salt"
	hashed := "2d063cf83d35313a2f65333b8bba12266a1c40c7"
	assert.Equal(t, hashed, hash(pw, s))

	stored := "some_salt:" + hashed
	vp, err := VerifyPassword(pw, stored)
	assert.NoError(t, err)
	assert.True(t, vp)

	vp, err = VerifyPassword("wrong password", stored)
	assert.NoError(t, err)
	assert.False(t, vp)
}

func TestVerifyPassword_badStoredValue(t *testing.T) {
	noColonInThisString := "abcdefg"
	_, err := VerifyPassword("some_password", noColonInThisString)
	assert.Error(t, err)
}
