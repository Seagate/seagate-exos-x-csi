#!/bin/bash

lastpodcount=0
lastpending=0
lastinit=0
lastrunning=0
lastcompleted=0
lastiniterror=0
lastpvcs_count=0
lastpvcs_pending=0
lastpvcs_bound=0
lastpvcs_failed=0
lastpvcs_terminating=0
lastpvs_count=0
lastpvs_bound=0
lastpvs_released=0

doprint=0
skipped=0

function update()
{
    doprint=0

    pods="$(kubectl get pod -A 2> /dev/null | grep "pod-" | sed '/^\s*$/d')"
    podcount=$(echo -n "$pods" | grep -c '^')
    pending=$(echo -n "$pods" | grep Pending | grep -c '^')
    init=$(echo -n "$pods" | grep Init:0 | grep -c '^')
    running=$(echo -n "$pods" | grep Running | grep -c '^')
    completed=$(echo -n "$pods" | grep Completed | grep -c '^')
    initerror=$(echo -n "$pods" | grep Init:Error | grep -c '^')

    pvcs="$(kubectl get pvc -A 2> /dev/null | grep "my-marvelous-storage" | sed '/^\s*$/d')"
    pvcs_count=$(echo -n "$pvcs" | grep -c '^')
    pvcs_pending=$(echo -n "$pvcs" | grep Pending | grep -c '^')
    pvcs_bound=$(echo -n "$pvcs" | grep Bound | grep -c '^')
    pvcs_failed=$(echo -n "$pvcs" | grep Failed | grep -c '^')
    pvcs_terminating=$(echo -n "$pvcs" | grep Terminating | grep -c '^')

    pvs="$(kubectl get pv -A 2> /dev/null | grep "my-marvelous-storage" | sed '/^\s*$/d')"
    pvs_count=$(echo -n "$pvs" | grep -c '^')
    pvs_bound=$(echo -n "$pvs" | grep Bound | grep -c '^')
    pvs_released=$(echo -n "$pvs" | grep Released | grep -c '^')

    # PODs
    if [ "$lastpodcount" != "$podcount" ]; then
        lastpodcount=$podcount
        doprint=1
    fi
    if [ "$lastpending" != "$pending" ]; then
        lastpending=$pending
        doprint=1
    fi
    if [ "$lastinit" != "$init" ]; then
        lastinit=$init
        doprint=1
    fi
    if [ "$lastrunning" != "$running" ]; then
        lastrunning=$running
        doprint=1
    fi
    if [ "$lastcompleted" != "$completed" ]; then
        lastcompleted=$completed
        doprint=1
    fi
    if [ "$lastiniterror" != "$initerror" ]; then
        lastiniterror=$initerror
        doprint=1
    fi

    # PVCs
    if [ "$lastpvcs_count" != "$pvcs_count" ]; then
        lastpvcs_count=$pvcs_count
        doprint=1
    fi
    if [ "$lastpvcs_pending" != "$pvcs_pending" ]; then
        lastpvcs_pending=$pvcs_pending
        doprint=1
    fi
    if [ "$lastpvcs_bound" != "$pvcs_bound" ]; then
        lastpvcs_bound=$pvcs_bound
        doprint=1
    fi
    if [ "$lastpvcs_failed" != "$pvcs_failed" ]; then
        lastpvcs_failed=$pvcs_failed
        doprint=1
    fi
    if [ "$lastpvcs_terminating" != "$pvcs_terminating" ]; then
        lastpvcs_terminating=$pvcs_terminating
        doprint=1
    fi

    # PVs
    if [ "$lastpvs_count" != "$pvs_count" ]; then
        lastpvs_count=$pvs_count
        doprint=1
    fi
    if [ "$lastpvs_bound" != "$pvs_bound" ]; then
        lastpvs_bound=$pvs_bound
        doprint=1
    fi
    if [ "$lastpvs_released" != "$pvs_released" ]; then
        lastpvs_released=$pvs_released
        doprint=1
    fi
}

function printupdate()
{
    if [ "$doprint" -eq "1" ]; then
        if [ "$skipped" -eq "1" ]; then
            printf "\n"
            skipped=0
        fi
        printf "$(date) 🔹 "
        printf "Pods: % 4d (% 4d ⏳️, % 4d init, % 4d 🏃, % 4d ✅, % 4d initerror)" $podcount $pending $init $running $completed $initerror
        printf " 🔹 PVCs: % 4d (% 4d ⏳️, % 4d bound, % 4d 💥, % 4d terminating)" $pvcs_count $pvcs_pending $pvcs_bound $pvcs_failed $pvcs_terminating
        printf " 🔹 PVs: % 4d (% 4d bound, % 4d released)\n" $pvs_count $pvs_bound $pvs_released
    else
        skipped=1
        printf "."
    fi
}

while sleep 1; do
    update
    printupdate
done
