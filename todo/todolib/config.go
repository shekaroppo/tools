package todolib

import (
	"bytes"
	"errors"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
)

func AddTaskGroup(args map[string]string) (string, error) {
	return "", InsertTaskGroup(args["groupName"], args["shortName"])
}

func DeleteTaskGroup(args map[string]string) (string, error) {
	groupId, err := strconv.Atoi(args["groupId"])
	if err != nil {
		return "", err
	}
	return "", RemoveTaskGroup(groupId)
}

func GetTaskGroupStr(args map[string]string) (string, error) {
	tgs, err := ListTaskGroups()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"groupId", "groupName", "shortName"})
	for _, tg := range tgs {
		table.Append([]string{strconv.Itoa(tg.GroupId), tg.GroupName, tg.ShortName})
	}
	table.Render()
	return buf.String(), nil
}

func AddTask(args map[string]string) (string, error) {
	var err error
	var task Task
	task.TaskStr = args["taskStr"]
	task.Added = time.Now().Format("2006-01-02")
	if args["due"] != "" {
		task.Due = args["due"]
	}

	estMinsStr, ok := args["estMins"]
	if !ok {
		return "", errors.New("Estimate not provided")
	}
	task.EstMins, err = strconv.Atoi(estMinsStr)
	if err != nil {
		return "", errors.New("Estimate not provided")
	}

	actMinsStr := args["actMins"]
	if actMinsStr != "" {
		task.ActMins, err = strconv.Atoi(estMinsStr)
		if err != nil {
			return "", err
		}
	}

	priorityStr, ok := args["priority"]
	if !ok {
		return "", errors.New("Priority not provided")
	}
	task.Priority, err = strconv.Atoi(priorityStr)
	if err != nil {
		return "", err
	}
	return "", nil
}
