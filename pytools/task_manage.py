#!/usr/bin/python
"""Python executable for all my todo.sh related tasks."""

import collections
import datetime
import imp
import os
import re
import signal
import subprocess
import sys
import time

import argparse
import termcolor

GOOGLE_DRIVE = os.getenv("GOOGLE_DRIVE")
FIXME_FILE = os.path.join(GOOGLE_DRIVE, "todo/todo.txt")
DONE_FILE = os.path.join(GOOGLE_DRIVE, "todo/done.txt")
LOCK_FILE = os.path.join(GOOGLE_DRIVE, "todo/todo_lock")
ALLOCATIONS_FILE = os.path.join(GOOGLE_DRIVE, "todo/allocations.conf")
WP_ALLOC_FILE = os.path.join(
    os.environ['HOME'], ".todo.actions.d/wp/allocations.conf")
PLUGIN_PATHS = os.environ.get('TASK_MANAGE_PLUGIN_PATHS', '')
PROJECT_MAP = {
    '': '',
    'op':  '+office +project',
    'om':  '+office +maintenance',
    'or':  '+office +reviews',
    'ol':  '+office +learning',
    'pl':  '+personal +learning',
    'pr':  '+personal +reading',
    'ppr': '+personal +project',
    'pf':  '+personal +fitness',
}
REVERSE_PROJECT_MAP = {y: x for x, y in PROJECT_MAP.items()}
COLOR_MAP = {
    "A": "yellow",
    "B": "green",
    "C": "cyan",
    "D": "magenta",
    "E": "blue",
    "F": "yellow",
    "G": "green",
    "H": "blue",
    "" : "white",
}
QUICK_ADD_TASKS = {
    1: '(A) badminton +personal +fitness due:today est:120',
}
TODAY_STR = " due:%s" % str(datetime.date.today())

PLUGINS = []
for plugin_path in PLUGIN_PATHS.split(','):
    PLUGINS.append(imp.load_source('', plugin_path))

def projects_in_alloc():
    allocations = {}
    with open(ALLOCATIONS_FILE, "r") as fobj:
        for line in fobj.readlines():
            project, time_mins = line.strip().split(":")
            allocations[project] = int(time_mins)
    return allocations

def print_task(task):
    print "%s %s" % (task[0] + 1, task[1])

def projects_without_nil():
    keys = PROJECT_MAP.keys()
    keys.remove('')
    return keys

def priorities():
    prio_list = ["A", "B", "C", "D"]
    return [x.lower() for x in prio_list] + prio_list

def todo():
    """Returns stripped off contents of todo file."""
    with open(FIXME_FILE, "r") as fobj:
        return [(i, x.strip()) for i, x in enumerate(fobj.readlines())]

def done():
    """Returns stripped off contents of done file."""
    with open(DONE_FILE, "r") as fobj:
        return [(i, x.strip()) for i, x in enumerate(fobj.readlines())]

def get_priority(task_str):
    match = re.search(r"\(([A-Za-z])\) ", task_str)
    if match:
        return match.group(1)
    return None

def remove_priority(task_str):
    return re.sub(r"\([A-Za-z]\) ", "", task_str)

def add_priority_task_str(task_str, priority):
    task_str = remove_priority(task_str)
    return "(%s) %s" % (priority.upper(), task_str)

def add_priority(task_obj, priority):
    return (task_obj[0], add_priority_task_str(task_obj[1], priority))

def done_task(task):
    """Returns the format of completed task as in done file."""
    task_str = remove_priority(task[1])
    return (task[0], "x %s %s" % (datetime.date.today(), task_str))

def remove_min(task):
    """Returns the task string with 'min:xx' removed."""
    return (task[0], re.sub(r"\s*min:\d+", "", task[1]))

def remove_due_today(task):
    return (task[0], re.sub(r"\s*due:\d+-\d+-\d+", "", task[1]))

def remove_due_today_and_min(task):
    return remove_due_today(remove_min(task))

def remove_done(task):
    """Returns the done string removed."""
    return (task[0], re.sub(r"^x \d{4}-\d{2}-\d{2} ", "", task[1]))

def sanitized_task(task_str):
    return re.sub(r'\s+', ' ', task_str)

def flush_todo(todo_contents):
    """Flush the contents of todo into todo.txt file."""
    with open(FIXME_FILE, "w") as fobj:
        fobj.write("\n".join([sanitized_task(x[1]) for x in todo_contents]))

def flush_done(done_contents):
    """Flush the contents of done into done.txt file."""
    with open(DONE_FILE, "w") as fobj:
        fobj.write("\n".join([x[1] for x in done_contents]))

