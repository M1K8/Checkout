module github.com/M1K8/Checkout

go 1.15

require (
	github.com/M1K8/internal/api v0.0.1
	github.com/gorilla/mux v1.8.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/testify v1.6.1 // indirect
)

replace github.com/M1K8/internal/api => ./internal/api
