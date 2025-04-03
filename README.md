
# MultiDialogo - MultiCarrier Email API

## Requirements

- docker
- docker compose v2
- git

## Provisioning

### Scripts

#### How to start local development environment dependencies

```shell
docker compose --profile devcontainer-deps up -d --build
```

```shell
docker compose --profile devcontainer-deps down --remove-orphans
```

#### Run tests

```shell
./run-tests-local.sh
```

A coverage report will be exported at `.coverage/report.html`

```shell
open ".coverage/report.html"
```

#### Simulate deployment stages

```shell
./run-tests-ci.sh
```

#### Deploy local docker image

```shell
docker build --target deploy -t multicarrier-email-api:local .
```

### Graphic tools

- database administration (dbadmin): http://localhost:9001
