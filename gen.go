package main

//go:generate mkdir -p pb
//go:generate protoc --go-grpc_out=./pb --go_out=./pb mimic.proto
