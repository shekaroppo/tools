package mflib

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createTestDb(t *testing.T) string {
	f, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	os.Setenv("MFDB", f.Name())
	return f.Name()
}

func TestInitDb(t *testing.T) {
	os.Unsetenv("MFDB")

	// Raises error when MFDB is not set.
	err := InitDb()
	assert.Equal(t, MfDbNotSet(true), err)

	// Initializes DB when MFDB is set.
	tempFile := createTestDb(t)
	defer os.Remove(tempFile)
	err = InitDb()
	assert.Nil(t, err)
}

func createMfs() []MutualFund {
	mf1 := MutualFund{1, "mf1", "url1", "folio1", "type1", "amfi1"}
	mf2 := MutualFund{2, "mf2", "url2", "folio2", "type2", "amfi2"}
	var mfs []MutualFund
	mfs = append(mfs, mf1, mf2)
	return mfs
}

func createMfps(t *testing.T, mfs []MutualFund) []MutualFundPurchase {
	var mfps []MutualFundPurchase
	var err error
	for i := 1; i < 5; i++ {
		var mfp MutualFundPurchase
		mfp.id = i
		mfp.MutualFund = mfs[i%len(mfs)]
		mfp.amount = 10000 * float64(i)
		mfp.nav = 100 * float64(i)
		mfp.time, err = time.Parse("2006-01-02", "2017-01-01")
		assert.Nil(t, err)
		// Add multiples 10 days
		mfp.time = mfp.time.Add(time.Duration(240*i) * time.Hour)
		mfps = append(mfps, mfp)
	}
	return mfps
}

func TestInsertListRemoveMf(t *testing.T) {
	mfs := createMfs()

	// All functions return error when MFDB is not set.
	os.Unsetenv("MFDB")
	err := InsertMutualFund(mfs[0])
	assert.Equal(t, MfDbNotSet(true), err)
	_, err = ListMutualFunds()
	assert.Equal(t, MfDbNotSet(true), err)
	err = RemoveMutualFund(mfs[0])
	assert.Equal(t, MfDbNotSet(true), err)

	tempFile := createTestDb(t)
	InitDb()
	defer os.Remove(tempFile)

	err = InsertMutualFund(mfs...)
	assert.Nil(t, err)

	retMfs, err := ListMutualFunds()
	assert.Nil(t, err)
	assert.Equal(t, mfs, retMfs)

	err = RemoveMutualFund(mfs[1])
	assert.Nil(t, err)

	retMfs, err = ListMutualFunds()
	assert.Equal(t, 1, len(retMfs))
	assert.Equal(t, mfs[0], retMfs[0])

	mf, err := GetMutualFunds(1)
	assert.Equal(t, mf, mfs[0])
}

func TestMutualFundHelper(t *testing.T) {
	tempFile := createTestDb(t)
	InitDb()
	defer os.Remove(tempFile)

	mf := MutualFund{1, "mf1", "url1", "folio1", "type1", "amfi1"}
	InsertMutualFundHelper(mf.name, mf.mcUrl, mf.folio, mf.mftype, mf.amfiSchemeCode)
	retMf, err := GetMutualFunds(1)
	assert.Nil(t, err)
	assert.Equal(t, mf, retMf)

	output, err := ListMutualFundHelper()
	assert.Nil(t, err)
	expOuptut :=
		`+------+------+--------+-------+------------+
| MFID | NAME | FOLIO  | TYPE  | SCHEMECODE |
+------+------+--------+-------+------------+
|    1 | mf1  | folio1 | type1 | amfi1      |
+------+------+--------+-------+------------+
`
	assert.Equal(t, expOuptut, output)

	err = RemoveMutualFundHelper(1)
	assert.Nil(t, err)
	retMfs, err := ListMutualFunds()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(retMfs))

}

func TestMutualFundPurchaseHelper(t *testing.T) {
	tempFile := createTestDb(t)
	InitDb()
	defer os.Remove(tempFile)

	mfs := createMfs()
	err := InsertMutualFund(mfs...)
	assert.Nil(t, err)

	err = InsertMutualFundPurchaseHelper(1, 1000.0, 100.0, "2017-01-01")
	assert.Nil(t, err)
	mfps, err := ListMutualFundPurchases()
	assert.Nil(t, err)
	assert.Equal(t, len(mfps), 1)

	output, err := ListMutualFundPurchaseHelper()
	assert.Nil(t, err)
	expOutput :=
		`+-------+------+-------+----------+---------+--------+------------+
| MFPID | NAME | TYPE  |  AMOUNT  |   NAV   | UNITS  |    DATE    |
+-------+------+-------+----------+---------+--------+------------+
|     1 | mf1  | type1 | 1000.000 | 100.000 | 10.000 | 2017-01-01 |
+-------+------+-------+----------+---------+--------+------------+
`
	assert.Equal(t, expOutput, output)

	err = RemoveMutualFundPurchaseHelper(1)
	assert.Nil(t, err)
	mfps, err = ListMutualFundPurchases()
	assert.Nil(t, err)
	assert.Equal(t, len(mfps), 0)
}

