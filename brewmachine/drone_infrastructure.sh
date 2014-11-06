#!/bin/sh
iwconfig ath0 mode managed essid secret_robot; ifconfig ath0 192.168.0.40 netmask 255.255.255.0 up; route add default gw 192.168.0.1
