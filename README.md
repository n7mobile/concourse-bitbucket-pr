# concourse-bitbucket-pr

Resource for ConcourseCI pipelines. Provides simple management of PullRequests:

* Get list of opened PullRequests on BitBucket Cloud
* Checkout repository at commit referenced by PullRequest
* Set status for processed PullRequest on BitBucket Cloud

## Why?

There is a plenty of the ready-to-use projects doing same thing and even more ([just search](https://github.com/search?q=concourse+bitbucket+pullrequest)). 

Unfortunately, using some of them, we hit a limitation of the number of API calls to BitBucket service (1000 requests per hour). As this limit is not so low, just watching for new PRs and set status for them should not exceed this value. But it was. The problem was calling */commits* endpoint in order to fetch full SHA1 of updated commit instead of getting just current HEAD of cloned repository.

Instead of making a fix and PR to one of the existing project, we decided to create resource - absolutely simple, managable and tailored to our case, which does not break the whole CI for our 20+ projects.

## Source Configuration

* `workspace`: *Required.* Name of BitBucket organization/team.

* `slug`: *Required.* Name of BitBucket repository.

* `username`: *Required.* Username of BitBucket account with access to repository. Provided account is used for git clone (HTTPS) and BitBukcet REST API.

* `password`: *Required.* User password or user app password (in case of 2FA).

* `debug`: *Optional.* Default *`false`*. Prints additional logs during processing.

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
          version_path: pull-request
          action: set:commit.build.status
          status: INPROGRESS
          url: example.com
          name: PR resource test
          description: Some userful description
```

## Behavior

### `check`: List of open PullRequests

List of the all open PullRequests for a given repository is fetched by BitBucket API v2. Paging is handled.

Version object is generated as:
```javascript
{
    "commit": "", /* Prefix of SHA1 commit */
    "id":     "", /* Identifier of PullRequest */
    "title":  "", /* Title of PullRequest */
    "branch": "", /* Source branch of PullRequest */
}
```

### `in`: Checkout by commit hash

Generally, `git clone && git checkout 757c47d4` performed in [libgit2](https://libgit2.org).

Version object passed by Concourse is stored as `.concourse.version.json` and avaialable to use by following steps.

### `out`: Set build status

Set a build status on the commit. Particular commit is identified by the hash in pull request.

#### Parameters

* `repo_path`: *Required.* Name of the previous *`get`: concourse-bitbucket-pr* step where checked-out repo may be found.

* `action`: *Required.* Identifier of the action to perform on PR. Currently, only `set:commit.build.status` is supported.

* `status`: *Required.* Commit build status to set. Possible values: `STOPPED`, `INPROGRESS`, `FAILED`, `SUCCESSFUL`

* `url`: *Required.* URL to build status and results, ie. *https://ci.example.com/builds/$BUILD_ID*

* `name`: *Optional.* Name of the build status entity.

* `description`: *Optional.* Description of the build status entity.

## Development

### Prerequisites

* golang is *required* - version 1.15.x or higher is required.
* docker is *required* - version 17.05.x or higher is required.
* make is *required* - version 4.1 of GNU make is tested.
* libgit2 is *required* - version 1.1.0 is tested.

### Operating

All task are defined in the self explanatory Makefile.