package main_test

// https://docs.cloudfoundry.org/cf-cli/develop-cli-plugins.html
// https://github.com/cloudfoundry/cli/blob/master/plugin/plugin_examples/test_rpc_server_example/test_rpc_server_example_test.go
// https://github.com/cloudfoundry/cli/blob/master/plugin/plugin_examples/basic_plugin.go
import (
	"bytes"
	"os/exec"
	"testing"
	"time"

	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/go_test"

	"code.cloudfoundry.org/cli/cf/util/testhelpers/rpcserver"
	"code.cloudfoundry.org/cli/cf/util/testhelpers/rpcserver/rpcserverfakes"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const validPluginPath = "./service_health_cli_plugin.exe"

func TestIndicatorRegistryAgent(t *testing.T) {
	t.Run("service-health displays the usage text when no args are passed", func(t *testing.T) {
		g := NewGomegaWithT(t)

		rpcHandlers := new(rpcserverfakes.FakeHandlers)
		rpcHandlers.GetServiceStub = func(serviceInstance string, retVal *plugin_models.GetService_Model) error {
			retVal = &plugin_models.GetService_Model{Name: "myService", Guid: "myGuid"}
			return nil
		}
		g.Expect(buffer.String()).To(ContainSubstring("cf service-health SERVICE_INSTANCE"))
		ts.Stop()
	})

	t.Run("service-health SERVICE_INSTANCE displays the indicator document for the service instance", func(t *testing.T) {

		// ApiEndpointStub        func(args string, retVal *string) error
	})
}

func InvokePlugin(rpcHandlers rpcserverfakes.FakeHandlers, args ...string) (int, string, string) {
	rpcHandlers.IsMinCliVersionStub = func(args string, retVal *bool) error {
		*retVal = true
		return nil
	}
	ts, err := rpcserver.NewTestRPCServer(rpcHandlers)

	err = ts.Start()

	binPath, err := go_test.Build("./", "-race")
	if err != nil {
		panic(err)
	}
	args := []string{ts.Port(), "service-health"}
	buffer := bytes.NewBuffer(nil)
	errBuffer := bytes.NewBuffer(nil)

	_, err = gexec.Start(exec.Command(binPath, args...), buffer, errBuffer)
	time.Sleep(time.Second * 2)
	// EventuallyWithOffset(1, session).Should(gexec.Exit())
	// session.Wait()

	g.Expect(err).ToNot(HaveOccurred())
}