def today():
    """Returns the current date in yyyy-mm-dd."""
    return str(datetime.date.today())

def t_car(parsed_args):
    """Adds the tasks provided to done file and remove min: from todo."""
    todo_contents = todo()
    done_contents = done()
    for task in parsed_args.tasks:
        task = task - 1
        task_obj = todo_contents[task]
        done_contents.append(done_task(task_obj))
        todo_contents[task] = remove_min(task_obj)
    flush_done(done_contents)
    flush_todo(todo_contents)

def t_do(parsed_args):
    """Completes the tasks provided to done file and remove from todo."""
    todo_contents = todo()
    done_contents = done()
    for task in sorted(parsed_args.tasks, reverse=True):
        task = task - 1
        task_obj = todo_contents[task]
        print "Completing task '%s'" % task_obj[1]
        if 'min:' not in task_obj[1]:
            sys.stdout.write("Task has no 'min' set. Complete this task ? ")
            resp = raw_input()
            if resp.lower() != 'y':
                print "Skipped marking task %d as done" % (task+1)
                continue
        print "Marked task %d as done" % (task+1)
        done_contents.append(done_task(task_obj))
        del todo_contents[task]
    flush_done(done_contents)
    flush_todo(todo_contents)

def t_pri(parsed_args):
    """Assigns priority for tasks."""
    todo_contents = todo()
    priority = parsed_args.priority.upper()
    for task in parsed_args.tasks:
        task = task - 1
        task_obj = todo_contents[task]
        task_str = remove_priority(task_obj[1])
        task_str = "(%s) " % priority + task_str
        print "Task %s changed to '%s'" % (task_obj[0]+1, task_str)
        todo_contents[task] = (task_obj[0], task_str)
    flush_todo(todo_contents)

def t_depri(parsed_args):
    """Assigns priority for tasks."""
    todo_contents = todo()
    for task in parsed_args.tasks:
        task = task - 1
        task_obj = todo_contents[task]
        task_str = remove_priority(task_obj[1])
        print "Task %s changed to '%s'" % (task_obj[0]+1, task_str)
        todo_contents[task] = (task_obj[0], task_str)
    flush_todo(todo_contents)

def t_dng(parsed_args):
    """Prints all the tasks in done.txt that have the given search string."""
    done_contents = done()
    matched = []
    task_str_lower = parsed_args.task_str.lower()
    for task in done_contents:
        if task_str_lower in task[1].lower():
            matched.append(task)
    for task in matched:
        print "%s %s" % (task[0]+1, task[1])

def colored(task_str):
    priority = ""
    match = re.search(r"\(([A-Za-z])\)", task_str)
    if match:
        priority = match.group(1)
    return termcolor.colored(task_str, COLOR_MAP.get(priority, "white"))

def print_todo_tasks(parsed_args, tasks):
    """Prints the given task according to the correct format."""
    def compare(task1, task2):
        """Compare function for two tasks."""
        prio1 = 'Z'
        prio2 = 'Z'
        match1 = re.search(r"\(([A-Za-z])\)", task1[1])
        match2 = re.search(r"\(([A-Za-z])\)", task2[1])
        if match1:
            prio1 = match1.group(1).upper()
        if match2:
            prio2 = match2.group(1).upper()
        if (ord(prio1) - ord(prio2)) != 0:
            return ord(prio1) - ord(prio2)
        else:
            return task1[0] - task2[0]

    def _processed_task_str(task_str):
        if not(hasattr(parsed_args, 'added') and parsed_args.added):
            task_str = re.sub(r'\sadded:\d{4}-\d{2}-\d{2}', '', task_str)
        return task_str

    print "\n".join([colored("%02d  %s" % (x[0]+1, _processed_task_str(x[1])))
                     for x in sorted(tasks, cmp=compare)])

def print_done_tasks(tasks):
    """Prints the done task according to the correct format."""
    print "\n".join(
        ["%02d  %s" % (x[0]+1, x[1])
         for x in sorted(tasks, key=lambda y: y[0])])

def min_time(task_str):
    match = re.search(r"min:(\d+)", task_str)
    if match:
        return match.group(1)
    return None

def est_time(task_str):
    match = re.search(r"est:(\d+)", task_str)
    if match:
        return match.group(1)
    return None

def min_or_est(task_str, consider_min=None):
    min_str = min_time(task_str)
    est_str = est_time(task_str)
    final = est_str
    if min_str and (consider_min or int(min_str) > int(est_str)):
        final = min_str
    return final

