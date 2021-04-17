module github.com/tap2joy/Gateway

go 1.14

require (
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/tap2joy/Protocols v0.0.0-00010101000000-000000000000
	go.elastic.co/apm/module/apmgrpc v1.11.0
	google.golang.org/grpc v1.37.0
	google.golang.org/grpc/examples v0.0.0-20210409234925-fab5982df20a // indirect
	google.golang.org/protobuf v1.25.0
)

replace github.com/tap2joy/Protocols => ./proto
