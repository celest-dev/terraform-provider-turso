# Generating the provider

Follows the steps here: https://developer.hashicorp.com/terraform/plugin/code-generation/workflow-example

```bash
curl https://raw.githubusercontent.com/tursodatabase/turso-docs/main/api-reference/openapi.json --output openapi.json

tfplugingen-openapi generate \
    --config ./generator_config.yml \
    --output ./provider-code-spec.json \
    ./openapi.json

tfplugingen-framework generate resources \
  --input ./provider-code-spec.json \
  --output ./internal 

```