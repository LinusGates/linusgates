#!/bin/bash
cols=$(tput cols)
lines=$(tput lines)

sudo rm keyCalibration.json #Cheeky way to also ask for sudo password ahead of time
go build && sudo ./linusgates --hLines $cols --vLines $lines
