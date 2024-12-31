package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/veksh/terraform-provider-godaddy-dns/internal/model"
	"github.com/veksh/terraform-provider-godaddy-dns/libs/ratelimiter"
)

// also: https://github.com/go-resty/resty

// to view actual records
// curlie -v GET "https://api.godaddy.com/v1/domains/veksh.in/records" -H "Authorization: sso-key $GODADDY_API_KEY:$GODADDY_API_SECRET"

const (
	HTTP_TIMEOUT = 10
	// burst RL: not currently used
	HTTP_RPS   = 1
	HTTP_BURST = 60
	// window RL: window size, max requests per window
	HTTP_RATE_WINDOW = time.Duration(60) * time.Second
	HTTP_RATE_RPW    = 60
	DOMAINS_URL      = "/v1/domains/"
)

var _ model.DNSApiClient = Client{}

// mb also http client here
type Client struct {
	apiURL     string
	key        string
	secret     string
	httpClient http.Client
}

func NewClient(apiURL string, key string, secret string) (*Client, error) {
	// t := http.DefaultTransport.(*http.Transport).Clone()
	httpTransport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: HTTP_TIMEOUT * time.Second}).DialContext,
		TLSHandshakeTimeout:   HTTP_TIMEOUT * time.Second,
		ResponseHeaderTimeout: HTTP_TIMEOUT * time.Second,
	}
	// TODO: mb make it pluggable as a parameter
	// rateLimiter, err := ratelimiter.NewBucketRL(HTTP_RPS, HTTP_BURST)
	rateLimiter, err := ratelimiter.NewWindowRL(HTTP_RATE_WINDOW, HTTP_RATE_RPW)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create rate limiter")
	}
	httpClient := http.Client{
		Transport: &rateLimitedHTTPTransport{
			limiter: rateLimiter,
			next:    httpTransport,
		},
	}
	return &Client{
		apiURL:     apiURL,
		key:        key,
		secret:     secret,
		httpClient: httpClient,
	}, nil
}

