
# MultiDialogo - MailCulator server

## Requirements

- docker
- docker compose v2
- git

## Scripts

### How to start/stop local development environment

```bash
/bin/sh ./local/start-local-env.sh
```

```bash
/bin/sh ./local/stop-local-env.sh
```

### Run tests locally

```bash
/bin/sh ./local/test.sh
```

### API clients

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

