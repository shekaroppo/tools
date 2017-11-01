package todolib

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

func pmr(msg string) error {
	log.Print(msg)
	return errors.New(msg)
}

func per(err error) error {
	log.Print(err.Error())
	return err
}

func AddTaskGroup(c *cli.Context) error {
	if c.NArg() != 2 {
		return per(CommandArgumentError(2))
	}

	groupName := c.Args().Get(0)
	shortName := c.Args().Get(1)

	taskGroup, err := ListTaskGroupByName(groupName)
	if err == nil {
		return pmr("Task group already exists")
	}

	InsertTaskGroup(groupName, shortName)

	taskGroup, err = ListTaskGroupByName(groupName)
	if err == nil {
		log.Print("Inserted task group ", groupName,
			" with group id ", strconv.Itoa(taskGroup.GroupId))
	} else {
		log.Println(err)
	}

	return err
}

func PrintTaskGroups(c *cli.Context) error {
	tgs, err := ListTaskGroups()
	if err != nil {
		return per(err)
	}
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"groupId", "groupName", "shortName"})
	for _, tg := range tgs {
		table.Append([]string{strconv.Itoa(tg.GroupId), tg.GroupName, tg.ShortName})
	}
	table.Render()
	log.Print(buf.String())
	return nil
}

func RemoveTaskGroupCli(c *cli.Context) error {
	if c.NArg() != 1 {
		return per(CommandArgumentError(1))
	}

	groupIdStr := c.Args().Get(0)
	groupId, err := strconv.Atoi(groupIdStr)
	if err != nil {
		return per(InvalidArgument{groupIdStr, "group id"})
	}

	err = RemoveTaskGroup(groupId)
	if err == nil {
		log.Println(fmt.Sprintf("Group %d removed successfully", groupId))
	} else {
		log.Println(err)
	}
	return err
}

func AddTask(c *cli.Context) error {
	var err error
	var task Task

	if c.NArg() != 4 {
		return per(CommandArgumentError(4))
	}

	// <group> <priority> <est> <task>
	task.ShortName = c.Args().Get(0)

	priorityStr := c.Args().Get(1)
	task.Priority, err = strconv.Atoi(priorityStr)
	if err != nil {
		return per(InvalidArgument{priorityStr, "priority"})
	}

	estMinsStr := c.Args().Get(2)
	task.EstMins, err = strconv.Atoi(estMinsStr)
	if err != nil {
		return per(InvalidArgument{estMinsStr, "estimate in minutes"})
	}

	task.TaskStr = c.Args().Get(3)
	if task.TaskStr == "" {
		return pmr("Empty task provided")
	}

	id, err := InsertTask(task)
	if err != nil {
		return per(err)
	}

	log.Print("Added task " + strconv.Itoa(id))
	return nil
}

func RemoveTaskCli(c *cli.Context) error {
	if c.NArg() != 1 {
		return per(CommandArgumentError(1))
	}

	taskIdStr := c.Args().Get(0)
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		return per(InvalidArgument{taskIdStr, "task id"})
	}

	err = RemoveTask(taskId)
	if err == nil {
		log.Println(fmt.Sprintf("Task %d removed successfully", taskId))
	} else {
		log.Println(err)
	}

	return err
}

func doHelper(c *cli.Context) (Task, error) {
	var task Task
	if c.NArg() != 1 {
		return task, per(CommandArgumentError(1))
	}

	taskIdStr := c.Args().Get(0)
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		return task, per(InvalidArgument{taskIdStr, "task id"})
	}

	task, err = ListTask(taskId)
	if err != nil {
		return Task{}, per(err)
	}
	if task.Done {
		return task, pmr(fmt.Sprintf("Task %d is already marked as done", taskId))
	}

	task.Done = true
	err = UpdateTask(task)
	if err == nil {
		log.Printf("Marked task %d as done\n", taskId)
	} else {
		log.Println(err)
	}

	return task, err
}

func Do(c *cli.Context) error {
	_, err := doHelper(c)
	return err
}

func CloseAndReadd(c *cli.Context) error {
	task, err := doHelper(c)
	if err != nil {
		return err
	}
	task.Done = false
	id, err := InsertTask(task)
	if err != nil {
		return per(err)
	}
	log.Print("Added task " + strconv.Itoa(id))
	return nil
}