def time_str(time_min):
    return "%d min (%.2f hrs)" % (time_min, time_min/60.0)

def time_for_tasks(tasks, consider_min=None):
    time_sum = 0
    for task in tasks:
        time_for_task = min_or_est(task[1], consider_min=consider_min)
        time_sum += int(time_for_task)
    return time_sum

def time_taken(tasks, consider_min=None):
    """Returns the string for the time taken for all the tasks."""
    time_sum = time_for_tasks(tasks, consider_min)
    return time_str(time_sum)

def t_dna(parsed_args):
    """Adds the tasks provided in done file back to todo file."""
    todo_contents = todo()
    done_contents = done()
    task = parsed_args.task
    task = task - 1
    task_obj = remove_due_today_and_min(done_contents[task])
    task_obj = remove_done(task_obj)
    task_obj = add_priority(task_obj, parsed_args.priority)
    todo_contents.append(task_obj)
    print "Added task '%s %s'" % (len(todo_contents), task_obj[1])
    flush_todo(todo_contents)
    todo_contents = todo()
    return todo_contents

def t_min(parsed_args):
    """Set min for a task."""
    if parsed_args.done:
        done_contents = done()
        task = parsed_args.task - 1
        task_obj = done_contents[task]
        task_str = re.sub(r"\s*min:\d+", "", task_obj[1])
        assert parsed_args.min > 0
        task_str = task_str + " min:%s" % parsed_args.min
        done_contents[task] = (task_obj[0], task_str)
        flush_done(done_contents)
    else:
        todo_contents = todo()
        task = parsed_args.task - 1
        task_obj = todo_contents[task]
        task_str = re.sub(r"\s*min:\d+", "", task_obj[1])
        if int(parsed_args.min):
            task_str = task_str + " min:%s" % parsed_args.min
        todo_contents[task] = (task_obj[0], task_str)
        flush_todo(todo_contents)

def t_est(parsed_args):
    """Set est for a task."""
    todo_contents = todo()
    task = parsed_args.task - 1
    task_obj = todo_contents[task]
    assert int(parsed_args.est) > 0
    task_str = re.sub(r"\s*est:\d+", "", task_obj[1])
    task_str = task_str + " est:%s" % parsed_args.est
    todo_contents[task] = (task_obj[0], task_str)
    flush_todo(todo_contents)

def t_donow(parsed_args):
    """Start doing the task."""
    todo_contents = todo()
    task = parsed_args.task - 1
    task_str = todo_contents[task][1].lower()
    process = subprocess.Popen(["ps", "-eaf"], stdout=subprocess.PIPE)
    stdout, _ = process.communicate()
    lines_with_donow = [x for x in stdout.split("\n") if 'donow' in x]
    if len(lines_with_donow) > 1:
        print "Error: donow is already running from some other terminal !!"
        return
    print "Doing task '%s %s'" % (todo_contents[task][0]+1, task_str)
    time_start = 0
    time_elapsed = 0
    abs_time_start = datetime.datetime.now()
    match = re.search(r"min:(\d+)", task_str)
    if match:
        time_start = int(match.group(1))
    def signal_handler(signal_no, frame):
        """Signal handler for do_now"""
        _ = signal_no
        _ = frame
        todo_contents = todo()
        new_task = [x for x in todo_contents if x[1].lower() == task_str][0]
        assert new_task
        time_elapsed_min1 = time_elapsed / 60
        time_elapsed_min2 = (
            (datetime.datetime.now() - abs_time_start).seconds / 60)
        if abs(time_elapsed_min2 - time_elapsed_min1) > 2:
            resp = ''
            while resp == '1' or resp == '2':
                print ("Discrepancy between absoulte time difference and "
                       "elapsed time counter.")
                print "1) Elapsed time : %s mins" % time_elapsed_min2
                print "2) Absolute time : %s mins" % time_elapsed_min1
                print "Enter which one to take:"
                resp = raw_input()
            if resp == '1':
                time_elapsed_min = time_elapsed_min1
            else:
                time_elapsed_min = time_elapsed_min2
        else:
            time_elapsed_min = time_elapsed_min1
        if time_elapsed_min:
            new_task = remove_min(new_task)
            new_task = (new_task[0], new_task[1]+" min:%s" %
                        (time_start+time_elapsed_min))
            print "Updating %s to '%s'" % (new_task[0]+1, new_task[1])
            todo_contents[new_task[0]] = new_task
            flush_todo(todo_contents)
        else:
            print "Not updating as time spent is < 1 min."
        sys.exit(0)

    signal.signal(signal.SIGINT, signal_handler)
    print 'Press Ctrl+C when done'
    while True:
        time.sleep(30)
        time_elapsed += 30
        if time_elapsed % 300 == 0:
            print "%s minutes elapsed" % (time_elapsed / 60)

