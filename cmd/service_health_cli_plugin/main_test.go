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

func TestIndicatorRegistryAgent(t *testing.T) {
	t.Run("service-health displays the usage text when no args are passed", func(t *testing.T) {
		g := NewGomegaWithT(t)

		rpcHandlers := new(rpcserverfakes.FakeHandlers)

		code, stdOut, stdErr := InvokePlugin(rpcHandlers)

		g.Expect(stdOut).To(ContainSubstring("cf service-health SERVICE_INSTANCE"))
		g.Expect(stdErr).To(Equal(""))
		g.Expect(code).To(Equal(1))
	})

	t.Run("service-health SERVICE_INSTANCE displays only the indicators for the service instance", func(t *testing.T) {
		g := NewGomegaWithT(t)

		rpcHandlers := new(rpcserverfakes.FakeHandlers)
		rpcHandlers.GetServiceStub = func(serviceInstance string, retVal *plugin_models.GetService_Model) error {
			retVal = &plugin_models.GetService_Model{Name: "myService", Guid: "myGuid"}
			return nil
		}
		rpcHandlers.ApiEndpointCalls(func(s string, i *string) error {
			*i = "localhost:47321"
			return nil
		})

		// TODO put registry up at 47321 with an indicator document for myService

		code, stdOut, stdErr := InvokePlugin(rpcHandlers, "myService")

		g.Expect(stdOut).To(ContainSubstring("cf service-health SERVICE_INSTANCE"))
		g.Expect(stdErr).To(Equal(""))
		g.Expect(code).To(Equal(0))

		// ApiEndpointStub        func(args string, retVal *string) error

	})
}


func InvokePlugin(rpcHandlers *rpcserverfakes.FakeHandlers, args ...string) (exitCode int, stdOut string, stdErr string) {
	rpcHandlers.IsMinCliVersionStub = func(args string, retVal *bool) error {
		*retVal = true
		return nil
	}
	ts, err := rpcserver.NewTestRPCServer(rpcHandlers)

	err = ts.Start()
	defer ts.Stop()

	binPath, err := go_test.Build("./", "-race")
	if err != nil {
		panic(err)
	}
	args = append([]string{ts.Port(), "service-health"}, args...)
	buffer := bytes.NewBuffer(nil)
	errBuffer := bytes.NewBuffer(nil)

	session, err := gexec.Start(exec.Command(binPath, args...), buffer, errBuffer)
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 2)

	return session.ExitCode(), buffer.String(), errBuffer.String()
}