func ChangeGroup(c *cli.Context) error {
	if c.NArg() != 2 {
		err := CommandArgumentError(2)
		log.Print(err.Error())
		return err
	}

	shortName := c.Args().Get(0)
	taskGroup, err := ListTaskGroupByShortName(shortName)
	if err != nil {
		return per(err)
	}

	taskIdStr := c.Args().Get(1)
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		return per(InvalidArgument{taskIdStr, "task id"})
	}

	task, err := ListTask(taskId)
	if err != nil {
		return per(err)
	}

	task.TaskGroup = taskGroup
	err = UpdateTask(task)
	if err == nil {
		log.Printf("Changed task %d to group %s\n", taskId, taskGroup.GroupName)
	} else {
		log.Println(err)
	}

	return err
}

type TaskChangeFunc func(*Task, int) string

func ChangeIntField(
	c *cli.Context, changeFunc TaskChangeFunc, fieldNameStr string) error {
	if c.NArg() != 2 {
		err := CommandArgumentError(2)
		log.Print(err.Error())
		return err
	}

	fieldStr := c.Args().Get(0)
	field, err := strconv.Atoi(fieldStr)
	if err != nil {
		return per(InvalidArgument{fieldStr, fieldNameStr})
	}

	taskIdStr := c.Args().Get(1)
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		return per(InvalidArgument{taskIdStr, "task id"})
	}

	task, err := ListTask(taskId)
	if err != nil {
		return per(err)
	}

	msg := changeFunc(&task, field)
	err = UpdateTask(task)
	if err == nil {
		log.Printf(msg)
	} else {
		log.Println(err)
	}

	return err
}

func ChangePriority(c *cli.Context) error {
	changeFunc := func(task *Task, priority int) string {
		task.Priority = priority
		return fmt.Sprintf("Changed task %d to priority %d\n", task.TaskId, priority)
	}
	return ChangeIntField(c, changeFunc, "priority")
}

func ChangeEst(c *cli.Context) error {
	changeFunc := func(task *Task, estMins int) string {
		task.EstMins = estMins
		return fmt.Sprintf("Changed estimate for task %d to %d minutes\n", task.TaskId, estMins)
	}
	return ChangeIntField(c, changeFunc, "estimate in minutes")
}

func ChangeAct(c *cli.Context) error {
	changeFunc := func(task *Task, actMins int) string {
		task.ActMins = actMins
		return fmt.Sprintf("Changed actual for task %d to %d minutes\n", task.TaskId, actMins)
	}
	return ChangeIntField(c, changeFunc, "actual in minutes")
}

func InitDbCli(c *cli.Context) error {
	err := InitDb()
	if err != nil {
		return per(err)
	}
	fileName := os.Getenv("TODODB")
	log.Println("Initialized DB at " + fileName)
	return nil
}

func GetApp() *cli.App {
	log.SetFlags(0)
	app := cli.NewApp()
	InitConfigCommands(app)
	InitListCommands(app)
	return app
}

func InitConfigCommands(app *cli.App) {
	app.Commands = append(
		app.Commands,
		cli.Command{
			Name:   "init",
			Usage:  "Initialize a new DB",
			Action: InitDbCli,
		},
		cli.Command{
			Name:   "ag",
			Usage:  "Add a new task group",
			Action: AddTaskGroup,
		},
		cli.Command{
			Name:   "rmg",
			Usage:  "Remove task group",
			Action: RemoveTaskGroupCli,
		},
		cli.Command{
			Name:   "a",
			Usage:  "Add a task",
			Action: AddTask,
		},
		cli.Command{
			Name:   "rm",
			Usage:  "Remove task",
			Action: RemoveTaskCli,
		},
		cli.Command{
			Name:   "cg",
			Usage:  "Change group",
			Action: ChangeGroup,
		},
		cli.Command{
			Name:   "do",
			Usage:  "Complete a task",
			Action: Do,
		},
		cli.Command{
			Name:   "car",
			Usage:  "Close and readd a task",
			Action: CloseAndReadd,
		},
		cli.Command{
			Name:   "pri",
			Usage:  "Change priority",
			Action: ChangePriority,
		},
		cli.Command{
			Name:   "est",
			Usage:  "Change estimate",
			Action: ChangeEst,
		},
		cli.Command{
			Name:   "act",
			Usage:  "Change actual",
			Action: ChangeAct,
		},
	)
}