def t_dndonow(parsed_args):
    """Adds the tasks provided in done file back to todo file
    and does do now."""
    todo_contents = t_dna(parsed_args)
    parsed_args.task = len(todo_contents)
    t_donow(parsed_args)

def t_dnt(parsed_args):
    """Does tail of done.txt."""
    done_contents = done()
    print_done_tasks(done_contents[-parsed_args.number_of_lines:])

def t_ls(parsed_args, todo_contents=None):
    """Does the listing of all tasks in project."""
    project = PROJECT_MAP[parsed_args.project_short_name]
    if todo_contents is None:
        todo_contents = todo()
    priority = ("(%s)" % parsed_args.priority.upper()
                if parsed_args.priority else '')
    project_filtered_tasks = [x for x in todo_contents if project in x[1]]
    priority_filtered_tasks = [x for x in project_filtered_tasks
                               if priority in x[1]]
    if parsed_args.numonly:
        for task in priority_filtered_tasks:
            print task[0] + 1
    else:
        print_todo_tasks(parsed_args, priority_filtered_tasks)
        print "Total time : %s" % (time_taken(priority_filtered_tasks))

def t_g(parsed_args):
    """Prints the graph for the given project."""
    project = PROJECT_MAP.get(parsed_args.project_short_name, "")
    command = "todo.sh wp graph \"%s\"" % project
    os.system(command)

def t_s(parsed_args):
    """Prints the summary for the given project."""
    project = PROJECT_MAP.get(parsed_args.project_short_name, "")
    command = "todo.sh wp summary \"%s\"" % project
    os.system(command)

def add(todo_contents, task_str, project_short_name, priority=None):
    project = PROJECT_MAP.get(project_short_name, "")
    assert project
    assert "est:" in task_str
    today_str = str(datetime.date.today())
    task_str = "%s added:%s %s" % (task_str, today_str, project)
    if priority:
        task_str = "(%s) " % priority.upper() + task_str
    print "Added task '%s %s'" % (len(todo_contents)+1, task_str)
    todo_contents.append((len(todo_contents), task_str))

def _get_due_str(due_on):
    try:
        num_days = int(due_on)
        due_on = str(datetime.date.today() + datetime.timedelta(days=num_days))
    except ValueError:
        assert re.match(r'\d{4}-\d{2}-\d{2}', due_on)
    return ' due:' + due_on

def t_a(parsed_args):
    """Add the given task to todo.txt."""
    todo_contents = todo()
    task_str = parsed_args.task_str
    if parsed_args.dt:
        task_str += TODAY_STR
    elif parsed_args.due_on:
        task_str += _get_due_str(parsed_args.due_on)
    add(todo_contents, task_str, parsed_args.project_short_name,
        parsed_args.priority)
    flush_todo(todo_contents)
    return todo_contents

def t_sdd(parsed_args):
    todo_contents = todo()
    due_on = parsed_args.due_on
    for task in parsed_args.tasks:
        task = task - 1
        task_obj = todo_contents[task]
        task_obj = remove_due_today(task_obj)
        task_obj = (task_obj[0], task_obj[1] + _get_due_str(due_on))
        todo_contents[task] = task_obj
    flush_todo(todo_contents)

def t_cp(parsed_args):
    todo_contents = todo()
    for task in parsed_args.tasks:
        task = task - 1
        old_task_str = todo_contents[task][1]
        project = PROJECT_MAP[parsed_args.project_short_name]
        task_str = re.sub(r"(\+\w+)\s+(\+\w+)*", project, old_task_str)
        print "Changing '%s' to '%s'" % (old_task_str, task_str)
        todo_contents[task] = (todo_contents[task][0], task_str)
    flush_todo(todo_contents)

def t_rm(parsed_args):
    """Remove a task."""
    todo_contents = todo()
    for task in sorted(parsed_args.tasks, reverse=True):
        task = task - 1
        task_obj = todo_contents[task]
        print "Removed task '%s %s'" % (task_obj[0] + 1, task_obj[1])
        del todo_contents[task]
    flush_todo(todo_contents)

def t_ad(parsed_args):
    """Adds the given task and does do now."""
    todo_contents = t_a(parsed_args)
    parsed_args.task = len(todo_contents)
    t_donow(parsed_args)

