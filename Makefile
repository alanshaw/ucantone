.PHONY: cover, gen

cover:
	mkdir -p coverage
	go test -coverprofile=./coverage/c.out -v ./...
	go tool cover -html=./coverage/c.out

gen:
	rm ./examples/types/cbor_gen.go || true
	cd ./examples/types/gen && go run ./main.go
	rm ./result/datamodel/cbor_gen.go || true
	cd ./result/datamodel/gen && go run ./main.go
	rm ./testutil/datamodel/cbor_gen.go || true
	cd ./testutil/datamodel/gen && go run ./main.go
	rm ./ucan/container/datamodel/cbor_gen.go || true
	cd ./ucan/container/datamodel/gen && go run ./main.go
	rm ./ucan/delegation/datamodel/cbor_gen.*.go || true
	cd ./ucan/delegation/datamodel/gen && go run ./main.go
	rm ./ucan/invocation/datamodel/cbor_gen.*.go || true
	cd ./ucan/invocation/datamodel/gen && go run ./main.go
	rm ./ucan/receipt/datamodel/cbor_gen.go || true
	cd ./ucan/receipt/datamodel/gen && go run ./main.go
	rm ./validator/datamodel/cbor_gen.go || true
	cd ./validator/datamodel/gen && go run ./main.go
	rm ./validator/internal/fixtures/datamodel/dag_json_gen.go || true
	cd ./validator/internal/fixtures/datamodel/gen && go run ./main.go
