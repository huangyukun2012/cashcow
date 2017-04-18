#! /bin/bash

case "$(pidof server | wc -w)" in

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
