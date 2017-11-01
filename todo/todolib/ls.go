package todolib

import (
	"bytes"
	"log"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

func listTasksHelper(c *cli.Context, done int,
	doneDate string, dueDate string) error {
	tasks, err := ListTasks(done, doneDate, dueDate)
	if err != nil {
		return per(err)
	}
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)

	var headers []string
	headers = append(headers, "TaskId")
	if done != 0 {
		headers = append(headers, "X")
	}
	headers = append(headers, "Task", "Group", "Prio", "Est", "Act", "Due")
	table.SetHeader(headers)

	for _, task := range tasks {
		var fields []string
		fields = append(fields, strconv.Itoa(task.TaskId))
		if done != 0 {
			if task.DoneDate != "" {
				fields = append(fields, "X")
			} else {
				fields = append(fields, "")
			}
		}
		fields = append(
			fields, task.TaskStr, task.GroupName,
			strconv.Itoa(task.Priority), strconv.Itoa(task.EstMins),
			strconv.Itoa(task.ActMins), task.DueDate)
		table.Append(fields)
	}
	table.Render()
	log.Print(buf.String())
	return nil
}

func ListTasksCli(c *cli.Context) error {
	return listTasksHelper(c, 0, "", "")
}

func ListAllTasks(c *cli.Context) error {
	return listTasksHelper(c, -1, "", "")
}

func SummaryToday(c *cli.Context) error {
	today := nowHelper().Format("2006-01-02")
	log.Println("Completed tasks:")
	listTasksHelper(c, 1, today, "")
	log.Println("\nTasks due today:")
	listTasksHelper(c, 0, "", today)
	return nil
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
		cli.Command{
			Name:   "st",
			Usage:  "Summary of today tasks",
			Action: SummaryToday,
		},
	)
}
