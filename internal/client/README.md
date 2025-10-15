# OpenAPI Code Generation

This directory contains code generated from the Coolify OpenAPI specification.

## Overview

The code in this directory is automatically generated using [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) from the OpenAPI spec located at `/client.yaml`. Do not modify any generated files manually as changes will be overwritten.

## Generation Process

The code generation is configured via:

1. `generate.go` - Contains the go:generate directive to run oapi-codegen
2. `oapi-codegen.yaml` - Contains the configuration for oapi-codegen
3. `overlay.yaml` - Contains an overlay specification to make modifications to the yaml spec before generating
4. `openapi.yaml` - The original OpenAPI specification file

To regenerate the code, run:

```
go generate ./...
```
