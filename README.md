
# MultiDialogo - MailCulator server

## Requirements

- docker
- docker compose v2
- git

## Provisioning

### Scripts

#### How to start local development environment dependencies

```bash
docker compose --profile devcontainer-deps up -d --build
```

```bash
docker compose --profile devcontainer-deps down --remove-orphans
```

#### Run tests

```bash
/bin/sh ./run-tests-local.sh
```

A coverage report will be exported at `.coverage/report.html`

```bash
open ".coverage/report.html"
```

#### Simulate deployment stages

```bash
/bin/sh ./run-tests-ci.sh
```

### Graphic tools

- database administration (dbadmin): http://localhost:9001
- smtp (mailpit): http://localhost:9002


## API clients generation

Generate clients from open api:

```bash
docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate \
    -i /local/openapi.yaml \
    -g <language> \
    -o /local/client/<language>
```

For example for php:

```bash
docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate \
    -i /local/openapi.yaml \
    -g php \
    -o /local/client/php
```

Then you will find the generated client in the directory client/php in the root path of this repository.