def _ndt(todo_contents, tasks_to_iterate):
    for task_obj in tasks_to_iterate:
        if TODAY_STR in task_obj[1]:
            print "Making task %s '%s' not due on any day" % (task_obj[0]+1, task_obj[1])
            task_obj = remove_due_today(task_obj)
            todo_contents[task_obj[0]] = task_obj
    flush_todo(todo_contents)

def t_ndt(parsed_args):
    """Makes the given task as not due today."""
    todo_contents = todo()
    if "all" in parsed_args.tasks:
        tasks_to_iterate = todo_contents
    else:
        tasks_to_iterate = [todo_contents[x-1] for x in parsed_args.tasks]
    _ndt(todo_contents, tasks_to_iterate)

def t_dt(parsed_args):
    """Makes the given task as due today."""
    todo_contents = todo()
    for task in parsed_args.tasks:
        task = task - 1
        task_obj = todo_contents[task]
        if TODAY_STR not in task_obj[1]:
            print "Making task %s '%s' as due today" % (task+1, task_obj[1])
            task_obj = (task_obj[0], task_obj[1] + TODAY_STR)
            todo_contents[task] = task_obj
    flush_todo(todo_contents)

def t_ldt(parsed_args):
    """Returns the tasks which are due today."""
    todo_contents = todo()
    project = PROJECT_MAP.get(parsed_args.project_short_name, "")
    print_todo_tasks(
        parsed_args, [x for x in todo_contents
                      if TODAY_STR in x[1] and project in x[1]])

def t_lpdd(parsed_args):
    """Returns the tasks which are due today."""
    todo_contents = todo()
    selected_tasks = []
    for x in todo_contents:
        due_match = re.search(r"due:(\d+-\d+-\d+)", x[1])
        if due_match:
            due_date = datetime.datetime.strptime(
                due_match.group(1), "%Y-%m-%d").date()
            today_date = datetime.datetime.today().date()
            if (due_date - today_date).days < 0:
                selected_tasks.append(x)
    print_todo_tasks(parsed_args, selected_tasks)

def t_lndd(parsed_args):
    """Returns the tasks which are due today."""
    todo_contents = todo()
    project = PROJECT_MAP.get(parsed_args.project_short_name, "")
    todo_contents = [x for x in todo_contents
                     if 'due:' not in x[1] and project in x[1]]
    t_ls(parsed_args, todo_contents)

def get_group_tasks(tasks):
    group_tasks = collections.defaultdict(list)
    if tasks is None:
        tasks = []
    for task in tasks:
        match = re.search(r"(\+\w+( \+\w+))", task[1])
        group_tasks[match.group(1)].append(task)
    return group_tasks

def print_group_sum(tasks, consider_min=None):
    group_tasks = get_group_tasks(tasks)
    for group, tasks_in_group in sorted(group_tasks.items()):
        if group == "+personal +planning":
            continue
        print "Time for '%s': %s" % (
            group, time_taken(tasks_in_group, consider_min=consider_min))
    return set([x for x in group_tasks])

def t_st(parsed_args):
    """Returns the summary of tasks completed today."""
    todo_contents = todo()
    done_contents = done()
    day_str = parsed_args.day_str
    try:
        delta = int(day_str)
        day_str = str((datetime.datetime.now() - datetime.timedelta(days=delta)).date())
    except ValueError:
        pass
    tasks_due_today_pending = [
        remove_due_today(x) for x in todo_contents if TODAY_STR in x[1]]
    tasks_done_today = [x for x in done_contents if day_str in x[1]]
    print "---------------------"
    print "Report for %s" % day_str
    print "---------------------"
    print "Tasks done today (%s):" % len(tasks_done_today)
    print_done_tasks(tasks_done_today)
    print_group_sum(tasks_done_today, consider_min=True)
    print "Total time tasks done: %s" % time_taken(tasks_done_today, consider_min=True)
    print "---------------------"
    if day_str == today():
        print "Tasks pending today (%s):" % len(tasks_due_today_pending)
        print_todo_tasks(parsed_args, tasks_due_today_pending)
        print_group_sum(tasks_due_today_pending)
        print (
            "Total time tasks due: %s" % (time_taken(tasks_due_today_pending)))

def get_project_short_name():
    """Gets the short project name as selected in cmdline."""
    project_short_name = ""
    if len(sys.argv) > 2:
        project_short_name = sys.argv[2]
    return project_short_name

