#!/bin/sh

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TASKS_DIR=$DIR/tasks.d

echo ">>> Running subtasks in $TASKS_DIR..."
for SH in $TASKS_DIR/*.sh ; do
	if [ -x $SH ] ; then
		exec $SH
	fi
done



