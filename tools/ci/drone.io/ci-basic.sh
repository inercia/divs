#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TASKS_DIR=$DIR/tasks.d

WORKSPACE=`pwd`
while [[ $# -gt 0 ]] && [[ ."$1" = .--* ]] ;
do
    opt="$1";
    shift;                           # expose next argument
    case "$opt" in
        "--" ) break 2;;
        "--workspace" )
           WORKSPACE="$1"; shift;;
        "--workspace="* )            # alternate format: --first=date
           WORKSPACE="${opt#*=}";;
        *)
           echo >&2 "Invalid option: $@"; exit 1;;
   esac
done

echo ">>> Running subtasks in $TASKS_DIR..."
run-parts --exit-on-error $TASKS_DIR



