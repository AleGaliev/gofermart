.PHONY: test clean


build-gophermart:
	@echo "🚀 Building server..."
	go build -o cmd/gophermart/gophermart cmd/gophermart/*.go

test_iter1:
	@echo "🧪 Running tests iter1"
	./gophermarttest -test.v -test.run=^TestGophermart$$ -gophermart-binary-path=cmd/gophermart/gophermart -gophermart-host=localhost -gophermart-port=8080 -gophermart-database-uri="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" -accrual-binary-path=cmd/accrual/accrual_darwin_amd64 -accrual-host=localhost -accrual-port=3333 -accrual-database-uri="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

clean:
	rm -rf cmd/server/server

rebuild: clean build

