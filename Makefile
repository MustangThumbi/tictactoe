run-orders:
	@go run services/orders/*.go

run-kitchen:
	@go run services/kitchen/*.go

gen:
	@protoc \
		--proto_path=proto proto/tictactoe.proto \
		--go_out=backend/genproto/tictactoe --go_opt=paths=source_relative \
		--go-grpc_out=backend/genproto/tictactoe --go-grpc_opt=paths=source_relative
