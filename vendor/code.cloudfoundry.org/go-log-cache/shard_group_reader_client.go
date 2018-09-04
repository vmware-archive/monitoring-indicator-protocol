package logcache

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/go-log-cache/rpc/logcache_v1"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/golang/protobuf/jsonpb"
)

// ShardGroupReaderClient reads and interacts from LogCache via the RESTful or gRPC
// Group API.
type ShardGroupReaderClient struct {
	addr string

	unmarshaler *jsonpb.Unmarshaler
	httpClient  HTTPClient
	grpcClient  logcache_v1.ShardGroupReaderClient
}

// NewShardGroupReaderClient creates a ShardGroupReaderClient.
func NewShardGroupReaderClient(addr string, opts ...ClientOption) *ShardGroupReaderClient {
	c := &ShardGroupReaderClient{
		addr: addr,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		unmarshaler: &jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		},
	}

	for _, o := range opts {
		o.configure(c)
	}

	return c
}

// BuildReader is used to create a Reader (useful for things like Walk) with a
// RequesterID. It simply wraps the ShardGroupReaderClient.Read method.
func (c *ShardGroupReaderClient) BuildReader(requesterID uint64) Reader {
	return Reader(func(
		ctx context.Context,
		name string,
		start time.Time,
		opts ...ReadOption,
	) ([]*loggregator_v2.Envelope, error) {
		return c.Read(ctx, name, start, requesterID, opts...)
	})
}

// Read queries the LogCache and returns the given envelopes. To override any
// query defaults (e.g., end time), use the according option.
func (c *ShardGroupReaderClient) Read(
	ctx context.Context,
	name string,
	start time.Time,
	requesterID uint64,
	opts ...ReadOption,
) ([]*loggregator_v2.Envelope, error) {
	if c.grpcClient != nil {
		return c.grpcRead(ctx, name, start, requesterID, opts)
	}

	u, err := url.Parse(c.addr)
	if err != nil {
		return nil, err
	}
	u.Path = "v1/shard_group/" + name
	q := u.Query()
	q.Set("start_time", strconv.FormatInt(start.UnixNano(), 10))
	q.Set("requester_id", strconv.FormatUint(requesterID, 10))

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

	var r logcache_v1.ShardGroupReadResponse
	if err := c.unmarshaler.Unmarshal(resp.Body, &r); err != nil {
		return nil, err
	}

	return r.Envelopes.Batch, nil
}

func (c *ShardGroupReaderClient) grpcRead(
	ctx context.Context,
	name string,
	start time.Time,
	requesterID uint64,
	opts []ReadOption,
) ([]*loggregator_v2.Envelope, error) {
	u := &url.URL{}
	q := u.Query()
	// allow the given options to configure the URL.
	for _, o := range opts {
		o(u, q)
	}

	req := &logcache_v1.ShardGroupReadRequest{
		Name:        name,
		RequesterId: requesterID,
		StartTime:   start.UnixNano(),
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

	resp, err := c.grpcClient.Read(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Envelopes.Batch, nil
}

// SetGroup adds a group of sourceIDs to the given group. If the group doesn't
// exist, then it is created. If the group already has the given sub-group,
// then it is a NOP.
func (c *ShardGroupReaderClient) SetShardGroup(
	ctx context.Context,
	name string,
	sourceIDs ...string,
) error {
	if c.grpcClient != nil {
		return c.grpcSetGroup(ctx, name, sourceIDs)
	}

	u, err := url.Parse(c.addr)
	if err != nil {
		return err
	}
	u.Path = fmt.Sprintf("v1/shard_group/%s", name)

	marshalled, err := (&jsonpb.Marshaler{}).MarshalToString(
		&logcache_v1.GroupedSourceIds{
			SourceIds: sourceIDs,
		},
	)

	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", u.String(), strings.NewReader(marshalled))
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	return nil
}

func (c *ShardGroupReaderClient) grpcSetGroup(
	ctx context.Context,
	name string,
	sourceIDs []string,
) error {
	_, err := c.grpcClient.SetShardGroup(ctx, &logcache_v1.SetShardGroupRequest{
		Name: name,
		SubGroup: &logcache_v1.GroupedSourceIds{
			SourceIds: sourceIDs,
		},
	})
	return err
}

// GroupMeta gives the information about given group.
type GroupMeta struct {
	// SubGroups are the collection of sub-groups that the overall group is
	// managing.
	SubGroups []SubGroup

	// RequesterIDs is the active list of requesters that are currently
	// reading from the group.
	RequesterIDs []uint64
}

// SubGroup is a group of SourceIDs that are read together.
type SubGroup struct {
	// SourceIDs are the SourceIDs that are read from the group.
	SourceIDs []string
}

// Group returns the meta information about a group.
func (c *ShardGroupReaderClient) ShardGroup(ctx context.Context, name string) (GroupMeta, error) {
	if c.grpcClient != nil {
		return c.grpcGroup(ctx, name)
	}

	u, err := url.Parse(c.addr)
	if err != nil {
		return GroupMeta{}, err
	}
	u.Path = fmt.Sprintf("v1/shard_group/%s/meta", name)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return GroupMeta{}, err
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return GroupMeta{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GroupMeta{}, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return GroupMeta{}, err
	}

	gresp := logcache_v1.ShardGroupResponse{}
	if err := json.Unmarshal(data, &gresp); err != nil {
		return GroupMeta{}, err
	}

	gm := GroupMeta{
		RequesterIDs: gresp.RequesterIds,
	}

	for _, group := range gresp.GetSubGroups() {
		gm.SubGroups = append(gm.SubGroups, SubGroup{
			SourceIDs: group.SourceIds,
		})
	}

	return gm, nil
}

func (c *ShardGroupReaderClient) grpcGroup(ctx context.Context, name string) (GroupMeta, error) {
	resp, err := c.grpcClient.ShardGroup(ctx, &logcache_v1.ShardGroupRequest{
		Name: name,
	})

	if err != nil {
		return GroupMeta{}, err
	}

	gm := GroupMeta{
		RequesterIDs: resp.RequesterIds,
	}

	for _, group := range resp.GetSubGroups() {
		gm.SubGroups = append(gm.SubGroups, SubGroup{
			SourceIDs: group.SourceIds,
		})
	}

	return gm, nil
}
