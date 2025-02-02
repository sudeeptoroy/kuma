package api_server_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	config "github.com/kumahq/kuma/pkg/config/api-server"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("Global Insights Endpoints", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	var stop chan struct{}

	BeforeEach(func() {
		core.Now = func() time.Time {
			now, _ := time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")
			return now
		}

		resourceStore = memory.NewStore()

		metrics, err := metrics.NewMetrics("Standalone")
		Expect(err).ToNot(HaveOccurred())

		apiServer = createTestApiServer(resourceStore, config.DefaultApiServerConfig(), true, metrics)

		client := resourceApiClient{
			address: apiServer.Address(),
			path:    "/global-insights",
		}

		stop = make(chan struct{})

		go func() {
			defer GinkgoRecover()
			Expect(apiServer.Start(stop)).To(Succeed())
		}()

		waitForServer(&client)
	}, 5)

	AfterEach(func() {
		close(stop)
		core.Now = time.Now
	})

	BeforeEach(func() {
		Expect(resourceStore.Create(
			context.Background(),
			system.NewZoneResource(),
			store.CreateByKey("zone-1", core_model.NoMesh),
		)).To(Succeed())

		Expect(resourceStore.Create(
			context.Background(),
			system.NewZoneResource(),
			store.CreateByKey("zone-2", core_model.NoMesh),
		)).To(Succeed())

		Expect(resourceStore.Create(
			context.Background(),
			core_mesh.NewZoneIngressResource(),
			store.CreateByKey("zone-ingress-1", core_model.NoMesh),
		)).To(Succeed())

		Expect(resourceStore.Create(
			context.Background(),
			core_mesh.NewMeshResource(),
			store.CreateByKey("mesh-1", core_model.NoMesh),
		)).To(Succeed())

		Expect(resourceStore.Create(
			context.Background(),
			core_mesh.NewMeshResource(),
			store.CreateByKey("mesh-2", core_model.NoMesh),
		)).To(Succeed())

		Expect(resourceStore.Create(
			context.Background(),
			core_mesh.NewMeshResource(),
			store.CreateByKey("mesh-3", core_model.NoMesh),
		)).To(Succeed())
	})

	globalInsightsJSON := `
{
  "type": "GlobalInsights",
  "creationTime": "2018-07-17T16:05:36.995Z",
  "meshes": {
    "total": 3
  },
  "zones": {
    "total": 2
  },
  "zoneIngresses": {
    "total": 1
  }
}
`

	Describe("On GET", func() {
		It("should return an existing resource", func() {
			// when
			response, err := http.Get("http://" + apiServer.Address() + "/global-insights")
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(response.StatusCode).To(Equal(200))
			body, err := ioutil.ReadAll(response.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(MatchJSON(globalInsightsJSON))
		})
	})
})
