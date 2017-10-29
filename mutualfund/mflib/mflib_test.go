package mflib

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMoneyControlNavHelper(t *testing.T) {
	var mf MutualFund
	mf.mcUrl = "http://www.moneycontrol.com/mutual-funds/nav/dsp-br-money-manager-direct/MDS625"
	var mfs []MutualFund
	mfs = append(mfs, mf)
	mfToNav, err := moneyControlNavHelper(mfs)
	if err != nil && strings.Contains(err.Error(), "no such host") {
		return
	}
	assert.Nil(t, err)
	assert.Equal(t, len(mfToNav), 1)
	for _, nav := range mfToNav {
		assert.NotEqual(t, nav, 0)
	}
}
