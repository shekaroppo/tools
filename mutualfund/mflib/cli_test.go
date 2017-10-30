package mflib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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

func getAppOutput(args []string) string {
	var buf bytes.Buffer
	log.SetFlags(0)
	log.SetOutput(&buf)
	app := GetCliApp()
	app.Run(args)
	log.SetOutput(os.Stderr)
	return buf.String()
}

func p(t *testing.T, args []string, _ string) {
	fmt.Println(getAppOutput(args))
}

func assertCommandOutput(t *testing.T, args []string, expOutput string) {
	actOutput := getAppOutput(args)
	assert.Equal(t, expOutput, actOutput)
}

func createMutualFunds(t *testing.T) {
	args := []string{"mutualfund", "imf", "mf1", "url1", "folio1", "type1", "amfi1"}
	expOutput := "New mutual fund inserted successfully\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"mutualfund", "imf", "mf2", "url2", "folio2", "type2", "amfi2"}
	expOutput = "New mutual fund inserted successfully\n"
	assertCommandOutput(t, args, expOutput)
}

func TestInsertListRemoveMfCli(t *testing.T) {
	os.Unsetenv("MFDB")
	args := []string{"mutualfund", "init"}
	expOutput := "'MFDB' environment variable is not set\n"
	assertCommandOutput(t, args, expOutput)

	tempFile := createTestDb(t)
	defer os.Remove(tempFile)

	args = []string{"mutualfund", "init"}
	expOutput = "Initialized mutual fund database at " + tempFile + "\n"
	assertCommandOutput(t, args, expOutput)

	createMutualFunds(t)

	args = []string{"mutualfund", "lmf"}
	expOutput =
		`+------+------+--------+-------+------------+
| MFID | NAME | FOLIO  | TYPE  | SCHEMECODE |
+------+------+--------+-------+------------+
|    1 | mf1  | folio1 | type1 | amfi1      |
|    2 | mf2  | folio2 | type2 | amfi2      |
+------+------+--------+-------+------------+
`
	assertCommandOutput(t, args, expOutput)

	args = []string{"mutualfund", "rmf", "1"}
	err := CommandArgumentError(1)
	expOutput = "Mutual fund removed successfully\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"mutualfund", "lmf"}
	expOutput =
		`+------+------+--------+-------+------------+
| MFID | NAME | FOLIO  | TYPE  | SCHEMECODE |
+------+------+--------+-------+------------+
|    2 | mf2  | folio2 | type2 | amfi2      |
+------+------+--------+-------+------------+
`
	assertCommandOutput(t, args, expOutput)

	// Negative test cases
	args = []string{"mutualfund", "imf", "mf2", "url2", "folio2", "type2"}
	err = CommandArgumentError(5)
	expOutput = err.Error() + "\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"mutualfund", "rmf"}
	err = CommandArgumentError(1)
	expOutput = err.Error() + "\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"mutualfund", "rmf", "foo"}
	expOutput = "Invalid argument 'foo'\n"
	assertCommandOutput(t, args, expOutput)
}

func createMutualFundPurchases(t *testing.T) {
	args := []string{"mutualfund", "imfp", "1", "10000", "100", "2017-01-01"}
	expOutput := "New mutual fund purchase inserted successfully\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"mutualfund", "imfp", "1", "10000", "200", "2017-02-01"}
	expOutput = "New mutual fund purchase inserted successfully\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"mutualfund", "imfp", "2", "10000", "400", "2017-01-01"}
	expOutput = "New mutual fund purchase inserted successfully\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"mutualfund", "imfp", "2", "10000", "1000", "2017-02-01"}
	expOutput = "New mutual fund purchase inserted successfully\n"
	assertCommandOutput(t, args, expOutput)
}

