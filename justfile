# Run acceptance tests
testacc TESTARGS="":
    TF_ACC=1 go test ./... -v {{TESTARGS}} -timeout 120m

# Format all Go source files
fmt:
    @gofmt -w $(find . -name "*.go" -not -path "*/vendor/*")