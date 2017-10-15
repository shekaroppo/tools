#!/usr/bin/python

import os
import re
import subprocess
import sys

import argparse

NOTES_DIR = os.getenv('NOTES_DIR')
EDITOR = os.getenv('EDITOR', 'vim')
DEFAULT_NUM_ENTRIES = 5

if not NOTES_DIR:
    print 'Error !! Environment variable "NOTES_DIR" not defined'
    sys.exit(1)


def _get_notes():
    note_files = [os.path.join(NOTES_DIR, note_file)
                  for note_file in os.listdir(NOTES_DIR)]
    notes_stats = [(os.stat(note_file), note_file)
                   for note_file in note_files]
    notes_stats = sorted(notes_stats, key=lambda x: x[0].st_atime)
    notes_stats = list(reversed(notes_stats))
    return notes_stats


def list_notes(parsed_args):
    notes_stats = _get_notes()
    if parsed_args.reversed:
        notes_stats = list(reversed(notes_stats))
    if not parsed_args.num_entries:
        num_entries = DEFAULT_NUM_ENTRIES
    elif parsed_args.all:
        num_entries = len(notes_stats)
    else:
        num_entries = parsed_args.num_entries

    notes_stats = notes_stats[:num_entries]
    for note_stats in notes_stats:
        print os.path.basename(note_stats[1])


def edit_note(parsed_args):
    notes_stats = _get_notes()
    matching_names = [(x, y[1])
                      for x, y in enumerate(notes_stats)
                      if parsed_args.name in y[1]]
    if len(matching_names) > 1:
        print ('Multiple matches for the filename. '
               'Please choose specific one:')
        print '\n'.join(os.path.basename(y) for _, y in matching_names)
        return
    elif len(matching_names) == 0:
        if parsed_args.create:
            filename = os.path.join(NOTES_DIR, parsed_args.name)
        else:
            print 'No matching notes found. Use -c to create one.'
            return
    else:
        filename = matching_names[0][1]

    subprocess.call([EDITOR, filename])


def grep_notes(parsed_args):
    notes_stats = _get_notes()
    note_name_to_index = {}
    for idx, note_stat in enumerate(notes_stats):
        note_name_to_index[note_stat[1]] = idx
    grep_str = parsed_args.grep_str
    command = ["grep", "-i", "-l", grep_str]
    command.extend(note_name_to_index.keys())
    process = subprocess.Popen(command, stdout=subprocess.PIPE)
    stdout, _ = process.communicate()
    file_names_matching = [x for x in stdout.split('\n') if x]
    for file_name in file_names_matching:
        assert file_name in note_name_to_index
        print os.path.basename(file_name)


def clean_notes(_):
    for note_file in os.listdir(NOTES_DIR):
        if re.match(r'^.*\.sw(o|p|n)$', note_file):
            print 'Removing file', note_file
            file_full_path = os.path.join(NOTES_DIR, note_file)
            os.remove(file_full_path)


FUNCTION_MAP = {
    'ls': list_notes,
    'clean': clean_notes,
    'edit': edit_note,
    'grep': grep_notes,
}


def add_subcommand(parser, command):
    subparser = parser.add_parser(command)
    subparser.set_defaults(subcommand=command)
    return subparser


def main():
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers(title='subcommands')

    ls_parser = add_subcommand(subparsers, 'ls')
    ls_parser.add_argument('-n', '--num-entries', type=int, default=0)
    ls_parser.add_argument('-r', '--reversed', action='store_true')
    ls_parser.add_argument('-a', '--all', action='store_true')

    edit_parser = add_subcommand(subparsers, 'edit')
    edit_parser.add_argument('name', type=str)
    edit_parser.add_argument('-c', '--create', action='store_true')

    add_subcommand(subparsers, 'clean')

    grep_parser = add_subcommand(subparsers, 'grep')
    grep_parser.add_argument('grep_str', type=str)

    parsed_args = parser.parse_args()
    FUNCTION_MAP[parsed_args.subcommand](parsed_args)

if __name__ == "__main__":
    main()
