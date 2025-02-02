package mesh_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
	_ "github.com/kumahq/kuma/pkg/plugins/runtime/gateway/register"
)

// GatewayGenerateor is a ResourceGenerator that creates GatewayResource objects.
type GatewayGenerator func() *GatewayResource

func (g GatewayGenerator) New() model.Resource {
	if g != nil {
		return g()
	}

	return nil
}

var _ = Describe("Gateway", func() {
	DescribeValidCases(
		GatewayGenerator(NewGatewayResource),
		Entry("HTTPS listener", `
type: Gateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - hostname: www-1.example.com
    port: 443
    protocol: HTTP
    tags:
      name: https`,
		),
	)

	DescribeErrorCases(
		GatewayGenerator(NewGatewayResource),
		ErrorCase("doesn't have any selectors",
			validators.Violation{
				Field:   `selectors`,
				Message: `must have at least one element`,
			}, `
type: Gateway
name: gateway
mesh: default
selectors:
tags:
  product: edge
conf:
  listeners:
  - port: 443
    protocol: HTTPS
    tags:
      name: https
`),

		ErrorCase("has a service tag",
			validators.Violation{
				Field:   `tags["kuma.io/service"]`,
				Message: `tag name must not be "kuma.io/service"`,
			}, `
type: Gateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
  kuma.io/service: gateway
conf:
  listeners:
  - port: 443
    protocol: HTTP
    tags:
      name: https
`),

		ErrorCase("doesn't have a configuration spec",
			validators.Violation{
				Field:   "conf",
				Message: "cannot be empty",
			}, `
type: Gateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
`),

		ErrorCase("has an invalid port",
			validators.Violation{
				Field:   "conf.listeners[0].port",
				Message: "port must be in the range [1, 65535]",
			}, `
type: Gateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - protocol: HTTP
    tags:
      name: https
`),

		ErrorCase("has an empty protocol",
			validators.Violation{
				Field:   "conf.listeners[0].protocol",
				Message: "cannot be empty",
			}, `
type: Gateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - port: 99
    tags:
      name: https
`),

		ErrorCase("has an empty TLS mode",
			validators.Violation{
				Field:   "conf.listeners[0].tls.mode",
				Message: "cannot be empty",
			}, `
type: Gateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - protocol: HTTPS
    port: 99
    tags:
      name: https
    tls:
      options:
`),

		ErrorCase("has a passthrough TLS secret",
			validators.Violation{
				Field:   "conf.listeners[0].tls.certificate",
				Message: "must be empty in TLS passthrough mode",
			}, `
type: Gateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - protocol: HTTPS
    port: 99
    tags:
      name: https
    tls:
      mode: PASSTHROUGH
      certificate:
        secret: foo
`),

		ErrorCase("is missing a TLS termination secret",
			validators.Violation{
				Field:   "conf.listeners[0].tls.certificate",
				Message: "cannot be empty in TLS termination mode",
			}, `
type: Gateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - protocol: HTTPS
    port: 99
    tags:
      name: https
    tls:
      mode: TERMINATE
`),

		ErrorCase("has an invalid wildcard",
			validators.Violation{
				Field:   "conf.listeners[0].hostname",
				Message: "invalid wildcard domain",
			}, `
type: Gateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - hostname: "*.foo.*.example.com"
    protocol: HTTP
    port: 99
    tags:
      name: https
`),

		ErrorCase("has an invalid hostname",
			validators.Violation{
				Field:   "conf.listeners[0].hostname",
				Message: "invalid hostname",
			}, `
type: Gateway
name: gateway
mesh: default
selectors:
  - match:
      kuma.io/service: gateway
tags:
  product: edge
conf:
  listeners:
  - hostname: "foo.example$.com"
    protocol: HTTP
    port: 99
    tags:
      name: https
`),
	)
})
