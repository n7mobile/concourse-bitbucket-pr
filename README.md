# bitbucket-pr

Write a description of the resource here.

## Source Configuration

* `a`: *Required.* This is a required setting.

* `b`: *Optional.* This is an optional setting.

* `c`: *Optional. Default `true`* This is an optional setting with a default value.

### Example

```yaml
resource_types:
- name: bitbucket-pr
  type: registry-image
  source:
    repository: n7mobile/ci-bitbucket-pr

resources:
- name: bitbucket-pr
  type: bitbucket-pr
  check_every: 5m
  source:
    log_level: debug

jobs:
- name: do-it
  plan:
  - get: bitbucket-pr
    trigger: true
  - put: bitbucket-pr
    params:
      version_path: bitbucket-pr/version
```

## Behavior

### `check`: Check for something

Write a description of what is checked here.

### `in`: Fetch something

Write a description of what is fetched here.

#### Parameters

* `a`: *Required.* This is a required parameter.

* `b`: *Optional.* This is an optional parameter.

### `out`: Put something somewhere

Write a description of what is being put somewhere.

#### Parameters

* `a`: *Required.* This is a required parameter.

* `b`: *Optional. Default `true`* This is an optional parameter with a default value.

## Development

### Prerequisites

* golang is *required* - version 1.11.x or higher is required.
* docker is *required* - version 17.05.x or higher is required.
* make is *required* - version 4.1 of GNU make is tested.

### Running the tests

The Makefile includes a `test` target, and tests are also run inside the Docker build.

Run the tests with the following command:

```sh
make test
```

### Building and publishing the image

The Makefile includes targets for building and publishing the docker image. Each of these
takes an optional `VERSION` argument, which will tag and/or push the docker image with
the given version.

```sh
make VERSION=1.2.3
make publish VERSION=1.2.3
```
