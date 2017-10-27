package todolib

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTestDb(t *testing.T) string {
	f, err := ioutil.TempFile("", "")
	assert.Nil(t, err)
	os.Setenv("TODODB", f.Name())
	return f.Name()
}

func TestInitDb(t *testing.T) {
	os.Unsetenv("TODODB")

	// Raises error when MFDB is not set.
	err := InitDb()
	assert.Equal(t, TodoDbNotSet(true), err)

	// Initializes DB when MFDB is set.
	tempFile := createTestDb(t)
	defer os.Remove(tempFile)
	err = InitDb()
	assert.Nil(t, err)
}

func createTaskGroups(t *testing.T) []TaskGroup {
	taskGroups := []TaskGroup{
		TaskGroup{1, "office project", "op"},
		TaskGroup{2, "personal learning", "pl"},
		TaskGroup{3, "office review", "or"},
	}
	for _, taskGroup := range taskGroups {
		InsertTaskGroup(taskGroup.GroupName, taskGroup.ShortName)
	}
	return taskGroups
}

func TestInsertRemoveGetTaskGroup(t *testing.T) {
	tempFile := createTestDb(t)
	defer os.Remove(tempFile)
	InitDb()
	expGroups := createTaskGroups(t)
	retGroups, err := ListTaskGroups()
	assert.Nil(t, err)
	assert.Equal(t, expGroups, retGroups)
	retGroup, err := ListTaskGroupByName("office project")
	assert.Equal(t, retGroup, expGroups[0])
	retGroup, err = ListTaskGroupByName("foo")
	assert.NotNil(t, err)
	err = RemoveTaskGroup(3)
	assert.Nil(t, err)
	expGroups = expGroups[:2]
	retGroups, err = ListTaskGroups()
	assert.Nil(t, err)
	assert.Equal(t, expGroups, retGroups)
}

func createTasks(t *testing.T, taskGroups []TaskGroup) []Task {
	tasks := []Task{
		Task{1, taskGroups[0], 1, "office project task 1", false, "2017-10-28", "2017-10-28", 10, 10, 0},
		Task{2, taskGroups[0], 1, "office project task 2", true, "2017-10-29", "2017-11-28", 10, 10, 0},
		Task{3, taskGroups[1], 2, "personal task 1", false, "2017-10-30", "2017-11-29", 10, 10, 0},
		Task{4, taskGroups[1], 2, "personal task 2", true, "2017-10-28", "2017-10-28", 10, 10, 0},
		Task{5, taskGroups[2], 3, "review task 1", false, "2017-10-28", "2017-10-28", 10, 10, 0},
	}
	for _, task := range tasks {
		InsertTask(task)
	}
	return tasks
}

func TestInsertRemoveUpdateTask(t *testing.T) {
	tempFile := createTestDb(t)
	defer os.Remove(tempFile)
	InitDb()
	expGroups := createTaskGroups(t)
	expTasks := createTasks(t, expGroups)
	retTasks, err := ListTasks()
	assert.Nil(t, err)
	assert.Equal(t, expTasks, retTasks)
	expTasks[0].TaskStr = "office project new task 1"
	expTasks[1].Done = false
	expTasks[2].TaskGroup = expGroups[2]
	expTasks[2].GroupId = 3
	expTasks[3].Due = "2017-12-31"
	expTasks[4].Added = "2017-12-30"
	for _, task := range expTasks {
		UpdateTask(task)
	}
	retTasks, err = ListTasks()
	assert.Nil(t, err)
	assert.Equal(t, expTasks, retTasks)
}