func TestInsertListRemoveMfpCli(t *testing.T) {
	tempFile := createTestDb(t)
	defer os.Remove(tempFile)

	args := []string{"mutualfund", "init"}
	expOutput := "Initialized mutual fund database at " + tempFile + "\n"
	assertCommandOutput(t, args, expOutput)

	createMutualFunds(t)
	createMutualFundPurchases(t)

	args = []string{"mutualfund", "lmfp"}
	expOutput =
		`+-------+------+-------+-----------+----------+---------+------------+
| MFPID | NAME | TYPE  |  AMOUNT   |   NAV    |  UNITS  |    DATE    |
+-------+------+-------+-----------+----------+---------+------------+
|     1 | mf1  | type1 | 10000.000 |  100.000 | 100.000 | 2017-01-01 |
|     2 | mf1  | type1 | 10000.000 |  200.000 |  50.000 | 2017-02-01 |
|     3 | mf2  | type2 | 10000.000 |  400.000 |  25.000 | 2017-01-01 |
|     4 | mf2  | type2 | 10000.000 | 1000.000 |  10.000 | 2017-02-01 |
+-------+------+-------+-----------+----------+---------+------------+
`
	assertCommandOutput(t, args, expOutput)

	args = []string{"mutualfund", "rmfp", "3"}
	expOutput = "Mutual fund removed successfully\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"mutualfund", "lmfp"}
	expOutput =
		`+-------+------+-------+-----------+----------+---------+------------+
| MFPID | NAME | TYPE  |  AMOUNT   |   NAV    |  UNITS  |    DATE    |
+-------+------+-------+-----------+----------+---------+------------+
|     1 | mf1  | type1 | 10000.000 |  100.000 | 100.000 | 2017-01-01 |
|     2 | mf1  | type1 | 10000.000 |  200.000 |  50.000 | 2017-02-01 |
|     4 | mf2  | type2 | 10000.000 | 1000.000 |  10.000 | 2017-02-01 |
+-------+------+-------+-----------+----------+---------+------------+
`
	assertCommandOutput(t, args, expOutput)

	// Negative test cases
	args = []string{"mutualfund", "imfp", "foo", "10000", "100", "2017-01-01"}
	expOutput = "Invalid argument 'foo' for mutual fund id\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"mutualfund", "imfp", "5", "10000", "100", "2017-01-01"}
	expOutput = "Mutual fund with id=5 not found\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"mutualfund", "imfp", "1", "foo", "100", "2017-01-01"}
	expOutput = "Invalid argument 'foo' for amount\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"mutualfund", "imfp", "1", "10000", "foo", "2017-01-01"}
	expOutput = "Invalid argument 'foo' for NAV\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"mutualfund", "imfp", "1", "10000", "100", "01-01-2017"}
	expOutput = "Error parsing date '01-01-2017'. Please specify in format YYYY-MM-DD\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"mutualfund", "rmfp", "foo"}
	expOutput = "Invalid argument 'foo'\n"
	assertCommandOutput(t, args, expOutput)
}

func TestMutualFundSummary(t *testing.T) {
	tempFile := createTestDb(t)
	defer os.Remove(tempFile)

	args := []string{"mutualfund", "init"}
	expOutput := "Initialized mutual fund database at " + tempFile + "\n"
	assertCommandOutput(t, args, expOutput)

	createMutualFunds(t)
	createMutualFundPurchases(t)

	navHelper := func(mfs []MutualFund) (map[MutualFund]float64, error) {
		mfToNavMap := make(map[MutualFund]float64)
		for _, mf := range mfs {
			if mf.mfid == 1 {
				mfToNavMap[mf] = 500
			} else {
				mfToNavMap[mf] = 800
			}
		}
		return mfToNavMap, nil
	}

	nowHelper := func() time.Time {
		now, err := time.Parse("2006-01-02", "2017-03-01")
		assert.Nil(t, err)
		return now
	}

	NavHelperFunc = navHelper
	NowHelperFunc = nowHelper

	args = []string{"mutualfund", "smfs"}
	expOutput =
		`+------+-------+-------+---------+--------+--------+------------+--------+------------+
| MFID | NAME  | TYPE  | AVGDAYS | AMOUNT | UNITS  | CURRENTVAL |  APPR  |   PRJRET   |
+------+-------+-------+---------+--------+--------+------------+--------+------------+
|    2 | mf2   | type2 |      43 |  20000 |  35.00 |      28000 |  40.00 |    1639.36 |
|    1 | mf1   | type1 |      43 |  20000 | 150.00 |      75000 | 275.00 | 7457361.05 |
|      | Total |       |      43 |  40000 |        |     103000 | 157.50 |  306682.09 |
+------+-------+-------+---------+--------+--------+------------+--------+------------+
`
	assertCommandOutput(t, args, expOutput)
}

func TestDist(t *testing.T) {
	tempFile := createTestDb(t)
	defer os.Remove(tempFile)

	args := []string{"mutualfund", "init"}
	expOutput := "Initialized mutual fund database at " + tempFile + "\n"
	assertCommandOutput(t, args, expOutput)

	createMutualFunds(t)
	createMutualFundPurchases(t)
	args = []string{"mutualfund", "dis"}
	expOutput =
		`+-------+-----------+------------+
| TYPE  |  AMOUNT   | PERCENTAGE |
+-------+-----------+------------+
| type1 | 20000.000 |     50.000 |
| type2 | 20000.000 |     50.000 |
+-------+-----------+------------+
`
	assertCommandOutput(t, args, expOutput)
}
