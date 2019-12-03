package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	v1 "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/apis/indicatordocument/v1"
)

func TestBoshE2e(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	tokenBytes, err := exec.Command("cf", "oauth-token").Output()
	if err != nil {
		t.Fatal(err)
	}
	tokenString := trimBearer(strings.TrimSuffix(string(tokenBytes), "\n"))

	t.Run("agents put diego docs in registry", func(t *testing.T) {
		g := NewGomegaWithT(t)

		url := fmt.Sprintf(
			"https://indicator-protocol-acceptance-proxy.madlamp.cf-denver.com"+
				"/v1/indicator-documents/?token=%s", tokenString,
		)

		resp, err := http.Get(url)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(resp.StatusCode).To(Equal(200))
		var docs []v1.IndicatorDocument
		err = json.NewDecoder(resp.Body).Decode(&docs)
		g.Expect(err).ToNot(HaveOccurred())

		for _, doc := range docs {
			if doc.Spec.Product.Name == "diego" {
				return
			}
		}
		t.Error("Did not find any documents with product name `diego`")
	})
}
