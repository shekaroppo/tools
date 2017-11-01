package todolib

import (
	"bytes"
	"log"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

func listTasksHelper(c *cli.Context, done int) error {
	tasks, err := ListTasks(done)
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

func ListTasksCli(c *cli.Context) error {
	return listTasksHelper(c, 0)
}

func ListAllTasks(c *cli.Context) error {
	return listTasksHelper(c, -1)
}

func InitListCommands(app *cli.App) {
	app.Commands = append(
		app.Commands,
		cli.Command{
			Name:   "lsg",
			Usage:  "List task groups",
			Action: PrintTaskGroups,
		},
		cli.Command{
			Name:   "ls",
			Usage:  "List tasks",
			Action: ListTasksCli,
		},
		cli.Command{
			Name:   "lsa",
			Usage:  "List all tasks",
			Action: ListAllTasks,
		},
	)
}
