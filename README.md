# concourse-bitbucket-pr

Resource for ConcourseCI pipelines. Provides simple management of PullRequests:

* Get list of opened PullRequests on BitBucket Cloud
* Checkout repository at commit referenced by PullRequest
* Set status for processed PullRequest on BitBucket Cloud

Docker image is hosted on dockerhub [n7docker/concourse-bitbucket-pr](https://hub.docker.com/r/n7docker/concourse-bitbucket-pr).

## Why?

There is a plenty of the ready-to-use projects doing same thing and even more ([just search](https://github.com/search?q=concourse+bitbucket+pullrequest)). 

Unfortunately, using some of them, we hit a limitation of the number of API calls to BitBucket service (1000 requests per hour). As this limit is not so low, just watching for new PRs and set status for them should not exceed this value. But it was. The problem was calling */commits* endpoint in order to fetch full SHA1 of updated commit instead of getting just current HEAD of cloned repository.

Instead of making a fix and PR to one of the existing project, we decided to create resource - absolutely simple, manageable and tailored to our case, which does not break the whole CI for our 20+ projects.

## Who?

Resource is actively maintained by [N7Mobile sp. z o.o.](https://n7mobile.com). We love OpenSource as it's free in many meanings.

## Source Configuration

* `workspace`: *Required.* Name of BitBucket organization/team.

* `slug`: *Required.* Name of BitBucket repository.

* `username`: *Required.* Username of BitBucket account with access to repository. Provided account is used for git clone (HTTPS) and BitBucket REST API.

* `password`: *Required.* User password or user app password (in case of 2FA).

* `debug`: *Optional.* Default *`false`*. Prints additional logs during processing.

* `recurse_submodules`: *Optional.* Default *`false`*. Checkouts all submodules, if repo has any. Credentails are taken from the configuration of parent repo.

### Example

Example files are placed in the `examples` directory, unexpectedly.

```yaml
resource_types:
  - name: bitbucket-pr
    type: registry-image
    source:
      repository: n7mobile/concourse-bitbucket-pr

resources:
  - name: pull-request
    type: bitbucket-pr
    check_every: 1m
    source:
      workspace: n7mobile
      slug: pr-test-repo
      username: ((ci_user))
      password: ((ci_pass))
      debug: true

jobs:
  - name: foo
    plan:
      - get: pull-request
        trigger: true
        version: every
      - put: pull-request
        params:
          repo_path: pull-request
          action: set:commit.build.status
          status: INPROGRESS
          url: example.com
          name: PR resource test
          description: Some userful description
```

## Behavior

### `check`: List of open PullRequests

List of the all open PullRequests for a given repository is fetched by BitBucket API v2. Paging is handled.

In order to retrieve full SHA1 of the last PR commit, bare checkout of a git is performed. It's minimalizes usage of the BitBucket API.

Version object is generated as:
```javascript
{
    "ref": "", /* Full SHA1 of the commit */
    "id": ""   /* Identifier of Pullrequest */
}
```

### `in`: Checkout by commit hash

Generally, `git clone && git checkout 757c47d4` performed by [libgit2](https://libgit2.org).

Version object passed by Concourse is stored as `.concourse.version.json` and available to use by `out` step.

### `out`: Set build status

Set a build status on the commit. Particular commit is identified by the hash in pull request.

#### Parameters

* `repo_path`: *Required.* Name of the previous *`get`: concourse-bitbucket-pr* step where checked-out repo may be found.

* `action`: *Optional.* Default *`set:commit.build.status`*. Identifier of the action to perform on PR. Currently, only `set:commit.build.status` is supported.

* `key`: *Optional.* Default *`BUILD`*. Key of the commit build status. Passing multiple build statuses identified by a single key will overwrite each other. In case of the multiple builds from single commit (i.e. flutter -> iOS + Android), identifier of the current build `iOS-$BUILD_NAME` and `Android-$BUILD_NAME` should be passed.

* `status`: *Required.* Commit build status to set. Possible values: `STOPPED`, `INPROGRESS`, `FAILED`, `SUCCESSFUL`

* `url`: *Required.* URL to build status and results, ie. *https://ci.example.com/builds/$BUILD_ID*

* `name`: *Optional.* Default *`$BUILD_JOB_NAME #$BUILD_ID`*. Name of the build status entity.

* `description`: *Optional.* Default *`Concourse Build CI`*. Description of the build status entity.

## Development

### Prerequisites

* golang is *required* - version 1.15.x or higher is required.
* docker is *required* - version 17.05.x or higher is required.
* make is *required* - version 4.1 of GNU make is tested.
* libgit2 is *required* - version 1.1.0 is tested.

### Operating

All task are defined in the self explanatory Makefile.

## License

```
MIT License

Copyright (c) [year] [fullname]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```