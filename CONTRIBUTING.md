## How changes are merged

Development happens in a private mirror of this repository. Pull requests opened
here are reviewed as usual, but a maintainer applies accepted changes in the
mirror instead of merging them here; they reach this repository on the next sync.

### For maintainers

- **Never merge a pull request in the public repository**, including Dependabot
  PRs. The sync script only fast-forwards public `main`, so any commit merged
  there blocks every later sync until it's brought into the mirror's history.
- Merge changes into the mirror's `main`, then sync them to public:

  ```sh
  ./scripts/sync_to_public.sh          # dry run: lists the commits to publish
  ./scripts/sync_to_public.sh --push   # fast-forwards public main
  ```

- Releases run only in the public repository. After public CI passes on the
  synced commit, tag it there to trigger the release:

  ```sh
  git fetch public main
  git tag -a vX.Y.Z public/main -m "vX.Y.Z"
  git push public vX.Y.Z
  ```

## Requirements

- [Go](https://golang.org/doc/install) >= 1.26
- [Terraform](https://www.terraform.io/downloads.html) >= v1.4.6
- [golangci-lint](https://golangci-lint.run/) v2.12.2, to match CI

## Building the provider

1.  Clone the repository:

    ```sh
    git clone https://github.com/pinecone-io/terraform-provider-pinecone
    ```

1.  Build and install the provider. The install directory depends on the `GOPATH`
    environment variable.

    ```sh
    go install .
    ```

## Testing a local build

`go install .` puts the binary in `$(go env GOPATH)/bin`. To make Terraform use it
instead of the Registry release, create a CLI config file:

```hcl
provider_installation {
  dev_overrides {
    "pinecone-io/pinecone" = "/path/to/go/bin"
  }
  direct {}
}
```

Point Terraform at it with `export TF_CLI_CONFIG_FILE=/path/to/that/file`, set your
credentials as described in [Authentication](README.md#authentication), then run
`terraform plan` or `terraform apply` from any directory with a config (the
[examples](./examples) are a good start). `terraform init` isn't needed while dev
overrides are active.

## Running tests and lint

Unit tests and lint don't need credentials:

```sh
go test ./pinecone/...
golangci-lint run
```

Acceptance tests create real resources in Pinecone, so they need credentials and
may incur costs:

```sh
export PINECONE_API_KEY=...
export PINECONE_CLIENT_ID=...
export PINECONE_CLIENT_SECRET=...
export PINECONE_PROJECT_ID=...
TF_ACC=1 go test -timeout 15m -v ./pinecone/provider/
```

Use `-run` to run a subset of tests, for example `-run TestAccIndexResource`.

## Generating documentation

The files in `docs/` are generated from the provider schema, `examples/`, and
`templates/`. After changing any of them, regenerate and commit the result:

```sh
go generate ./...
```

CI fails if `docs/` doesn't match the generated output.
