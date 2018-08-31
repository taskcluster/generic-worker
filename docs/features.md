---
title: Features
order: 10
---

Features are capabilities that can be enabled in the generic worker for use by
a task.

These features are enabled by declaring them within the task payload in the
`features` object.

Note: Some features require additional information within the task definition.
Features may also require scopes.  Consult the documentation for each feature
to understand the requirements.

Example:

```js
{
  "payload": {
    "features": {
      "chainOfTrust": true
    }
  }
}
```

## Feature: `chainOfTrust`

#### Since: generic-worker 5.3.0

Enabling this feature will mean that the generic worker will publish an
additional task artifact `public/chainOfTrust.json.asc`. This will be a clear
text openpgp-signed json object, storing the SHA 256 hashes of the task
artifacts, plus some information about the worker. This is signed by a openpgp
private key, both generated and stored on the worker. This private key is never
transmitted across the network. In future you will be able to verify the
signature of this artifact against the public openpgp key of the worker type,
to be confident that it really was created by the worker. However currently
this is not possible, since we do not yet publish the openpgp public key
anywhere. When this has been implemented, this page will be updated with
details about how to retrieve the public key, for signature verification.

The worker uses the openpgp private key from the file location specified by the
[worker configuration
setting](/docs/reference/workers/generic-worker#set-up-your-env)
`signingKeyLocation`.

No scopes are presently required for enabling this feature.

References:

* [Bugzilla bug](https://bugzilla.mozilla.org/show_bug.cgi?id=1287112)
* [Source code](https://github.com/taskcluster/generic-worker/blob/master/chain_of_trust.go)


## Feature: `taskclusterProxy`

#### Since: generic-worker 10.6.0

The taskcluster proxy provides an easy and safe way to make authenticated
taskcluster requests within the scope(s) of a particular task.

For example lets say we have a task like this:

```js
{
  "scopes": ["a", "b"],
  "payload": {
    "features": {
      "taskclusterProxy": true
    }
  }
}
```

A web service will execute (typically on port 80) of the local machine for the
duration of the task, with which you can proxy unauthenticated requests to
various taskcluster services. The proxy will inject the Authorization http
header for you and proxy the request to the target service, granting the
request the scopes of the task (in this case ["a", "b"]).

| Target Destination                             | Proxy Address                            |
|------------------------------------------------|------------------------------------------|
| https://queue.taskcluster.net/<PATH>           | http://localhost/queue/<PATH>            |
| https://index.taskcluster.net/<PATH>           | http://localhost/index/<PATH>            |
| https://aws-provisioner.taskcluster.net/<PATH> | http://localhost/aws-provisioner/<PATH>  |
| https://secrets.taskcluster.net/<PATH>         | http://localhost/secrets/<PATH>          |
| https://auth.taskcluster.net/<PATH>            | http://localhost/auth/<PATH>             |
| https://hooks.taskcluster.net/<PATH>           | http://localhost/hooks/<PATH>            |
| https://purge-cache.taskcluster.net/<PATH>     | http://localhost/purge-cache/<PATH>      |

For example (using curl) inside a task container.

```sh
cat secret | curl --header 'Content-Type: application/json' --request PUT --data @- http://localhost/secrets/v1/secret/<secretName>
```

You can also use the `baseUrl` parameter in the taskcluster-client

```js
var taskcluster = require('taskcluster-client');
var queue = new taskcluster.Queue({
 baseUrl: 'http://localhost/queue'
 });

queue.createTask(...);
```

References:

* [taskcluster-proxy](https://github.com/taskcluster/taskcluster-proxy)
