#!/bin/bash

clear

# Build LinusGates
set -e
go build -ldflags="-s -w" -o linusgates
set +e

# Hide typed characters and the cursor
stty -echo
tput civis

# Start LinusGates as root
sudo ./linusgates

# Restore terminal
stty echo
tput cnorm
