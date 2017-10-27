package todolib

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddRemoveListTaskGroup(t *testing.T) {
	tempFile := createTestDb(t)
	defer os.Remove(tempFile)
	InitDb()

	AddTaskGroup(
		map[string]string{"groupName": "office project", "shortName": "op"})
	AddTaskGroup(
		map[string]string{"groupName": "office review", "shortName": "or"})
	AddTaskGroup(
		map[string]string{"groupName": "personal learning", "shortName": "pl"})

	expOutput :=
		`+---------+-------------------+-----------+
| GROUPID |     GROUPNAME     | SHORTNAME |
+---------+-------------------+-----------+
|       1 | office project    | op        |
|       2 | office review     | or        |
|       3 | personal learning | pl        |
+---------+-------------------+-----------+
`
	actOutput, err := GetTaskGroupStr(nil)
	assert.Nil(t, err)
	assert.Equal(t, expOutput, actOutput)

	_, err = DeleteTaskGroup(map[string]string{"groupId": "5"})
	assert.Nil(t, err)
	_, err = DeleteTaskGroup(map[string]string{"groupId": "2"})
	assert.Nil(t, err)

	actOutput, err = GetTaskGroupStr(nil)
	expOutput =
		`+---------+-------------------+-----------+
| GROUPID |     GROUPNAME     | SHORTNAME |
+---------+-------------------+-----------+
|       1 | office project    | op        |
|       3 | personal learning | pl        |
+---------+-------------------+-----------+
`
	assert.Nil(t, err)
	assert.Equal(t, expOutput, actOutput)
}