def get_tasks():
    """Gets the tasks selected in cmdline."""
    tasks = []
    if len(sys.argv) > 2:
        for i in range(2, len(sys.argv)):
            tasks.append(int(sys.argv[i])-1)
    return tasks

def get_numbers():
    numbers = []
    if len(sys.argv) > 2:
        for i in range(2, len(sys.argv)):
            numbers.append(int(sys.argv[i]))
    return numbers

def t_wa(parsed_args):
    lines = []
    for project_short_name in projects_without_nil():
        time_in_mins = getattr(parsed_args, project_short_name)
        if not time_in_mins:
            continue
        time_in_mins = int(time_in_mins * 60.0)
        lines.append(
            "%s:%s\n" % (PROJECT_MAP[project_short_name], time_in_mins))
    with open(WP_ALLOC_FILE, "w") as fobj:
        fobj.writelines(lines)
    print "\nWriting following lines to %s" % WP_ALLOC_FILE
    print "".join(lines)

def t_pa(_):
    with open(WP_ALLOC_FILE, "r") as fobj:
        lines = fobj.readlines()
    for line in sorted(lines):
        project, duration = line.split(":")
        duration = round(int(duration) / 60.0, 2)
        print "%30s : %3.1f hrs" % (project, duration)

def t_reo(parsed_args):
    todo_contents = todo()
    new_todo_contents = [todo_contents[x-1] for x in parsed_args.tasks]
    remaining = [x for i, x in enumerate(todo_contents)
                 if i+1 not in parsed_args.tasks]
    print "Moving task in following order:"
    for task in new_todo_contents:
        print task[1]
    new_todo_contents = new_todo_contents + remaining
    flush_todo(new_todo_contents)

def t_sanity(_):
    todo_contents = todo()
    tasks_without_priority = []
    tasks_without_added_date = []
    for task in todo_contents:
        if not get_priority(task[1]):
            tasks_without_priority.append(task)
        if 'added:' not in task[1]:
            tasks_without_added_date.append(task)

    if tasks_without_priority:
        print "Tasks missing priority:"
        for task in tasks_without_priority:
            print_task(task)
    if tasks_without_added_date:
        print "Tasks without added date:"
        for task in tasks_without_added_date:
            print_task(task)

def t_qa(parsed_args):
    if parsed_args.num == 0:
        help_str = '\n'.join(
            ['%s => %s' % (x, y)
             for x, y in QUICK_ADD_TASKS.items()])
        print help_str
    else:
        todo_contents = todo()
        assert parsed_args.num in QUICK_ADD_TASKS.keys()
        task_str = QUICK_ADD_TASKS[parsed_args.num]
        today_str = str(datetime.date.today())
        task_str = task_str.replace('due:today', TODAY_STR)
        task_str += " added:%s" % today_str
        todo_contents.append((0, task_str))
        flush_todo(todo_contents)

def t_stuck(parsed_args):
    days = parsed_args.days
    todo_contents = todo()
    ndt = parsed_args.ndt
    tasks = []
    for task in todo_contents:
        match = re.search(r"added:(\d{4}-\d{2}-\d{2})", task[1])
        if not match:
            print "'%s' doesn't have added date" % task[1]
            continue
        added_date = datetime.datetime.strptime(
            match.group(1), "%Y-%m-%d").date()
        today_date = datetime.datetime.today().date()
        if ndt and TODAY_STR in task[1]:
            continue
        if (today_date - added_date).days > days:
            tasks.append(task)
    print_todo_tasks(parsed_args, tasks)

def t_daystart(parsed_args):
    todo_contents = todo()
    weekday_to_tasks = collections.defaultdict(list)
    day_tasks_file = os.getenv("DAY_TASKS_FILE")
    if day_tasks_file:
        with open(day_tasks_file, "r") as f:
            import pdb; pdb.set_trace()
            weekday_to_tasks = eval(f.read())
    for plugin in PLUGINS:
        if not hasattr(plugin, 'get_day_tasks'):
            continue
        daytasks = plugin.get_day_tasks(parsed_args)
        for key, value in daytasks.items():
            weekday_to_tasks[key].extend(value)
    weekday = datetime.datetime.today().weekday()
    for task in weekday_to_tasks[weekday]:
        args = [todo_contents]
        task[0] = task[0].replace('due:today', TODAY_STR)
        args.extend(task)
        add(*args)
    flush_todo(todo_contents)

def get_sub_parser(parser, command):
    sub_parser = parser.add_parser(command)
    sub_parser.set_defaults(subcommand=command)
    return sub_parser

