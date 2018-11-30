# K8s FDW for Postgres

![build status](https://api.travis-ci.org/liztio/k8s-fdw.svg?branch=master)

Ever wanted to fetch information about a Kubernetes cluster directly from Postgres? 
Now you can! 

## Let me play with it!

master is pushed to Docker Hub by Travis. You can run it with:

```shell
docker run -v <path-to-kubeconfig>:/kubeconfig --rm --name=k8s_fdw liztio/k8s_fdw:latest
```

Then you can get a Postgres shell with:

```shell
docker exec -ti k8s_fdw /bin/sh -c 'psql --user postgres'
```

### Examples

Execute following SQL statements to load an extension:

```sql
CREATE SERVER IF NOT EXISTS kind
  FOREIGN DATA WRAPPER k8s_fdw
  OPTIONS (kubeconfig '/kubeconfig');

CREATE FOREIGN TABLE IF NOT EXISTS pods (
  name      text OPTIONS (alias 'metadata.name')
, namespace text OPTIONS (alias 'metadata.namespace')
, container text OPTIONS (alias '{.spec.containers[0].image}')
, labels   jsonb OPTIONS (alias 'metadata.labels')
)
  SERVER kind
  OPTIONS (
    namespace 'kube-system'
  , apiVersion 'v1'
  , kind 'Pod'
  );
```

And finally, run the query:

```sql
SELECT * FROM pods;
```

### Build

This project uses [bazel][https://bazel.build] as a build system. To just build the SO file:

```shell
bazel build //src:libgo_fdw.so
```

### Install

To install an extension just copy it to your Postgres installation. At the end of the build process you'll see following lines:

```
/usr/bin/install -c -m 755 ./bazel-bin/src/libgo_fdw.so  '/usr/lib/postgresql/10/lib/go_fdw.so'
/usr/bin/install -c -m 644 .//k8s_fdw.control '/usr/share/postgresql/10/extension/'
/usr/bin/install -c -m 644 .//k8s_fdw--0.1.sql  '/usr/share/postgresql/10/extension/'
```

### Docker

You can execute the same commands to install an extension locally.

### Testing

## Hacking

An official documentation for FDW callbacks can be found [here](https://www.postgresql.org/docs/9.6/static/fdwhandler.html).
And the documentation for Postgres sources is [here](https://doxygen.postgresql.org).


### Docker Images

Bazel can create and install docker images for you.

``` shell
bazel run //:k8s_fdw_image
```

Will create an image named `bazel:k8s_fdw_image`. You can then run it with:

``` shell
docker run -v /tmp/config:/kubeconfig --rm --name=k8s_fdw bazel:k8s_fdw_image
```
