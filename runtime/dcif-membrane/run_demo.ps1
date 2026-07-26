Write-Host "Building and running the DCIF Membrane Demonstrator..."
go build -o dcif-membrane.exe types.go operator.go main.go
.\dcif-membrane.exe
