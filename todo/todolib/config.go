package todolib

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

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

type TaskChangeFunc func(*Task, *cli.Context) (string, error)

func ChangeField(
	c *cli.Context, nargs int, changeFunc TaskChangeFunc) (Task, error) {
	if c.NArg() != nargs+1 {
		err := CommandArgumentError(nargs + 1)
		log.Print(err.Error())
		return Task{}, err
	}

	taskIdStr := c.Args().Get(nargs)
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		return Task{}, per(InvalidArgument{taskIdStr, "task id"})
	}

	task, err := ListTask(taskId)
	if err != nil {
		return Task{}, per(err)
	}

	msg, err := changeFunc(&task, c)
	if err != nil {
		return Task{}, per(err)
	}

	err = UpdateTask(task)
	if err == nil {
		log.Printf(msg)
	} else {
		log.Println(err)
	}

	return task, err
}

func ChangePriority(c *cli.Context) error {
	changeFunc := func(task *Task, c *cli.Context) (string, error) {
		priorityStr := c.Args().Get(0)
		priority, err := strconv.Atoi(priorityStr)
		if err != nil {
			return "", InvalidArgument{priorityStr, "priority"}
		}
		task.Priority = priority
		return fmt.Sprintf("Changed task %d to priority %d\n", task.TaskId, priority), nil
	}
	_, err := ChangeField(c, 1, changeFunc)
	return err
}

func ChangeEst(c *cli.Context) error {
	changeFunc := func(task *Task, c *cli.Context) (string, error) {
		estMinsStr := c.Args().Get(0)
		estMins, err := strconv.Atoi(estMinsStr)
		if err != nil {
			return "", InvalidArgument{estMinsStr, "estimate in minutes"}
		}
		task.EstMins = estMins
		return fmt.Sprintf("Changed estimate for task %d to %d minutes\n", task.TaskId, estMins), nil
	}
	_, err := ChangeField(c, 1, changeFunc)
	return err
}

func ChangeAct(c *cli.Context) error {
	changeFunc := func(task *Task, c *cli.Context) (string, error) {
		actMinsStr := c.Args().Get(0)
		actMins, err := strconv.Atoi(actMinsStr)
		if err != nil {
			return "", InvalidArgument{actMinsStr, "actual in minutes"}
		}
		task.ActMins = actMins
		return fmt.Sprintf("Changed actual for task %d to %d minutes\n", task.TaskId, actMins), nil
	}
	_, err := ChangeField(c, 1, changeFunc)
	return err
}

func ChangeGroup(c *cli.Context) error {
	changeFunc := func(task *Task, c *cli.Context) (string, error) {
		shortName := c.Args().Get(0)
		taskGroup, err := ListTaskGroupByShortName(shortName)
		if err != nil {
			return "", err
		}
		task.TaskGroup = taskGroup
		return fmt.Sprintf("Changed task %d to group %s\n", task.TaskId, task.GroupName), nil
	}
	_, err := ChangeField(c, 1, changeFunc)
	return err
}

func SetDueDate(c *cli.Context) error {
	changeFunc := func(task *Task, c *cli.Context) (string, error) {
		dueDateStr := c.Args().Get(0)
		_, err := time.Parse("2006-01-02", dueDateStr)
		if err != nil {
			dueInDays, err := strconv.Atoi(dueDateStr)
			if err != nil {
				return "", InvalidArgument{dueDateStr, "due date"}
			}
			dueDateStr = nowHelper().Add(
				time.Duration(time.Duration(dueInDays) * 24 * time.Hour)).Format("2006-01-02")
		}
		task.DueDate = dueDateStr
		return fmt.Sprintf("Set due date for task %d to %s\n", task.TaskId, dueDateStr), nil
	}
	_, err := ChangeField(c, 1, changeFunc)
	return err
}

func doHelper(c *cli.Context) (Task, error) {
	changeFunc := func(task *Task, c *cli.Context) (string, error) {
		if task.DoneDate != "" {
			return "", errors.New(fmt.Sprintf("Task %d is already marked as done", task.TaskId))
		}
		task.DoneDate = nowHelper().Format("2006-01-02")
		return fmt.Sprintf("Marked task %d as done\n", task.TaskId), nil
	}
	return ChangeField(c, 0, changeFunc)
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
	task.DoneDate = ""
	id, err := InsertTask(task)
	if err != nil {
		return per(err)
	}
	log.Print("Added task " + strconv.Itoa(id))
	return nil
}

func WaitForCtrlC() {
	var end_waiter sync.WaitGroup
	end_waiter.Add(1)
	var signal_channel chan os.Signal
	signal_channel = make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	go func() {
		<-signal_channel
		end_waiter.Done()
	}()
	end_waiter.Wait()
}

type DoNowWait func()
type NowHelper func() time.Time

var doNowWait DoNowWait = WaitForCtrlC
var nowHelper NowHelper = time.Now

func DoNow(c *cli.Context) error {
	changeFunc := func(task *Task, c *cli.Context) (string, error) {
		if task.DoneDate != "" {
			return "", errors.New(fmt.Sprintf("Task %d is already marked as done", task.TaskId))
		}
		start := nowHelper()
		doNowWait()
		stop := nowHelper()
		task.ActMins += int(stop.Sub(start) / time.Minute)
		return fmt.Sprintf("Changed actual for task %d to %d minutes\n", task.TaskId, task.ActMins), nil
	}
	_, err := ChangeField(c, 0, changeFunc)
	return err
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
		cli.Command{
			Name:   "donow",
			Usage:  "Start doing a task",
			Action: DoNow,
		},
		cli.Command{
			Name:   "sdd",
			Usage:  "Set due date for task",
			Action: SetDueDate,
		},
	)
}