def ndt_args(value):
    if value == "all":
        return "all"
    else:
        try:
            return int(value)
        except Exception as exception:
            raise argparse.ArgumentTypeError(str(exception))

def config_done_commands(subparsers):
    '''
    Commands that parse the done.txt file and act on tasks that are already
    done.
    '''
    command_map = {}

    # Grep on the done.txt file.
    dng = get_sub_parser(subparsers, 'dng')
    dng.add_argument('task_str', type=str)
    command_map['dng'] = {'method': t_dng}

    # Do a tail on the done.txt file.
    dnt = get_sub_parser(subparsers, 'dnt')
    dnt.add_argument('number_of_lines', type=int, nargs='?', default=10)
    command_map['dnt'] = {'method': t_dnt}

    # Restart a task that is already done and mark as doing not.
    dndonow = get_sub_parser(subparsers, 'dndonow')
    dndonow.add_argument('task', type=int)
    command_map['dndonow'] = {'method': t_dndonow}

    # Add a task that is done back to todo.txt.
    dna = get_sub_parser(subparsers, 'dna')
    dna.add_argument('priority', choices=priorities())
    dna.add_argument('task', type=int)
    command_map['dna'] = {'method': t_dna}

    return command_map

def config_list_commands(subparsers):
    '''
    Commands that are used to list the tasks.
    '''
    command_map = {}

    # Basic list of all the tasks.
    list_tasks = get_sub_parser(subparsers, 'ls')
    list_tasks.add_argument('project_short_name', nargs='?',
                            choices=PROJECT_MAP.keys(), default="")
    list_tasks.add_argument('priority', nargs='?',
                            choices=['a', 'b', 'c', 'd', 'e', 'f', ''], default="")
    list_tasks.add_argument('--numonly', action='store_true')
    command_map['ls'] = {'method': t_ls}

    # List what is due today.
    ldt = get_sub_parser(subparsers, 'ldt')
    ldt.add_argument('project_short_name', nargs='?',
                     choices=PROJECT_MAP.keys(), default="")
    command_map['ldt'] = {'method': t_ldt}

    # List what is due today.
    lpdd = get_sub_parser(subparsers, 'lpdd')
    ldt.add_argument('project_short_name', nargs='?',
                     choices=PROJECT_MAP.keys(), default="")
    command_map['lpdd'] = {'method': t_lpdd}

    # List tasks without any due date.
    lndd = get_sub_parser(subparsers, 'lndd')
    lndd.add_argument('project_short_name', nargs='?',
                      choices=PROJECT_MAP.keys(), default="")
    lndd.add_argument('priority', nargs='?',
                      choices=['a', 'b', 'c', 'd', 'e', 'f', ''], default="")
    lndd.add_argument('--numonly', action='store_true')
    command_map['lndd'] = {'method': t_lndd}

    # List of the summary of both tasks that are due today and what was done
    # today.
    summary_today = get_sub_parser(subparsers, 'st')
    summary_today.add_argument('day_str', nargs="?", type=str, default=today())
    command_map['st'] = {'method': t_st}

    # Graph for this week.
    graph = get_sub_parser(subparsers, 'g')
    graph.add_argument('project_short_name', nargs='?',
                       choices=PROJECT_MAP.keys(), default="")
    command_map['g'] = {'method': t_g}

    # Summary for this week.
    summary = get_sub_parser(subparsers, 's')
    summary.add_argument('project_short_name', nargs='?',
                         choices=PROJECT_MAP.keys(), default="")
    command_map['s'] = {'method': t_s}

    return command_map