// see API docs: https://developer.godaddy.com/doc/endpoint/domains/
// ok for both Get and Add (patch), Put (replace) is partial (no type and name)
// first 4 are always present, rest are only for MX (priority) and SRV
type apiDNSRecord struct {
	Type     string `json:"type,omitempty"`
	Name     string `json:"name,omitempty"`
	Data     string `json:"data"`
	TTL      uint32 `json:"ttl"`
	Priority uint16 `json:"priority,omitempty"`
	Service  string `json:"service,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Port     uint16 `json:"port,omitempty"`
	Weight   uint16 `json:"weight,omitempty"`
}

type apiErrorResponce struct {
	Error   string `json:"code"`    // like "INVALID_VALUE_ENUM"
	Message string `json:"message"` // like "type not any of: A, ..."
}

func (c Client) makeRecordsRequest(ctx context.Context, path string, method string, body io.Reader) (*http.Response, error) {

	requestURL, _ := url.JoinPath(c.apiURL, DOMAINS_URL, path)

	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create request")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("sso-key %s:%s", c.key, c.secret))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "http request error")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		var errRes apiErrorResponce
		if err = json.NewDecoder(resp.Body).Decode(&errRes); err == nil {
			return nil, errors.New("api error: " + errRes.Message)
		}
		return nil, fmt.Errorf("bad http reply status (%s)", resp.Status)
	}
	return resp, nil
}

// in real API call
// - name and then type are optional (to get all records of type or just all records)
// - there are also "offset" and "limit" in query params for paged output
func (c Client) GetRecords(ctx context.Context, rDomain model.DNSDomain,
	rType model.DNSRecordType, rName model.DNSRecordName) ([]model.DNSRecord, error) {

	rPath, _ := url.JoinPath(string(rDomain), "records", string(rType), string(rName))

	resp, err := c.makeRecordsRequest(ctx, rPath, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responceRecords []apiDNSRecord
	err = json.NewDecoder(resp.Body).Decode(&responceRecords)
	if err != nil {
		return nil, errors.Wrap(err, "cannot decode json reply")
	}
	res := make([]model.DNSRecord, 0, len(responceRecords))
	for _, rr := range responceRecords {
		res = append(res, model.DNSRecord{
			Type:     model.DNSRecordType(rr.Type),
			Name:     model.DNSRecordName(rr.Name),
			Data:     model.DNSRecordData(rr.Data),
			TTL:      model.DNSRecordTTL(rr.TTL),
			Priority: model.DNSRecordPrio(rr.Priority),
			Protocol: model.DNSRecordSRVProto(rr.Protocol),
			Service:  model.DNSRecordSRVService(rr.Service),
			Port:     model.DNSRecordSRVPort(rr.Port),
			Weight:   model.DNSRecordSRVWeight(rr.Weight),
		})
	}
	return res, nil
}

// create (add) records for rType+rName
// existing are staying in place; there could be several records for type + name (eg MX)
func (c Client) AddRecords(ctx context.Context, rDomain model.DNSDomain,
	records []model.DNSRecord) error {

	rPath, _ := url.JoinPath(string(rDomain), "records")

	recs := make([]apiDNSRecord, 0, len(records))
	for _, mr := range records {
		rec := apiDNSRecord{
			Name: string(mr.Name),
			Type: string(mr.Type),
			Data: string(mr.Data),
			TTL:  uint32(mr.TTL),
		}
		if mr.Type == model.REC_MX || mr.Type == model.REC_SRV {
			rec.Priority = uint16(mr.Priority)
		}
		if mr.Type == model.REC_SRV {
			rec.Protocol = string(mr.Protocol)
			rec.Service = string(mr.Service)
			rec.Port = uint16(mr.Port)
			rec.Weight = uint16(mr.Weight)
		}
		recs = append(recs, rec)
	}

	jsonData, err := json.Marshal(&recs)
	if err != nil {
		return errors.Wrap(err, "cannot marshal json")
	}

	resp, err := c.makeRecordsRequest(ctx, rPath, http.MethodPatch, bytes.NewReader(jsonData))
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	return nil
}

// replace all records for rType+rName with the given
//   - there could be several records with the same type + name (eg MX)
//     and there is no way to update just one: they all get replaced
func (c Client) SetRecords(ctx context.Context, rDomain model.DNSDomain,
	rType model.DNSRecordType, rName model.DNSRecordName, records []model.DNSUpdateRecord) error {

	rPath, _ := url.JoinPath(string(rDomain), "records", string(rType), string(rName))

	recs := make([]apiDNSRecord, 0, len(records))
	for _, mr := range records {
		rec := apiDNSRecord{
			Data: string(mr.Data),
			TTL:  uint32(mr.TTL),
		}
		if rType == model.REC_MX {
			rec.Priority = uint16(mr.Priority)
		}
		if rType == model.REC_SRV {
			rec.Priority = uint16(mr.Priority)
			rec.Protocol = string(mr.Protocol)
			rec.Service = string(mr.Service)
			rec.Port = uint16(mr.Port)
			rec.Weight = uint16(mr.Weight)
		}
		recs = append(recs, rec)
	}

	jsonData, err := json.Marshal(&recs)
	if err != nil {
		return errors.Wrap(err, "cannot marshal json")
	}

	resp, err := c.makeRecordsRequest(ctx, rPath, http.MethodPut, bytes.NewReader(jsonData))
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	return nil
}

// delete all records for this type + name (no way to delete e.g. only 1 MX)
func (c Client) DelRecords(ctx context.Context, rDomain model.DNSDomain, rType model.DNSRecordType, rName model.DNSRecordName) error {

	rPath, _ := url.JoinPath(string(rDomain), "records", string(rType), string(rName))

	resp, err := c.makeRecordsRequest(ctx, rPath, http.MethodDelete, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	return nil
}
