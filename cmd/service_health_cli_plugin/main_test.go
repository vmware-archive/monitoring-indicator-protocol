package main_test

// https://docs.cloudfoundry.org/cf-cli/develop-cli-plugins.html
// https://github.com/cloudfoundry/cli/blob/master/plugin/plugin_examples/test_rpc_server_example/test_rpc_server_example_test.go
// https://github.com/cloudfoundry/cli/blob/master/plugin/plugin_examples/basic_plugin.go

const validPluginPath = "./test_rpc_server_example.exe"

func TestIndicatorRegistryAgent(t *testing.T) {
	g := NewGomegaWithT(t)
rpcHandlers = new(rpcserverfakes.FakeHandlers)
		ts, err = rpcserver.NewTestRPCServer(rpcHandlers)

	binPath, err := go_test.Build("./", "-race")
	g.Expect(err).ToNot(HaveOccurred())

	t.Run("it sends indicator documents to the registry on an interval", func(t *testing.T) {
		g := NewGomegaWithT(t)

