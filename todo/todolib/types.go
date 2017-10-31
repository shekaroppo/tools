package todolib

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type TaskGroup struct {
	GroupId   int    `db:"group_id"`
	GroupName string `db:"group_name"`
	ShortName string `db:"short_name"`
}

type Task struct {
	TaskId    int `db:"task_id"`
	TaskGroup `db:"task_group"`
	GroupId   int    `db:"group_id"`
	TaskStr   string `db:"task_str"`
	Done      bool
	Added     string
	Due       string
	EstMins   int `db:"est_mins"`
	ActMins   int `db:"act_mins"`
	Priority  int
}

type Allocation struct {
	TaskGroup
	allocId   int
	allocMins int
}

type WeekAllocation struct {
	weekAllocId int
	beginDay    time.Time
	allocation  []Allocation
}

type TodoDbNotSet bool

func (m TodoDbNotSet) Error() string {
	return "'TODODB' environment variable is not set"
}

var dbSchema = [6]string{
	`create table task_group (
		group_id integer PRIMARY KEY AUTOINCREMENT,
		group_name text,
		short_name text)`,
	`create table task (
		task_id integer PRIMARY KEY AUTOINCREMENT,
		task_str integer,
		added text,
		due text,
		est_mins int,
		act_mins int,
		priority int,
		done int,
		group_id int,
		foreign key(group_id) references task_group(group_id))`,
	`create table allocation (
		alloc_id integer PRIMARY KEY AUTOINCREMENT,
		alloc_mins int,
		group_id int,
		foreign key(group_id) references task_group(group_id))`,
	`create table week (
		week_id integer PRIMARY KEY AUTOINCREMENT,
		week_begin_day timestamp)`,
	`create table week_allocation (
		week_alloc_id integer PRIMARY KEY AUTOINCREMENT,
		week_id integer,
		alloc_id integer,
		foreign key(alloc_id) references allocation(alloc_id),
		foreign key(week_id) references week(week_id))`,
}

func GetDb() (*sqlx.DB, error) {
	mfDb := os.Getenv("TODODB")
	if mfDb == "" {
		return nil, TodoDbNotSet(true)
	}

	db, err := sqlx.Open("sqlite3", mfDb)
	if err != nil {
		return nil, err
	}

	return db, err
}

func InitDb() error {
	db, err := GetDb()
	if err != nil {
		return err
	}
	defer db.Close()

	for _, createStmt := range dbSchema {
		db.MustExec(createStmt)
	}
	return nil
}

func InsertTaskGroup(groupName, shortName string) error {
	db, err := GetDb()
	if err != nil {
		return err
	}
	defer db.Close()

	tx := db.MustBegin()
	tx.MustExec(
		"insert into task_group (group_name, short_name) values ($1,$2)",
		groupName, shortName)
	tx.Commit()
	return nil
}

func ListTaskGroups() ([]TaskGroup, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	taskGroups := []TaskGroup{}
	err = db.Select(&taskGroups, "select * from task_group order by group_id")
	if err != nil {
		return nil, err
	}
	return taskGroups, nil
}

func ListTaskGroupByName(groupName string) (TaskGroup, error) {
	taskGroup := TaskGroup{}
	db, err := GetDb()
	if err != nil {
		return taskGroup, err
	}
	defer db.Close()
	err = db.Get(&taskGroup, "SELECT * FROM task_group WHERE group_name=$1", groupName)
	return taskGroup, err
}

func ListTaskGroupByShortName(shortName string) (TaskGroup, error) {
	taskGroup := TaskGroup{}
	db, err := GetDb()
	if err != nil {
		return taskGroup, err
	}
	defer db.Close()
	err = db.Get(&taskGroup, "SELECT * FROM task_group WHERE short_name=$1", shortName)
	return taskGroup, err
}

func RemoveTaskGroup(groupId int) error {
	db, err := GetDb()
	if err != nil {
		return err
	}
	defer db.Close()

	tx := db.MustBegin()
	tx.MustExec("delete from task_group where group_id=$1", groupId)
	tx.Commit()
	return nil
}

func InsertTask(task Task) (int, error) {
	db, err := GetDb()
	if err != nil {
		return -1, err
	}
	defer db.Close()

	taskGroup, err := ListTaskGroupByShortName(task.ShortName)
	if err != nil {
		msg := "No task group with short name '" + task.ShortName + "'"
		return -1, errors.New(msg)
	}

	tx := db.MustBegin()
	result := tx.MustExec(
		`insert into task (task_str, added, due, est_mins,
								 act_mins, priority, group_id, done)
		 values
		 ( $1, $2, $3, $4, $5, $6, $7, $8 )`,
		task.TaskStr, task.Added, task.Due, task.EstMins,
		task.ActMins, task.Priority, taskGroup.GroupId, task.Done)
	tx.Commit()
	insertId, err := result.LastInsertId()
	if err != nil {
		return -1, nil
	}
	return int(insertId), nil
}

func ListTasksHelper(done int, taskId int) ([]Task, error) {
	db, err := GetDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tasks := []Task{}
	query := `SELECT task.*, task_group.short_name "task_group.short_name",
						  task_group.group_name "task_group.group_name",
       				  task_group.group_id "task_group.group_id"
			    FROM task JOIN task_group ON task.group_id = task_group.group_id `

	var conditions []string
	if done == 0 || done == 1 {
		conditions = append(
			conditions, " task.done="+strconv.Itoa(done))
	}
	if taskId != -1 {
		conditions = append(
			conditions, " task.task_id="+strconv.Itoa(taskId))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER by task.priority ASC, task.task_id ASC"
	err = db.Select(&tasks, query)
	return tasks, err
}

func ListTasks(done int) ([]Task, error) {
	return ListTasksHelper(done, -1)
}

func ListTask(taskId int) ([]Task, error) {
	return ListTasksHelper(-1, taskId)
}

func RemoveTask(taskId int) error {
	db, err := GetDb()
	if err != nil {
		return err
	}
	defer db.Close()

	tx := db.MustBegin()
	tx.MustExec("delete from task where task_id=$1", taskId)
	tx.Commit()
	return nil
}

func UpdateTask(task Task) error {
	db, err := GetDb()
	if err != nil {
		return err
	}
	defer db.Close()

	taskGroup, err := ListTaskGroupByName(task.GroupName)
	if err != nil {
		return err
	}

	tx := db.MustBegin()
	tx.MustExec(
		`update task set
			task_str=$1, added=$2, due=$3, est_mins=$4,
			act_mins=$5, priority=$6, group_id=$7, done=$8
		 where task_id=$9`,
		task.TaskStr, task.Added, task.Due, task.EstMins,
		task.ActMins, task.Priority, taskGroup.GroupId, task.Done,
		task.TaskId)
	tx.Commit()
	return nil
}
