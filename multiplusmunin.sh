#!/bin/bash
case $1 in
	   config)
curl --fail -s http://localhost:8080/muninconfig
		exit $?;;
esac

curl --fail -s http://localhost:8080/munin
