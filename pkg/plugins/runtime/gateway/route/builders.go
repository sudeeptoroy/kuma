package route

import (
	"crypto/sha256"
	"fmt"
	"sort"

	envoy_config_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

type RouteConfigurer interface {
	Configure(*envoy_config_route.Route) error
}

type RouteBuilder struct {
	configurers []RouteConfigurer
}

func (r *RouteBuilder) Configure(opts ...RouteConfigurer) *RouteBuilder {
	r.configurers = append(r.configurers, opts...)
	return r
}

func (r *RouteBuilder) Build() (envoy.NamedResource, error) {
	route := &envoy_config_route.Route{
		Match: &envoy_config_route.RouteMatch{},
	}

	for _, c := range r.configurers {
		if err := c.Configure(route); err != nil {
			return nil, err
		}
	}

	return route, nil
}

type RouteConfigureFunc func(*envoy_config_route.Route) error

func (f RouteConfigureFunc) Configure(r *envoy_config_route.Route) error {
	if f != nil {
		return f(r)
	}

	return nil
}

type RouteMustConfigureFunc func(*envoy_config_route.Route)

func (f RouteMustConfigureFunc) Configure(r *envoy_config_route.Route) error {
	if f != nil {
		f(r)
	}

	return nil
}

// DestinationClusterName generates a unique cluster name for the
// destination. The destination must contain at least a service tag.
func DestinationClusterName(d Destination) (string, error) {
	serviceName := d.Destination[mesh_proto.ServiceTag]
	if serviceName == "" {
		return "", errors.Errorf("missing %s tag", mesh_proto.ServiceTag)
	}

	clusterName := serviceName

	if len(d.Destination) == 1 {
		return clusterName, nil
	}

	// If cluster is splitting the target service with selector tags,
	// hash the tag names to generate a unique cluster name. This is
	// a stateless uniqifier so that different xDS resource generators
	// can use the same tags to arrive at the same cluster name without
	// coordination.

	var keys []string
	for k := range d.Destination {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	h := sha256.New()

	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte(d.Destination[k]))
	}

	// The qualifier is 16 hex digits. Unscientifically balancing the length
	// of the hex against the likelihood of collisions.
	clusterName = fmt.Sprintf("%s-%x", serviceName, h.Sum(nil)[:8])

	return clusterName, nil
}
