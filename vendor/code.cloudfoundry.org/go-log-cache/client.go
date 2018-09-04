package logcache

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"code.cloudfoundry.org/go-log-cache/rpc/logcache_v1"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"
)

// Client reads from LogCache via the RESTful or gRPC API.
type Client struct {
	addr string

	httpClient       HTTPClient
	grpcClient       logcache_v1.EgressClient
	promqlGrpcClient logcache_v1.PromQLQuerierClient
}

// NewIngressClient creates a Client.
func NewClient(addr string, opts ...ClientOption) *Client {
	c := &Client{
		addr: addr,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	for _, o := range opts {
		o.configure(c)
	}

	return c
}

// ClientOption configures the LogCache client.
type ClientOption interface {
	configure(client interface{})
}

// clientOptionFunc enables regular functions to be a ClientOption.
type clientOptionFunc func(client interface{})

// configure Implements clientOptionFunc.
func (f clientOptionFunc) configure(client interface{}) {
	f(client)
}

// HTTPClient is an interface that represents a http.Client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// WithHTTPClient sets the HTTP client. It defaults to a client that timesout
// after 5 seconds.
func WithHTTPClient(h HTTPClient) ClientOption {
	return clientOptionFunc(func(c interface{}) {
		switch c := c.(type) {
		case *Client:
			c.httpClient = h
		case *ShardGroupReaderClient:
			c.httpClient = h
		default:
			panic("unknown type")
		}
	})
}

// WithViaGRPC enables gRPC instead of HTTP/1 for reading from LogCache.
func WithViaGRPC(opts ...grpc.DialOption) ClientOption {
	return clientOptionFunc(func(c interface{}) {
		switch c := c.(type) {
		case *Client:
			conn, err := grpc.Dial(c.addr, opts...)
			if err != nil {
				panic(fmt.Sprintf("failed to dial via gRPC: %s", err))
			}

			c.grpcClient = logcache_v1.NewEgressClient(conn)
			c.promqlGrpcClient = logcache_v1.NewPromQLQuerierClient(conn)
		case *ShardGroupReaderClient:
			conn, err := grpc.Dial(c.addr, opts...)
			if err != nil {
				panic(fmt.Sprintf("failed to dial via gRPC: %s", err))
			}

			c.grpcClient = logcache_v1.NewShardGroupReaderClient(conn)
		default:
			panic("unknown type")
		}
	})
}

// Read queries the LogCache and returns the given envelopes. To override any
// query defaults (e.g., end time), use the according option.
func (c *Client) Read(
	ctx context.Context,
	sourceID string,
	start time.Time,
	opts ...ReadOption,
) ([]*loggregator_v2.Envelope, error) {
	if c.grpcClient != nil {
		return c.grpcRead(ctx, sourceID, start, opts)
	}

	u, err := url.Parse(c.addr)
	if err != nil {
		return nil, err
	}
	u.Path = "v1/read/" + sourceID
	q := u.Query()
	q.Set("start_time", strconv.FormatInt(start.UnixNano(), 10))

	// allow the given options to configure the URL.
	for _, o := range opts {
		o(u, q)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var r logcache_v1.ReadResponse
	if err := jsonpb.Unmarshal(resp.Body, &r); err != nil {
		return nil, err
	}

	return r.Envelopes.Batch, nil
}

// ReadOption configures the URL that is used to submit the query. The
// RawQuery is set to the decoded query parameters after each option is
// invoked.
type ReadOption func(u *url.URL, q url.Values)

// WithEndTime sets the 'end_time' query parameter to the given time. It
// defaults to empty, and therefore the end of the cache.
func WithEndTime(t time.Time) ReadOption {
	return func(u *url.URL, q url.Values) {
		q.Set("end_time", strconv.FormatInt(t.UnixNano(), 10))
	}
}

// WithLimit sets the 'limit' query parameter to the given value. It
// defaults to empty, and therefore 100 envelopes.
func WithLimit(limit int) ReadOption {
	return func(u *url.URL, q url.Values) {
		q.Set("limit", strconv.Itoa(limit))
	}
}

// WithEnvelopeTypes sets the 'envelope_types' query parameter to the given
// value. It defaults to empty, and therefore any envelope type.
func WithEnvelopeTypes(t ...logcache_v1.EnvelopeType) ReadOption {
	return func(u *url.URL, q url.Values) {
		for _, v := range t {
			q.Add("envelope_types", v.String())
		}
	}
}

// WithDescending set the 'descending' query parameter to true. It defaults to
// false, yielding ascending order.
func WithDescending() ReadOption {
	return func(u *url.URL, q url.Values) {
		q.Set("descending", "true")
	}
}

func (c *Client) grpcRead(ctx context.Context, sourceID string, start time.Time, opts []ReadOption) ([]*loggregator_v2.Envelope, error) {
	u := &url.URL{}
	q := u.Query()
	// allow the given options to configure the URL.
	for _, o := range opts {
		o(u, q)
	}

	req := &logcache_v1.ReadRequest{
		SourceId:  sourceID,
		StartTime: start.UnixNano(),
	}

	if v, ok := q["limit"]; ok {
		req.Limit, _ = strconv.ParseInt(v[0], 10, 64)
	}

	if v, ok := q["end_time"]; ok {
		req.EndTime, _ = strconv.ParseInt(v[0], 10, 64)
	}

	if v, ok := q["envelope_types"]; ok {
		req.EnvelopeTypes = []logcache_v1.EnvelopeType{
			logcache_v1.EnvelopeType(logcache_v1.EnvelopeType_value[v[0]]),
		}
	}

	if _, ok := q["descending"]; ok {
		req.Descending = true
	}

	resp, err := c.grpcClient.Read(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Envelopes.Batch, nil
}

// Meta returns meta information from the entire LogCache.
func (c *Client) Meta(ctx context.Context) (map[string]*logcache_v1.MetaInfo, error) {
	if c.grpcClient != nil {
		return c.grpcMeta(ctx)
	}

	u, err := url.Parse(c.addr)
	if err != nil {
		return nil, err
	}

	u.Path = "/v1/meta"

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var metaResponse logcache_v1.MetaResponse
	if err := jsonpb.Unmarshal(resp.Body, &metaResponse); err != nil {
		return nil, err
	}

	return metaResponse.Meta, nil
}

func (c *Client) grpcMeta(ctx context.Context) (map[string]*logcache_v1.MetaInfo, error) {
	resp, err := c.grpcClient.Meta(ctx, &logcache_v1.MetaRequest{})
	if err != nil {
		return nil, err
	}

	return resp.Meta, nil
}

// PromQLOption configures the URL that is used to submit the query. The
// RawQuery is set to the decoded query parameters after each option is
// invoked.
type PromQLOption func(u *url.URL, q url.Values)

// WithPromQLTime returns a PromQLOption that configures the 'time' query
// parameter for a PromQL query.
func WithPromQLTime(t time.Time) PromQLOption {
	return func(u *url.URL, q url.Values) {
		q.Set("time", strconv.FormatInt(t.UnixNano(), 10))
	}
}

// PromQL issues a PromQL query against Log Cache data.
func (c *Client) PromQL(
	ctx context.Context,
	query string,
	opts ...PromQLOption,
) (*logcache_v1.PromQL_QueryResult, error) {
	if c.promqlGrpcClient != nil {
		return c.grpcPromQL(ctx, query, opts)
	}

	u, err := url.Parse(c.addr)
	if err != nil {
		return nil, err
	}
	u.Path = "/v1/promql"
	q := u.Query()
	q.Set("query", query)

	// allow the given options to configure the URL.
	for _, o := range opts {
		o(u, q)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var promQLResponse logcache_v1.PromQL_QueryResult
	if err := jsonpb.Unmarshal(resp.Body, &promQLResponse); err != nil {
		return nil, err
	}

	return &promQLResponse, nil
}

func (c *Client) grpcPromQL(ctx context.Context, query string, opts []PromQLOption) (*logcache_v1.PromQL_QueryResult, error) {
	u := &url.URL{}
	q := u.Query()
	// allow the given options to configure the URL.
	for _, o := range opts {
		o(u, q)
	}

	req := &logcache_v1.PromQL_InstantQueryRequest{
		Query: query,
	}

	if v, ok := q["time"]; ok {
		req.Time, _ = strconv.ParseInt(v[0], 10, 64)
	}

	resp, err := c.promqlGrpcClient.InstantQuery(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
