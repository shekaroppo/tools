package todo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/rameshg87/tools/todo/todolib"
	"github.com/stretchr/testify/assert"
)

func createTestDb(t *testing.T) string {
	f, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	os.Setenv("TODODB", f.Name())
	return f.Name()
}

func getAppOutput(args []string) string {
	var buf bytes.Buffer
	log.SetFlags(0)
	log.SetOutput(&buf)
	app := GetApp()
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

func createTaskGroups(t *testing.T) {
	args := []string{"todo", "ag", "office project", "op"}
	expOutput := "Inserted task group office project with group id 1\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "ag", "office review", "or"}
	expOutput = "Inserted task group office review with group id 2\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "ag", "personal learning", "pl"}
	expOutput = "Inserted task group personal learning with group id 3\n"
	assertCommandOutput(t, args, expOutput)
}

func TestAddRemoveListTaskGroup(t *testing.T) {
	tempFile := createTestDb(t)
	defer os.Remove(tempFile)
	todolib.InitDb()

	createTaskGroups(t)

	args := []string{"todo", "ag", "office project"}
	expOutput := CommandArgumentError(2).Error() + "\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "ag", "office project", "op"}
	expOutput = "Task group already exists\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "lsg"}
	expOutput =
		`+---------+-------------------+-----------+
| GROUPID |     GROUPNAME     | SHORTNAME |
+---------+-------------------+-----------+
|       1 | office project    | op        |
|       2 | office review     | or        |
|       3 | personal learning | pl        |
+---------+-------------------+-----------+
`
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "rmg", "2"}
	expOutput = "Group 2 removed successfully\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "rmg"}
	expOutput = CommandArgumentError(1).Error() + "\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "lsg"}
	expOutput =
		`+---------+-------------------+-----------+
| GROUPID |     GROUPNAME     | SHORTNAME |
+---------+-------------------+-----------+
|       1 | office project    | op        |
|       3 | personal learning | pl        |
+---------+-------------------+-----------+
`
	assertCommandOutput(t, args, expOutput)
}

func createTasks(t *testing.T) {
	args := []string{"todo", "a", "op", "0", "60", "office project task 1"}
	expOutput := "Added task 1\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "a", "op", "1", "30", "office project task 2"}
	expOutput = "Added task 2\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "a", "pl", "1", "120", "personal learning task 1"}
	expOutput = "Added task 3\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "a", "pl", "0", "40", "personal learning task 2"}
	expOutput = "Added task 4\n"
	assertCommandOutput(t, args, expOutput)
}

func TestInsertRemoveListTasks(t *testing.T) {
	tempFile := createTestDb(t)
	defer os.Remove(tempFile)
	todolib.InitDb()

	createTaskGroups(t)
	createTasks(t)

	args := []string{"todo", "a"}
	expOutput := CommandArgumentError(4).Error() + "\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "a", "op", "foo", "60", "task 3"}
	expOutput = "Invalid argument 'foo' for priority\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "a", "op", "0", "foo", "task 3"}
	expOutput = "Invalid argument 'foo' for estimate in minutes\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "a", "op", "0", "30", ""}
	expOutput = "Empty task provided\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "a", "foo", "0", "30", "task 4"}
	expOutput = "No task group with short name 'foo'\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "ls"}
	expOutput =
		`+--------+--------------------------+-------------------+------+-----+-----+-----+
| TASKID |           TASK           |       GROUP       | PRIO | EST | ACT | DUE |
+--------+--------------------------+-------------------+------+-----+-----+-----+
|      1 | office project task 1    | office project    |    0 |  60 |   0 |     |
|      4 | personal learning task 2 | personal learning |    0 |  40 |   0 |     |
|      2 | office project task 2    | office project    |    1 |  30 |   0 |     |
|      3 | personal learning task 1 | personal learning |    1 | 120 |   0 |     |
+--------+--------------------------+-------------------+------+-----+-----+-----+
`
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "rm", "4"}
	expOutput = "Task 4 removed successfully\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "rm"}
	expOutput = CommandArgumentError(1).Error() + "\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "ls"}
	expOutput =
		`+--------+--------------------------+-------------------+------+-----+-----+-----+
| TASKID |           TASK           |       GROUP       | PRIO | EST | ACT | DUE |
+--------+--------------------------+-------------------+------+-----+-----+-----+
|      1 | office project task 1    | office project    |    0 |  60 |   0 |     |
|      2 | office project task 2    | office project    |    1 |  30 |   0 |     |
|      3 | personal learning task 1 | personal learning |    1 | 120 |   0 |     |
+--------+--------------------------+-------------------+------+-----+-----+-----+
`
	assertCommandOutput(t, args, expOutput)
}

