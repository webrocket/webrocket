#!/bin/sh

osname=`uname -s`
color=36

if [ "$osname" != "Darwin" ]; then
	echo="echo -e"
else
	echo=echo
fi

text=$(echo $@ | sed -e 's/\*\([^\*]*\)\*/\\033['$color';1m\1\\033[0m/g')
$echo "\n\033[35m--- \033["$color"m$text\033[0m"
