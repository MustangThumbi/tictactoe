PROTO_SRC = proto/tictactoe.proto
GO_OUT = genproto

gen:
	protoc -I. \
	       -I./proto \
	       -I./googleapis \
	       --go_out=$(GO_OUT) --go_opt=paths=source_relative \
	       --go-grpc_out=$(GO_OUT) --go-grpc_opt=paths=source_relative \
	       --grpc-gateway_out=$(GO_OUT) --grpc-gateway_opt=paths=source_relative,logtostderr=true \
	       $(PROTO_SRC)

clean:
	rm -rf $(GO_OUT)/tictactoe/*.go
