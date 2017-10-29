package bitmanlib

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

/*
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
*/

func getAppOutput(app *cli.App, args []string) string {
	var buf bytes.Buffer
	log.SetFlags(0)
	log.SetOutput(&buf)
	app.Run(args)
	log.SetOutput(os.Stderr)
	return buf.String()
}

func p(t *testing.T, app *cli.App, args []string, _ string) {
	fmt.Println(getAppOutput(app, args))
}

func assertCommandOutput(t *testing.T, app *cli.App, args []string, expOutput string) {
	actOutput := getAppOutput(app, args)
	assert.Equal(t, expOutput, actOutput)
}

func TestPrint(t *testing.T) {
	app := GetCliApp()

	args := []string{"bitman", "p", "5004", "0", "2"}
	expOutput := "4\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "p", "5004", "0", "1"}
	expOutput = "0\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "p", "0x4b0c", "12", "15"}
	expOutput = "4\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "p", "0x4b0c", "4", "11"}
	expOutput = "176\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "p", "0x4b0c", "0", "3"}
	expOutput = "12\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "p", "0x4b0c", "0", "3"}
	expOutput = "12\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "p", "0x86460000000000000000000", "77", "92"}
	expOutput = "17187\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "p", "5004", "2", "2"}
	expOutput = "1\n"
	assertCommandOutput(t, app, args, expOutput)

	// Negative test cases
	args = []string{"bitman", "p"}
	err := CommandArgumentError(3)
	expOutput = err.Error() + "\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "p", "foo", "2", "2"}
	expOutput = "Unable to parse 'foo' with base=10\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "p", "5004", "foo", "2"}
	expOutput = "Invalid start bit 'foo'\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "p", "5004", "2", "foo"}
	expOutput = "Invalid stop bit 'foo'\n"
	assertCommandOutput(t, app, args, expOutput)
}

func TestModify(t *testing.T) {
	app := GetCliApp()

	args := []string{"bitman", "m", "0x235004", "4", "7", "8"}
	expOutput := "0x235084\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "m", "0x4b0c", "4", "11", "0x7f"}
	expOutput = "0x4ffc\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "m", "0x4b0c", "15", "15", "1"}
	expOutput = "0xcb0c\n"
	assertCommandOutput(t, app, args, expOutput)

	// Negative test cases
	args = []string{"bitman", "m"}
	err := CommandArgumentError(4)
	expOutput = err.Error() + "\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "m", "foo", "2", "2", "1"}
	expOutput = "Unable to parse 'foo' with base=10\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "m", "5004", "foo", "2", "2"}
	expOutput = "Invalid start bit 'foo'\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "m", "5004", "2", "foo", "2"}
	expOutput = "Invalid stop bit 'foo'\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "m", "5004", "2", "2", "foo"}
	expOutput = "Unable to parse 'foo' with base=10\n"
	assertCommandOutput(t, app, args, expOutput)

	args = []string{"bitman", "m", "0x4b0c", "15", "15", "5"}
	expOutput = "Value '5' is greater than allowed value=1\n"
	assertCommandOutput(t, app, args, expOutput)
}
