.PHONY: testacc gen

default: testacc

# Run acceptance tests
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m


ROOT := $(PWD)
TMPDIR := $(shell mktemp -d)

# Download the OpenAPI spec
gen/openapi.json:
	# Update the OpenAPI spec to remove unsupported types
	cd $(TMPDIR); \
		curl -sLo openapi.json "https://raw.githubusercontent.com/tursodatabase/turso-docs/refs/heads/main/api-reference/openapi.json"; \
		jq -r '.components.schemas.Extensions |= .oneOf[0]' openapi.json > openapi.1.json; \
		jq -r '.components.schemas.Group |= .allOf[0]' openapi.1.json > openapi.2.json; \
		jq -r '.paths["/v1/organizations/{organizationName}/groups/{groupName}/unarchive"].post.operationId = "unarchiveGroup"' openapi.2.json > openapi.3.json; \
		jq -r '.components.schemas.Database.properties.schema |= . + {nullable: true}' openapi.3.json > openapi.4.json; \
		cp openapi.4.json $(ROOT)/gen/openapi.json

# Generate provider code from OpenAPI spec
gen: gen/openapi.json
	# Generate the provider code
	@go install github.com/hashicorp/terraform-plugin-codegen-openapi/cmd/tfplugingen-openapi@latest
	@go install github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework@latest
	tfplugingen-openapi generate \
		--config ./gen/generator_config.yml \
		--output ./gen/provider-code-spec.json \
		./gen/openapi.json
	tfplugingen-framework generate resources \
		--input ./gen/provider-code-spec.json \
		--output ./internal
	tfplugingen-framework generate data-sources \
		--input ./gen/provider-code-spec.json \
		--output ./internal

	# Generate the client code
	go generate ./...
	go mod tidy
	go fmt ./...
