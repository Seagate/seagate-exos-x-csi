#!/bin/bash

while sleep 1; do
    pods="$(kubectl get pod -A 2> /dev/null | grep "pod-" | sed '/^\s*$/d')"
    count=$(echo -n "$pods" | grep -c '^')
    pending=$(echo -n "$pods" | grep Pending | grep -c '^')
    init=$(echo -n "$pods" | grep Init: | grep -c '^')
    running=$(echo -n "$pods" | grep Running | grep -c '^')
    completed=$(echo -n "$pods" | grep Completed | grep -c '^')
    terminating=$(echo -n "$pods" | grep Terminating | grep -c '^')

    pvcs="$(kubectl get pvc -A 2> /dev/null | grep "gallium-storage" | sed '/^\s*$/d')"
    pvcs_count=$(echo -n "$pvcs" | grep -c '^')
    pvcs_pending=$(echo -n "$pvcs" | grep Pending | grep -c '^')
    pvcs_bound=$(echo -n "$pvcs" | grep Bound | grep -c '^')
    pvcs_failed=$(echo -n "$pvcs" | grep Failed | grep -c '^')
    pvcs_terminating=$(echo -n "$pvcs" | grep Terminating | grep -c '^')

    pvs="$(kubectl get pv -A 2> /dev/null | grep "gallium-storage" | sed '/^\s*$/d')"
    pvs_count=$(echo -n "$pvs" | grep -c '^')
    pvs_bound=$(echo -n "$pvs" | grep Bound | grep -c '^')
    pvs_released=$(echo -n "$pvs" | grep Released | grep -c '^')


    printf "$(date) 🔹 "
    printf "Pods: % 4d (% 4d ⏳️, % 4d init, % 4d 🏃, % 4d ✅, % 4d terminating)" $count $pending $init $running $completed $terminating
    printf " 🔹 PVCs: % 4d (% 4d ⏳️, % 4d bound, % 4d 💥, % 4d terminating)" $pvcs_count $pvcs_pending $pvcs_bound $pvcs_failed $pvcs_terminating
    printf " 🔹 PVs: % 4d (% 4d bound, % 4d released)\n" $pvs_count $pvs_bound $pvs_released
done