func TestInsertListRemoveMfp(t *testing.T) {
	var err error
	tempFile := createTestDb(t)
	InitDb()
	defer os.Remove(tempFile)

	mfs := createMfs()
	for _, mf := range mfs {
		err = InsertMutualFund(mf)
		assert.Nil(t, err)
	}

	mfps := createMfps(t, mfs)
	err = InsertMutualFundPurchase(mfps...)
	assert.Nil(t, err)

	retMfps, err := ListMutualFundPurchases()
	assert.Nil(t, err)
	assert.Equal(t, mfps, retMfps)

	err = RemoveMutualFundPurchase(mfps[3])
	assert.Nil(t, err)
	mfps = mfps[:3]
	retMfps, err = ListMutualFundPurchases()
	assert.Equal(t, mfps, retMfps)
}

func TestGetMutualFundSummary(t *testing.T) {
	var err error
	tempFile := createTestDb(t)
	InitDb()
	defer os.Remove(tempFile)

	mfs := createMfs()
	for _, mf := range mfs {
		err = InsertMutualFund(mf)
	}
	mfps := createMfps(t, mfs)
	InsertMutualFundPurchase(mfps...)

	navHelper := func(mfs []MutualFund) (map[MutualFund]float64, error) {
		mfToNavMap := make(map[MutualFund]float64)
		for _, mf := range mfs {
			if mf.mfid == 1 {
				mfToNavMap[mf] = 500
			} else {
				mfToNavMap[mf] = 600
			}
		}
		return mfToNavMap, nil
	}

	nowHelper := func() time.Time {
		now, err := time.Parse("2006-01-02", "2017-02-20")
		assert.Nil(t, err)
		return now
	}

	mfsums, err := GetMutualFundSummary(
		NavHelper(navHelper), NowHelper(nowHelper))
	assert.Nil(t, err)

	// (Amount, NAV, Diff days)
	// Total amount for mf1 = (20000, 200, 30), (40000, 400, 10)
	// Total amount for mf2 = (10000, 100, 40), (30000, 300, 20)
	// Weighted average for mf1 = 16
	// Weighted average for mf2 = 25
	// NAVs = 500, 600
	// Current Value = 100000, 120000
	for _, mfsum := range mfsums {
		if mfsum.mfid == 1 {
			assert.Equal(t, mfsum.MutualFund, mfs[0])
			assert.Equal(t, int(mfsum.avgDays), 16)
			assert.Equal(t, int(mfsum.amount), 60000)
			assert.Equal(t, int(mfsum.units), 200)
			assert.Equal(t, int(mfsum.currentValue), 100000)
			assert.Equal(t, int(mfsum.appreciation), 66)
		} else if mfsum.mfid == 2 {
			assert.Equal(t, mfsum.MutualFund, mfs[1])
			assert.Equal(t, int(mfsum.avgDays), 25)
			assert.Equal(t, int(mfsum.amount), 40000)
			assert.Equal(t, int(mfsum.units), 200)
			assert.Equal(t, int(mfsum.currentValue), 120000)
			assert.Equal(t, int(mfsum.appreciation), 200)
		}
	}

	output, err := MutualFundSummaryHelper(NavHelper(navHelper), NowHelper(nowHelper))
	assert.Nil(t, err)
	expOutput :=
		`+------+-------+-------+---------+------------+---------+------------+---------+---------------+
| MFID | NAME  | TYPE  | AVGDAYS |   AMOUNT   |  UNITS  | CURRENTVAL |  APPR   |    PRJRET     |
+------+-------+-------+---------+------------+---------+------------+---------+---------------+
|    1 | mf1   | type1 |      16 |  60000.000 | 200.000 | 100000.000 |  66.667 |  11505906.118 |
|    2 | mf2   | type2 |      25 |  40000.000 | 200.000 | 120000.000 | 200.000 | 924634879.227 |
|      | Total |       |      20 | 100000.000 |         | 220000.000 | 120.000 | 177506262.736 |
+------+-------+-------+---------+------------+---------+------------+---------+---------------+
`
	assert.Equal(t, expOutput, output)
}

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
