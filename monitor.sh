#! /bin/bash

checkserver() {
	res=`/usr/sbin/pidof server|wc -l`
	case "$res" in
	0)  echo "Restarting server:     $(date)" >> /tmp/monitor.log
    	t=`date`
    	/cashcow/release/server > "/tmp/std-$t.log" 2>&1  &
    	;;
	1)  # all ok
    	;;
	*)  echo "Removed double server: $(date)" >> /monitor.logv
    	kill $(pidof server | awk '{print $1}')
    	;;
	esac
}

while true
do
  # loop infinitely
	checkserver
	sleep 5
done
