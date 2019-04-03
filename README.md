## IRONdb release `0.15.3` includes native prometheus support

The irondb-prometheus-adapter is deprecated by native IRONdb support.  Users are encouraged to switch to native ingestion which has improved performance.  This adapter will receive minimal support going forward.

# irondb-prometheus-adapter

Prometheus Adapter to IRONdb.

Requires IRONdb `>= v0.12.0`.
Requires Go `>= 1.10`.

Also check out our launch [blog post](https://www.circonus.com/2018/06/prometheus-adapter/)!

## Docker Image Pull

You can now install the irondb-prometheus-adapter container image from dockerhub
with the below command:

```bash
docker pull irondb/irondb-prometheus-adapter
docker run -p1234:8080 irondb/irondb-prometheus-adapter:latest -addr :1234 -log debug -snowth http://127.0.0.1:8112
```

## Build

Included is a Makefile which will perform all the build tasks you will need
in order to build this service.  Below is a quick outline of typical build
cases:

```bash
make clean # clean and remove prior exe
make build # will build `irondb-prometheus-adapter` exe
make docker # will build a docker scratch container with irondb-prometheus-adapter inside
```

## Run

After building, you can run in two ways, directly or through docker:

```bash
docker run -d -p8080:8080 irondb-prometheus-adapter:<commit_id>
# the above will run irondb-prometheus-adapter in a container
docker run -p1234:8080 irondb-prometheus-adapter:860a64f4e9d793beaef25196e36a35da1480d88b -addr :1234 -log debug -snowth http:127.0.0.1:8112
# the above shows an example of specifying custom command with non-default args

./irondb-prometheus-adapter
# the above will run irondb-prometheus-adapter outside of a container
```

### Arguments

-addr :8080

The above will set the server to listen on port 8080.

-log debug

The above will set the server to use debug logging.  Log levels supported are:

* debug
* warn
* error
* off

-snowth http://127.0.0.1:8112 -snowth http://127.0.0.1:8113

The above will setup two base snowth endpoints that the service will use
to communicate with.  You can specify multiple `-snowth` params if you know
of more than one in a given topology.  The server will interrogate the
listed snowth servers to figure out other nodes in the topology.

### Endpoints

There are three endpoints currently:

```
POST /prometheus/2.0/write/:account/:check_uuid/:check_name
POST /prometheus/2.0/read/:account/:check_uuid/:check_name
GET /health-check
```

The `/health-check` endpoint will return with a "message", "commitID" and 
"buildTime" in the response.

The `write` endpoint will take in prometheus encoded metric data, and convert
it to IRONdb encoded flatbuffer data and submit to the snowth nodes.

The `read` endpoint will pull IRONdb encoded metric data from the snowth nodes
and convert it to prometheus encoded metric data for response to the caller.
