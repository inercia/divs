#!/bin/sh

TASKS_DIR=./tasks.d

echo ">>> Running subtasks in $TASKS_DIR..."
for SH in $TASKS_DIR/*.sh ; do
	if [ -x $SH ] ; then
		$SH $@
	fi
done



