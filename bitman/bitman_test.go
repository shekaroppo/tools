package bitman

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBitField(t *testing.T) {
	assert.Equal(t, "4", GetBitField("5004", 0, 2))
	assert.Equal(t, "0", GetBitField("5004", 0, 1))
	assert.Equal(t, "4", GetBitField("0x4b0c", 12, 15))
	assert.Equal(t, "176", GetBitField("0x4b0c", 4, 11))
	assert.Equal(t, "12", GetBitField("0x4b0c", 0, 3))
	assert.Equal(t, "17187", GetBitField("0x86460000000000000000000L", 77, 92))
	assert.Equal(t, "1", GetBitField("5004", 2, 2))
}

func TestSetBitField(t *testing.T) {
	assert.Equal(t, "0x235084", SetBitField("0x235004", 4, 7, "8"))
	assert.Equal(t, "0x4ffc", SetBitField("0x4b0c", 4, 11, "0x7f"))
	assert.Equal(t, "0xcb0c", SetBitField("0x4b0c", 15, 15, "1"))
	assert.Equal(t, "", SetBitField("0x4b0c", 15, 15, "5"))
}