def config_do_commands(subparsers):
    '''
    Commands that change the tasks in some way.
    '''
    command_map = {}
    # Make the task as due today.
    due_today = get_sub_parser(subparsers, 'dt')
    due_today.add_argument('tasks', nargs='+', type=int)
    command_map['dt'] = {'method': t_dt}

    # Make the task as not due today.
    not_due_today = get_sub_parser(subparsers, 'ndt')
    not_due_today.add_argument('tasks', nargs='+', type=ndt_args)
    command_map['ndt'] = {'method': t_ndt}

    # Set due date for a task.
    sdd = get_sub_parser(subparsers, 'sdd')
    sdd.add_argument('tasks', nargs='+', type=int)
    sdd.add_argument('due_on', type=str)
    command_map['sdd'] = {'method': t_sdd}

    # Add a new task.
    add_task = get_sub_parser(subparsers, 'a')
    add_task.add_argument('project_short_name', nargs='?',
                          choices=projects_without_nil())
    add_task.add_argument('priority', choices=priorities())
    add_task.add_argument('task_str', type=str)
    add_task.add_argument('-d', '--dt', action='store_true')
    add_task.add_argument('--due_on', type=str)
    command_map['a'] = {'method': t_a}

    # Adds a new task and does donow on that task.
    ad_task = get_sub_parser(subparsers, 'ad')
    ad_task.add_argument('project_short_name', nargs='?',
                         choices=projects_without_nil())
    ad_task.add_argument('priority', choices=priorities())
    ad_task.add_argument('task_str', type=str)
    command_map['ad'] = {'method': t_ad}

    # Remove a task.
    rm_task = get_sub_parser(subparsers, 'rm')
    rm_task.add_argument('tasks', nargs='+', type=int)
    command_map['rm'] = {'method': t_rm}

    # Change the project for a task.
    cp_for_task = get_sub_parser(subparsers, 'cp')
    cp_for_task.add_argument('project_short_name', nargs='?',
                             choices=projects_without_nil())
    cp_for_task.add_argument('tasks', nargs='+', type=int)
    command_map['cp'] = {'method': t_cp}

    # Close and readd a task.
    close_and_readd = get_sub_parser(subparsers, 'car')
    close_and_readd.add_argument('tasks', nargs='+', type=int)
    command_map['car'] = {'method': t_car}

    # Complete a task.
    do_task = get_sub_parser(subparsers, 'do')
    do_task.add_argument('tasks', nargs='+', type=int)
    command_map['do'] = {'method': t_do}

    # Start doing a  task now.
    donow = get_sub_parser(subparsers, 'donow')
    donow.add_argument('task', type=int)
    command_map['donow'] = {'method': t_donow}

    # Change the priority of a task.
    pri = get_sub_parser(subparsers, 'pri')
    pri.add_argument('tasks', nargs='+', type=int)
    pri.add_argument('priority', choices=priorities())
    command_map['pri'] = {'method': t_pri}

    # Change the amount of time spent on the task manually.
    tmin = get_sub_parser(subparsers, 'min')
    tmin.add_argument('task', type=int)
    tmin.add_argument('min', type=int)
    tmin.add_argument('--done', action='store_true')
    command_map['min'] = {'method': t_min}

    # Change the estimate for a task.
    test = get_sub_parser(subparsers, 'est')
    test.add_argument('task', type=int)
    test.add_argument('est', type=int)
    command_map['est'] = {'method': t_est}

    return command_map

def config_special_commands(subparsers):
    '''
    Special commands.
    '''
    command_map = {}

    # Check the sanity of todo.txt and done.txt.
    get_sub_parser(subparsers, 'sanity')
    command_map['sanity'] = {'method': t_sanity}

    # Quick add tasks.
    quick_add = get_sub_parser(subparsers, 'qa')
    quick_add.add_argument('num', type=int)
    command_map['qa'] = {'method': t_qa}

    write_alloc = get_sub_parser(subparsers, 'wa')
    for project_short_name in projects_without_nil():
        write_alloc.add_argument("--" + project_short_name, type=float)
    command_map['wa'] = {'method': t_wa}

    # Print all the stuck tasks.
    stuck = get_sub_parser(subparsers, 'stuck')
    stuck.add_argument('days', type=int)
    stuck.add_argument('--ndt', action='store_true')
    command_map['stuck'] = {'method': t_stuck}

    reo = get_sub_parser(subparsers, 'reo')
    reo.add_argument('tasks', nargs='+', type=int)
    command_map['reo'] = {'method': t_reo}

    get_sub_parser(subparsers, 'pa')
    command_map['pa'] = {'method': t_pa}

    get_sub_parser(subparsers, 'daystart')
    command_map['daystart'] = {'method': t_daystart}

    return command_map

def main():
    """Main entry point for the script."""
    main_module = sys.modules[__name__]
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers(title='subcommands')
    command_map = {}

    returned_command_map = config_done_commands(subparsers)
    command_map.update(returned_command_map)

    returned_command_map = config_list_commands(subparsers)
    command_map.update(returned_command_map)

    returned_command_map = config_special_commands(subparsers)
    command_map.update(returned_command_map)

    returned_command_map = config_do_commands(subparsers)
    command_map.update(returned_command_map)

    for plugin in PLUGINS:
        returned_command_map = plugin.commands(subparsers, main_module)
        command_map.update(returned_command_map)

    parsed_args = parser.parse_args()
    parsed_args.main_module = main_module
    command_map[parsed_args.subcommand]['method'](parsed_args)
    return

main()