func TestChangeFields(t *testing.T) {
	tempFile := createTestDb(t)
	defer os.Remove(tempFile)
	todolib.InitDb()

	createTaskGroups(t)
	createTasks(t)

	args := []string{"todo", "cg", "or"}
	expOutput := CommandArgumentError(2).Error() + "\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "cg", "qa", "1"}
	expOutput = InvalidArgument{"qa", "task group"}.Error() + "\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "cg", "or", "foo"}
	expOutput = InvalidArgument{"foo", "task id"}.Error() + "\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "cg", "or", "10"}
	expOutput = TaskNotFound(10).Error() + "\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "cg", "or", "1"}
	expOutput = "Changed task 1 to group office review\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "ls"}
	expOutput =
		`+--------+--------------------------+-------------------+------+-----+-----+-----+
| TASKID |           TASK           |       GROUP       | PRIO | EST | ACT | DUE |
+--------+--------------------------+-------------------+------+-----+-----+-----+
|      1 | office project task 1    | office review     |    0 |  60 |   0 |     |
|      4 | personal learning task 2 | personal learning |    0 |  40 |   0 |     |
|      2 | office project task 2    | office project    |    1 |  30 |   0 |     |
|      3 | personal learning task 1 | personal learning |    1 | 120 |   0 |     |
+--------+--------------------------+-------------------+------+-----+-----+-----+
`
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "do", "2"}
	expOutput = "Marked task 2 as done\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "do"}
	expOutput = CommandArgumentError(1).Error() + "\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "do", "foo"}
	expOutput = InvalidArgument{"foo", "task id"}.Error() + "\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "do", "10"}
	expOutput = TaskNotFound(10).Error() + "\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "do", "2"}
	expOutput = "Task 2 is already marked as done\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "ls"}
	expOutput =
		`+--------+--------------------------+-------------------+------+-----+-----+-----+
| TASKID |           TASK           |       GROUP       | PRIO | EST | ACT | DUE |
+--------+--------------------------+-------------------+------+-----+-----+-----+
|      1 | office project task 1    | office review     |    0 |  60 |   0 |     |
|      4 | personal learning task 2 | personal learning |    0 |  40 |   0 |     |
|      3 | personal learning task 1 | personal learning |    1 | 120 |   0 |     |
+--------+--------------------------+-------------------+------+-----+-----+-----+
`
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "lsa"}
	expOutput =
		`+--------+---+--------------------------+-------------------+------+-----+-----+-----+
| TASKID | X |           TASK           |       GROUP       | PRIO | EST | ACT | DUE |
+--------+---+--------------------------+-------------------+------+-----+-----+-----+
|      1 |   | office project task 1    | office review     |    0 |  60 |   0 |     |
|      4 |   | personal learning task 2 | personal learning |    0 |  40 |   0 |     |
|      2 | X | office project task 2    | office project    |    1 |  30 |   0 |     |
|      3 |   | personal learning task 1 | personal learning |    1 | 120 |   0 |     |
+--------+---+--------------------------+-------------------+------+-----+-----+-----+
`
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "car", "3"}
	expOutput = "Marked task 3 as done\nAdded task 5\n"
	assertCommandOutput(t, args, expOutput)

	args = []string{"todo", "lsa"}
	expOutput =
		`+--------+---+--------------------------+-------------------+------+-----+-----+-----+
| TASKID | X |           TASK           |       GROUP       | PRIO | EST | ACT | DUE |
+--------+---+--------------------------+-------------------+------+-----+-----+-----+
|      1 |   | office project task 1    | office review     |    0 |  60 |   0 |     |
|      4 |   | personal learning task 2 | personal learning |    0 |  40 |   0 |     |
|      2 | X | office project task 2    | office project    |    1 |  30 |   0 |     |
|      3 | X | personal learning task 1 | personal learning |    1 | 120 |   0 |     |
|      5 |   | personal learning task 1 | personal learning |    1 | 120 |   0 |     |
+--------+---+--------------------------+-------------------+------+-----+-----+-----+
`
	assertCommandOutput(t, args, expOutput)
}
