#!/usr/bin/env bash
echo "Building and running the PIC-4 Demonstrator..."
go build -o demonstrator.exe types.go validator.go attester.go permit_controller.go actuator.go main.go
./demonstrator.exe
