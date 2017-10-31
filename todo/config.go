package todo

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/rameshg87/tools/todo/todolib"
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

	taskGroup, err := todolib.ListTaskGroupByName(groupName)
	if err == nil {
		return pmr("Task group already exists")
	}

	todolib.InsertTaskGroup(groupName, shortName)

	taskGroup, err = todolib.ListTaskGroupByName(groupName)
	if err == nil {
		log.Print("Inserted task group ", groupName,
			" with group id ", strconv.Itoa(taskGroup.GroupId))
	}
	return err
}

func PrintTaskGroups(c *cli.Context) error {
	tgs, err := todolib.ListTaskGroups()
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

func RemoveTaskGroup(c *cli.Context) error {
	if c.NArg() != 1 {
		return per(CommandArgumentError(1))
	}

	groupIdStr := c.Args().Get(0)
	groupId, err := strconv.Atoi(groupIdStr)
	if err != nil {
		return pmr(fmt.Sprintf("Invalid argument '%s'", groupIdStr))
	}

	err = todolib.RemoveTaskGroup(groupId)
	if err == nil {
		log.Println(fmt.Sprintf("Group %d removed successfully", groupId))
	}
	return err
}

func AddTask(c *cli.Context) error {
	var err error
	var task todolib.Task

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

	id, err := todolib.InsertTask(task)
	if err != nil {
		return per(err)
	}

	log.Print("Added task " + strconv.Itoa(id))
	return nil
}

func listTasksHelper(c *cli.Context, done int) error {
	tasks, err := todolib.ListTasks(done)
	if err != nil {
		return per(err)
	}
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)

	var headers []string
	headers = append(headers, "TaskId")
	if done == -1 {
		headers = append(headers, "X")
	}
	headers = append(headers, "Task", "Group", "Prio", "Est", "Act", "Due")
	table.SetHeader(headers)

	for _, task := range tasks {
		var fields []string
		fields = append(fields, strconv.Itoa(task.TaskId))
		if done == -1 {
			if task.Done {
				fields = append(fields, "X")
			} else {
				fields = append(fields, "")
			}
		}
		fields = append(
			fields, task.TaskStr, task.GroupName,
			strconv.Itoa(task.Priority), strconv.Itoa(task.EstMins),
			strconv.Itoa(task.ActMins), task.Due)
		table.Append(fields)
	}
	table.Render()
	log.Print(buf.String())
	return nil
}

func ListTasks(c *cli.Context) error {
	return listTasksHelper(c, 0)
}

func ListAllTasks(c *cli.Context) error {
	return listTasksHelper(c, -1)
}

func RemoveTask(c *cli.Context) error {
	if c.NArg() != 1 {
		return per(CommandArgumentError(1))
	}

	taskIdStr := c.Args().Get(0)
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		return per(InvalidArgument{taskIdStr, "task id"})
	}

	err = todolib.RemoveTask(taskId)
	if err == nil {
		log.Println(fmt.Sprintf("Task %d removed successfully", taskId))
	}

	return err
}

func doHelper(c *cli.Context) (todolib.Task, error) {
	var task todolib.Task
	if c.NArg() != 1 {
		return task, per(CommandArgumentError(1))
	}

	taskIdStr := c.Args().Get(0)
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		return task, per(InvalidArgument{taskIdStr, "task id"})
	}

	tasks, err := todolib.ListTask(taskId)
	if len(tasks) == 0 || err != nil {
		return task, per(TaskNotFound(taskId))
	}

	task = tasks[0]
	if task.Done {
		return task, pmr(fmt.Sprintf("Task %d is already marked as done", taskId))
	}

	task.Done = true
	err = todolib.UpdateTask(task)
	if err == nil {
		log.Printf("Marked task %d as done\n", taskId)
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
	id, err := todolib.InsertTask(task)
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
	taskGroup, err := todolib.ListTaskGroupByShortName(shortName)
	if err != nil {
		return per(InvalidArgument{shortName, "task group"})
	}

	taskIdStr := c.Args().Get(1)
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		return per(InvalidArgument{taskIdStr, "task id"})
	}

	tasks, err := todolib.ListTask(taskId)
	if len(tasks) == 0 || err != nil {
		return per(TaskNotFound(taskId))
	}

	task := tasks[0]
	task.TaskGroup = taskGroup
	err = todolib.UpdateTask(task)
	if err == nil {
		log.Printf("Changed task %d to group %s\n", taskId, taskGroup.GroupName)
	}

	return err
}

func InitConfigCommands(app *cli.App) {
	app.Commands = append(
		app.Commands,
		cli.Command{
			Name:   "ag",
			Usage:  "Add a new task group",
			Action: AddTaskGroup,
		},
		cli.Command{
			Name:   "lsg",
			Usage:  "List task groups",
			Action: PrintTaskGroups,
		},
		cli.Command{
			Name:   "rmg",
			Usage:  "Remove task group",
			Action: RemoveTaskGroup,
		},
		cli.Command{
			Name:   "a",
			Usage:  "Add a task",
			Action: AddTask,
		},
		cli.Command{
			Name:   "ls",
			Usage:  "List tasks",
			Action: ListTasks,
		},
		cli.Command{
			Name:   "lsa",
			Usage:  "List all tasks",
			Action: ListAllTasks,
		},
		cli.Command{
			Name:   "rm",
			Usage:  "Remove task",
			Action: RemoveTask,
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
	)
}
