package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kevynb/terraform-provider-technitium/internal/model"
	"github.com/pkg/errors"
)

const (
	HTTP_TIMEOUT               = 10
	DOMAINS_URL                = "/api/zones/records"
	ZONES_URL                  = "/api/zones"
	TERRAFORM_PROVIDER_COMMENT = "Managed by terraform"
)

const (
	StatusOK           = "ok"
	StatusError        = "error"
	StatusInvalidToken = "invalid-token"
)

var _ model.DNSApiClient = Client{}

type Client struct {
	apiURL     string
	token      string
	httpClient http.Client
}

func NewClient(apiURL string, token string, skipCertificateVerification bool) (*Client, error) {
	httpTransport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: HTTP_TIMEOUT * time.Second}).DialContext,
		TLSHandshakeTimeout:   HTTP_TIMEOUT * time.Second,
		ResponseHeaderTimeout: HTTP_TIMEOUT * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: skipCertificateVerification},
	}

	httpClient := http.Client{
		Transport: httpTransport,
	}
	return &Client{
		apiURL:     apiURL,
		token:      token,
		httpClient: httpClient,
	}, nil
}

type apiResponse struct {
	Status            string          `json:"status"`
	Response          apiResponseBody `json:"response,omitempty"`
	ErrorMessage      string          `json:"errorMessage,omitempty"`
	InnerErrorMessage string          `json:"innerErrorMessage,omitempty"`
}
type apiResponseBody struct {
	Records []apiDNSRecordResponseItem `json:"records"`
	Zone    apiResponseZone            `json:"zone"`
}
type apiResponseZone struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Internal bool   `json:"internal"`
	Disabled bool   `json:"disabled"`
}
type apiDNSRecordResponseItem struct {
	Type     string                        `json:"type,omitempty"`
	Domain   string                        `json:"name,omitempty"`
	Disabled bool                          `json:"disabled,omitempty"`
	TTL      uint32                        `json:"ttl"`
	Comments string                        `json:"comments,omitempty"`
	RData    apiDNSRecordResponseItemRdata `json:"rData,omitempty"`
}
type apiDNSRecordResponseItemRdata struct {
	ExpiryTTL                      uint32 `json:"expiryTtl,omitempty"`
	IPAddress                      string `json:"ipAddress,omitempty"`
	Ptr                            bool   `json:"ptr,omitempty"`
	CreatePtrZone                  bool   `json:"createPtrZone,omitempty"`
	UpdateSvcbHints                bool   `json:"updateSvcbHints,omitempty"`
	NameServer                     string `json:"nameServer,omitempty"`
	Glue                           string `json:"glue,omitempty"`
	CName                          string `json:"cname,omitempty"`
	PtrName                        string `json:"ptrName,omitempty"`
	Exchange                       string `json:"exchange,omitempty"`
	Preference                     uint16 `json:"preference,omitempty"`
	Text                           string `json:"text,omitempty"`
	SplitText                      bool   `json:"splitText,omitempty"`
	Mailbox                        string `json:"mailbox,omitempty"`
	TxtDomain                      string `json:"txtDomain,omitempty"`
	Priority                       uint16 `json:"priority,omitempty"`
	Weight                         uint16 `json:"weight,omitempty"`
	Port                           uint16 `json:"port,omitempty"`
	Target                         string `json:"target,omitempty"`
	NaptrOrder                     uint16 `json:"naptrOrder,omitempty"`
	NaptrPreference                uint16 `json:"naptrPreference,omitempty"`
	NaptrFlags                     string `json:"naptrFlags,omitempty"`
	NaptrServices                  string `json:"naptrServices,omitempty"`
	NaptrRegexp                    string `json:"naptrRegexp,omitempty"`
	NaptrReplacement               string `json:"naptrReplacement,omitempty"`
	DName                          string `json:"dName,omitempty"`
	KeyTag                         uint16 `json:"keyTag,omitempty"`
	Algorithm                      string `json:"algorithm,omitempty"`
	DigestType                     string `json:"digestType,omitempty"`
	Digest                         string `json:"digest,omitempty"`
	SshfpAlgorithm                 string `json:"sshfpAlgorithm,omitempty"`
	SshfpFingerprintType           string `json:"sshfpFingerprintType,omitempty"`
	SshfpFingerprint               string `json:"sshfpFingerprint,omitempty"`
	TlsaCertificateUsage           string `json:"tlsaCertificateUsage,omitempty"`
	TlsaSelector                   string `json:"tlsaSelector,omitempty"`
	TlsaMatchingType               string `json:"tlsaMatchingType,omitempty"`
	TlsaCertificateAssociationData string `json:"tlsaCertificateAssociationData,omitempty"`
	SvcPriority                    uint16 `json:"svcPriority,omitempty"`
	SvcTargetName                  string `json:"svcTargetName,omitempty"`
	SvcParams                      string `json:"svcParams,omitempty"`
	AutoIpv4Hint                   bool   `json:"autoIpv4Hint,omitempty"`
	AutoIpv6Hint                   bool   `json:"autoIpv6Hint,omitempty"`
	UriPriority                    uint16 `json:"uriPriority,omitempty"`
	UriWeight                      uint16 `json:"uriWeight,omitempty"`
	Uri                            string `json:"uri,omitempty"`
	Flags                          string `json:"flags,omitempty"`
	Tag                            string `json:"tag,omitempty"`
	Value                          string `json:"value,omitempty"`
	AName                          string `json:"aname,omitempty"`
	Forwarder                      string `json:"forwarder,omitempty"`
	ForwarderPriority              uint16 `json:"forwarderPriority,omitempty"`
	DnssecValidation               bool   `json:"dnssecValidation,omitempty"`
	ProxyType                      string `json:"proxyType,omitempty"`
	ProxyAddress                   string `json:"proxyAddress,omitempty"`
	ProxyPort                      uint16 `json:"proxyPort,omitempty"`
	ProxyUsername                  string `json:"proxyUsername,omitempty"`
	ProxyPassword                  string `json:"proxyPassword,omitempty"`
	AppName                        string `json:"appName,omitempty"`
	ClassPath                      string `json:"classPath,omitempty"`
	RecordData                     string `json:"data,omitempty"`
}

// apiErrorResponse is kept for potential future use
// type apiErrorResponse struct {
// 	Error   string `json:"code"`    // like "INVALID_VALUE_ENUM"
// 	Message string `json:"message"` // like "type not any of: A, ..."
// }

func (c Client) makeRecordsRequest(ctx context.Context, path string, method string, queryParams url.Values, formData url.Values, apiResponse *apiResponse) error {
	// Ensure the token is always set
	switch method {
	case http.MethodGet:
		if queryParams == nil {
			queryParams = url.Values{}
		}
		queryParams.Set("token", c.token)
	case http.MethodPost:
		if formData == nil {
			formData = url.Values{}
		}
		formData.Set("token", c.token)
	}

	var requestURL string
	var body io.Reader
	if method == http.MethodGet {
		requestURL = fmt.Sprintf("%s%s%s?%s", c.apiURL, DOMAINS_URL, path, queryParams.Encode())
	} else {
		requestURL = fmt.Sprintf("%s%s%s", c.apiURL, DOMAINS_URL, path)
		body = strings.NewReader(formData.Encode())
		print("\n\n", formData.Encode(), "\n\n")
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return errors.Wrap(err, "cannot create HTTP request")
	}

	if method == http.MethodPost {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "HTTP request error")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Parse response to check for API errors
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return errors.Wrap(err, "cannot decode JSON response into the provided structure")
	}

	if apiResponse.Status != StatusOK {
		logMessage := fmt.Sprintf("API error: %s", apiResponse.ErrorMessage)
		if apiResponse.InnerErrorMessage != "" {
			logMessage = fmt.Sprintf("%s (Inner: %s)", logMessage, apiResponse.InnerErrorMessage)
		}
		return errors.New(logMessage)
	}

	return nil
}

func (c Client) makeZonesRequest(ctx context.Context, path string, method string, queryParams url.Values, formData url.Values, apiResponse interface{}) error {
	// Ensure the token is always set
	switch method {
	case http.MethodGet:
		if queryParams == nil {
			queryParams = url.Values{}
		}
		queryParams.Set("token", c.token)
	case http.MethodPost:
		if formData == nil {
			formData = url.Values{}
		}
		formData.Set("token", c.token)
	}

	var requestURL string
	var body io.Reader
	if method == http.MethodGet {
		requestURL = fmt.Sprintf("%s%s%s?%s", c.apiURL, ZONES_URL, path, queryParams.Encode())
	} else {
		requestURL = fmt.Sprintf("%s%s%s", c.apiURL, ZONES_URL, path)
		body = strings.NewReader(formData.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return errors.Wrap(err, "cannot create HTTP request")
	}

	if method == http.MethodPost {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "HTTP request error")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Parse response to check for API errors
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return errors.Wrap(err, "cannot decode JSON response into the provided structure")
	}

	// Check for API errors - this assumes the response has Status field
	if responseMap, ok := apiResponse.(map[string]interface{}); ok {
		if status, exists := responseMap["status"]; exists && status != StatusOK {
			logMessage := "API error"
			if errorMsg, exists := responseMap["errorMessage"]; exists {
				logMessage = fmt.Sprintf("API error: %s", errorMsg)
			}
			if innerErrorMsg, exists := responseMap["innerErrorMessage"]; exists && innerErrorMsg != "" {
				logMessage = fmt.Sprintf("%s (Inner: %s)", logMessage, innerErrorMsg)
			}
			return errors.New(logMessage)
		}
	}

	return nil
}

// GetRecords retrieves all DNS records for a given domain name (zone is inferred automatically).
func (c Client) GetRecords(ctx context.Context, domain model.DNSRecordName) ([]model.DNSRecord, error) {
	params := url.Values{}
	if domain != "" {
		params.Add("domain", string(domain))
	}
	params.Add("listZone", "true")

	var apiResponse apiResponse
	err := c.makeRecordsRequest(ctx, "/get", http.MethodGet, params, nil, &apiResponse)
	if err != nil {
		return nil, err
	}

	res := make([]model.DNSRecord, len(apiResponse.Response.Records))
	for i, rr := range apiResponse.Response.Records {
		res[i] = mapAPIDNSRecordToDNSRecord(rr, apiResponse.Response.Zone.Name)
	}

	return res, nil
}

// AddRecord adds DNS record for a given domain.
func (c Client) AddRecord(ctx context.Context, record model.DNSRecord) error {
	formData := url.Values{
		"type":   {string(record.Type)},
		"domain": {string(record.Domain)},
		"ttl":    {fmt.Sprintf("%d", record.TTL)},
	}

	formData.Add("comments", TERRAFORM_PROVIDER_COMMENT)

	if record.ExpiryTTL > 0 {
		formData.Add("expiryTtl", fmt.Sprintf("%d", record.ExpiryTTL))
	}
	if record.IPAddress != "" {
		formData.Add("ipAddress", record.IPAddress)
	}
	if record.Ptr {
		formData.Add("ptr", "true")
	}
	if record.CreatePtrZone {
		formData.Add("createPtrZone", "true")
	}
	if record.UpdateSvcbHints {
		formData.Add("updateSvcbHints", "true")
	}
	if record.NameServer != "" {
		formData.Add("nameServer", record.NameServer)
	}
	if record.Glue != "" {
		formData.Add("glue", record.Glue)
	}
	if record.CName != "" {
		formData.Add("cname", record.CName)
	}
	if record.PtrName != "" {
		formData.Add("ptrName", record.PtrName)
	}
	if record.Exchange != "" {
		formData.Add("exchange", record.Exchange)
	}
	if record.Preference > 0 {
		formData.Add("preference", fmt.Sprintf("%d", record.Preference))
	}
	if record.Text != "" {
		formData.Add("text", record.Text)
	}
	if record.SplitText {
		formData.Add("splitText", "true")
	}
	if record.Mailbox != "" {
		formData.Add("mailbox", record.Mailbox)
	}
	if record.TxtDomain != "" {
		formData.Add("txtDomain", record.TxtDomain)
	}
	if record.Priority > 0 {
		formData.Add("priority", fmt.Sprintf("%d", record.Priority))
	}
	if record.Weight > 0 {
		formData.Add("weight", fmt.Sprintf("%d", record.Weight))
	}
	if record.Port > 0 {
		formData.Add("port", fmt.Sprintf("%d", record.Port))
	}
	if record.Target != "" {
		formData.Add("target", string(record.Target))
	}
	if record.NaptrOrder > 0 {
		formData.Add("naptrOrder", fmt.Sprintf("%d", record.NaptrOrder))
	}
	if record.NaptrPreference > 0 {
		formData.Add("naptrPreference", fmt.Sprintf("%d", record.NaptrPreference))
	}
	if record.NaptrFlags != "" {
		formData.Add("naptrFlags", record.NaptrFlags)
	}
	if record.NaptrServices != "" {
		formData.Add("naptrServices", record.NaptrServices)
	}
	if record.NaptrRegexp != "" {
		formData.Add("naptrRegexp", record.NaptrRegexp)
	}
	if record.NaptrReplacement != "" {
		formData.Add("naptrReplacement", record.NaptrReplacement)
	}
	if record.DName != "" {
		formData.Add("dName", record.DName)
	}
	if record.KeyTag > 0 {
		formData.Add("keyTag", fmt.Sprintf("%d", record.KeyTag))
	}
	if record.Algorithm != "" {
		formData.Add("algorithm", record.Algorithm)
	}
	if record.DigestType != "" {
		formData.Add("digestType", record.DigestType)
	}
	if record.Digest != "" {
		formData.Add("digest", record.Digest)
	}
	if record.SshfpAlgorithm != "" {
		formData.Add("sshfpAlgorithm", record.SshfpAlgorithm)
	}
	if record.SshfpFingerprintType != "" {
		formData.Add("sshfpFingerprintType", record.SshfpFingerprintType)
	}
	if record.SshfpFingerprint != "" {
		formData.Add("sshfpFingerprint", record.SshfpFingerprint)
	}
	if record.TlsaCertificateUsage != "" {
		formData.Add("tlsaCertificateUsage", record.TlsaCertificateUsage)
	}
	if record.TlsaSelector != "" {
		formData.Add("tlsaSelector", record.TlsaSelector)
	}
	if record.TlsaMatchingType != "" {
		formData.Add("tlsaMatchingType", record.TlsaMatchingType)
	}
	if record.TlsaCertificateAssociationData != "" {
		formData.Add("tlsaCertificateAssociationData", record.TlsaCertificateAssociationData)
	}
	if record.SvcPriority > 0 {
		formData.Add("svcPriority", fmt.Sprintf("%d", record.SvcPriority))
	}
	if record.SvcTargetName != "" {
		formData.Add("svcTargetName", record.SvcTargetName)
	}
	if record.SvcParams != "" {
		formData.Add("svcParams", record.SvcParams)
	}
	if record.AutoIpv4Hint {
		formData.Add("autoIpv4Hint", "true")
	}
	if record.AutoIpv6Hint {
		formData.Add("autoIpv6Hint", "true")
	}
	if record.UriPriority > 0 {
		formData.Add("uriPriority", fmt.Sprintf("%d", record.UriPriority))
	}
	if record.UriWeight > 0 {
		formData.Add("uriWeight", fmt.Sprintf("%d", record.UriWeight))
	}
	if record.Uri != "" {
		formData.Add("uri", record.Uri)
	}
	if record.Flags != "" {
		formData.Add("flags", record.Flags)
	}
	if record.Tag != "" {
		formData.Add("tag", record.Tag)
	}
	if record.Value != "" {
		formData.Add("value", record.Value)
	}
	if record.AName != "" {
		formData.Add("aName", record.AName)
	}
	if record.Forwarder != "" {
		formData.Add("forwarder", record.Forwarder)
	}
	if record.ForwarderPriority > 0 {
		formData.Add("forwarderPriority", fmt.Sprintf("%d", record.ForwarderPriority))
	}
	if record.DnssecValidation {
		formData.Add("dnssecValidation", "true")
	}
	if record.ProxyType != "" {
		formData.Add("proxyType", record.ProxyType)
	}
	if record.ProxyAddress != "" {
		formData.Add("proxyAddress", record.ProxyAddress)
	}
	if record.ProxyPort > 0 {
		formData.Add("proxyPort", fmt.Sprintf("%d", record.ProxyPort))
	}
	if record.ProxyUsername != "" {
		formData.Add("proxyUsername", record.ProxyUsername)
	}
	if record.ProxyPassword != "" {
		formData.Add("proxyPassword", record.ProxyPassword)
	}
	if record.AppName != "" {
		formData.Add("appName", record.AppName)
	}
	if record.ClassPath != "" {
		formData.Add("classPath", record.ClassPath)
	}
	if record.RecordData != "" {
		formData.Add("recordData", record.RecordData)
	}

	formData.Add("overwrite", "false")

	if err := c.makeRecordsRequest(ctx, "/add", http.MethodPost, nil, formData, nil); err != nil {
		return err
	}

	return nil
}

// UpdateRecord updates DNS record for a given domain.
func (c Client) UpdateRecord(ctx context.Context, oldRecord model.DNSRecord, newRecord model.DNSRecord) error {
	formData := url.Values{
		"type":   {string(oldRecord.Type)},
		"domain": {string(oldRecord.Domain)},
		"ttl":    {fmt.Sprintf("%d", newRecord.TTL)},
	}

	// Api uses newXX to provide the new value of each field.
	// That rule doesn't hold for all fields though.
	if newRecord.Domain != oldRecord.Domain {
		formData.Add("newDomain", string(newRecord.Domain))
	}

	if oldRecord.IPAddress != "" {
		formData.Add("ipAddress", oldRecord.IPAddress)
	}
	if newRecord.IPAddress != "" {
		formData.Add("newIpAddress", newRecord.IPAddress)
	}

	// Reset it on update in case it was missed or updated manually the first time.
	formData.Add("comments", TERRAFORM_PROVIDER_COMMENT)

	if newRecord.ExpiryTTL > 0 {
		formData.Add("expiryTtl", fmt.Sprintf("%d", newRecord.ExpiryTTL))
	}

	if newRecord.Ptr {
		formData.Add("ptr", "true")
	}
	if newRecord.CreatePtrZone {
		formData.Add("createPtrZone", "true")
	}
	if newRecord.UpdateSvcbHints {
		formData.Add("updateSvcbHints", "true")
	}

	if oldRecord.NameServer != "" {
		formData.Add("nameServer", oldRecord.NameServer)
	}
	if newRecord.NameServer != "" {
		formData.Add("newNameServer", newRecord.NameServer)
	}
	if newRecord.Glue != "" {
		formData.Add("glue", newRecord.Glue)
	}

	if newRecord.CName != "" {
		formData.Add("cname", newRecord.CName)
	}

	if oldRecord.PtrName != "" {
		formData.Add("ptrName", oldRecord.PtrName)
	}
	if newRecord.PtrName != "" {
		formData.Add("newPtrName", newRecord.PtrName)
	}

	if oldRecord.Exchange != "" {
		formData.Add("exchange", oldRecord.Exchange)
	}
	if newRecord.Exchange != "" {
		formData.Add("newExchange", newRecord.Exchange)
	}

	if oldRecord.Preference > 0 {
		formData.Add("preference", fmt.Sprintf("%d", oldRecord.Preference))
	}
	if newRecord.Preference > 0 {
		formData.Add("newPreference", fmt.Sprintf("%d", newRecord.Preference))
	}

	if oldRecord.Text != "" {
		formData.Add("text", oldRecord.Text)
	}
	if newRecord.Text != "" {
		formData.Add("newText", newRecord.Text)
	}

	if oldRecord.SplitText {
		formData.Add("splitText", "true")
	}
	if newRecord.SplitText {
		formData.Add("newSplitText", "true")
	}

	if oldRecord.Mailbox != "" {
		formData.Add("mailbox", oldRecord.Mailbox)
	}
	if newRecord.Mailbox != "" {
		formData.Add("newMailbox", newRecord.Mailbox)
	}

	if oldRecord.TxtDomain != "" {
		formData.Add("txtDomain", oldRecord.TxtDomain)
	}
	if newRecord.TxtDomain != "" {
		formData.Add("newTxtDomain", newRecord.TxtDomain)
	}

	if oldRecord.Priority > 0 {
		formData.Add("priority", fmt.Sprintf("%d", oldRecord.Priority))
	}
	if newRecord.Priority > 0 {
		formData.Add("newPriority", fmt.Sprintf("%d", newRecord.Priority))
	}

	if oldRecord.Weight > 0 {
		formData.Add("weight", fmt.Sprintf("%d", oldRecord.Weight))
	}
	if newRecord.Weight > 0 {
		formData.Add("newWeight", fmt.Sprintf("%d", newRecord.Weight))
	}

	if oldRecord.Port > 0 {
		formData.Add("port", fmt.Sprintf("%d", oldRecord.Port))
	}
	if newRecord.Port > 0 {
		formData.Add("newPort", fmt.Sprintf("%d", newRecord.Port))
	}

	if oldRecord.Target != "" {
		formData.Add("target", string(oldRecord.Target))
	}
	if newRecord.Target != "" {
		formData.Add("newTarget", string(newRecord.Target))
	}

	if oldRecord.NaptrOrder > 0 {
		formData.Add("naptrOrder", fmt.Sprintf("%d", oldRecord.NaptrOrder))
	}
	if oldRecord.NaptrPreference > 0 {
		formData.Add("naptrPreference", fmt.Sprintf("%d", oldRecord.NaptrPreference))
	}
	if oldRecord.NaptrFlags != "" {
		formData.Add("naptrFlags", oldRecord.NaptrFlags)
	}
	if oldRecord.NaptrServices != "" {
		formData.Add("naptrServices", oldRecord.NaptrServices)
	}
	if oldRecord.NaptrRegexp != "" {
		formData.Add("naptrRegexp", oldRecord.NaptrRegexp)
	}
	if oldRecord.NaptrReplacement != "" {
		formData.Add("naptrReplacement", oldRecord.NaptrReplacement)
	}
	if newRecord.NaptrOrder > 0 {
		formData.Add("newNaptrOrder", fmt.Sprintf("%d", newRecord.NaptrOrder))
	}
	if newRecord.NaptrPreference > 0 {
		formData.Add("newNaptrPreference", fmt.Sprintf("%d", newRecord.NaptrPreference))
	}
	if newRecord.NaptrFlags != "" {
		formData.Add("newNaptrFlags", newRecord.NaptrFlags)
	}
	if newRecord.NaptrServices != "" {
		formData.Add("newNaptrServices", newRecord.NaptrServices)
	}
	if newRecord.NaptrRegexp != "" {
		formData.Add("newNaptrRegexp", newRecord.NaptrRegexp)
	}
	if newRecord.NaptrReplacement != "" {
		formData.Add("newNaptrReplacement", newRecord.NaptrReplacement)
	}

	if oldRecord.DName != "" {
		formData.Add("dName", oldRecord.DName)
	}

	if oldRecord.KeyTag > 0 {
		formData.Add("keyTag", fmt.Sprintf("%d", oldRecord.KeyTag))
	}
	if newRecord.KeyTag > 0 {
		formData.Add("newKeyTag", fmt.Sprintf("%d", newRecord.KeyTag))
	}

	if oldRecord.Algorithm != "" {
		formData.Add("algorithm", oldRecord.Algorithm)
	}
	if newRecord.Algorithm != "" {
		formData.Add("newAlgorithm", newRecord.Algorithm)
	}
	if oldRecord.DigestType != "" {
		formData.Add("digestType", oldRecord.DigestType)
	}
	if newRecord.DigestType != "" {
		formData.Add("newDigestType", newRecord.DigestType)
	}
	if oldRecord.Digest != "" {
		formData.Add("digest", oldRecord.Digest)
	}
	if newRecord.Digest != "" {
		formData.Add("newDigest", newRecord.Digest)
	}

	if oldRecord.SshfpAlgorithm != "" {
		formData.Add("sshfpAlgorithm", oldRecord.SshfpAlgorithm)
	}
	if newRecord.SshfpAlgorithm != "" {
		formData.Add("newSshfpAlgorithm", newRecord.SshfpAlgorithm)
	}
	if oldRecord.SshfpFingerprintType != "" {
		formData.Add("sshfpFingerprintType", oldRecord.SshfpFingerprintType)
	}
	if newRecord.SshfpFingerprintType != "" {
		formData.Add("newSshfpFingerprintType", newRecord.SshfpFingerprintType)
	}
	if oldRecord.SshfpFingerprint != "" {
		formData.Add("sshfpFingerprint", oldRecord.SshfpFingerprint)
	}
	if newRecord.SshfpFingerprint != "" {
		formData.Add("newSshfpFingerprint", newRecord.SshfpFingerprint)
	}

	if oldRecord.TlsaCertificateUsage != "" {
		formData.Add("tlsaCertificateUsage", oldRecord.TlsaCertificateUsage)
	}
	if newRecord.TlsaCertificateUsage != "" {
		formData.Add("newTlsaCertificateUsage", newRecord.TlsaCertificateUsage)
	}
	if oldRecord.TlsaSelector != "" {
		formData.Add("tlsaSelector", oldRecord.TlsaSelector)
	}
	if newRecord.TlsaSelector != "" {
		formData.Add("newTlsaSelector", newRecord.TlsaSelector)
	}
	if oldRecord.TlsaMatchingType != "" {
		formData.Add("tlsaMatchingType", oldRecord.TlsaMatchingType)
	}
	if newRecord.TlsaMatchingType != "" {
		formData.Add("newTlsaMatchingType", newRecord.TlsaMatchingType)
	}
	if oldRecord.TlsaCertificateAssociationData != "" {
		formData.Add("tlsaCertificateAssociationData", oldRecord.TlsaCertificateAssociationData)
	}
	if newRecord.TlsaCertificateAssociationData != "" {
		formData.Add("newTlsaCertificateAssociationData", newRecord.TlsaCertificateAssociationData)
	}

	if oldRecord.SvcPriority > 0 {
		formData.Add("svcPriority", fmt.Sprintf("%d", oldRecord.SvcPriority))
	}
	if newRecord.SvcPriority > 0 {
		formData.Add("newSvcPriority", fmt.Sprintf("%d", newRecord.SvcPriority))
	}
	if oldRecord.SvcTargetName != "" {
		formData.Add("svcTargetName", oldRecord.SvcTargetName)
	}
	if newRecord.SvcTargetName != "" {
		formData.Add("newSvcTargetName", newRecord.SvcTargetName)
	}
	if oldRecord.SvcParams != "" {
		formData.Add("svcParams", oldRecord.SvcParams)
	}
	if newRecord.SvcParams != "" {
		formData.Add("newSvcParams", newRecord.SvcParams)
	}

	if newRecord.AutoIpv4Hint {
		formData.Add("autoIpv4Hint", "true")
	}
	if newRecord.AutoIpv6Hint {
		formData.Add("autoIpv6Hint", "true")
	}

	if oldRecord.UriPriority > 0 {
		formData.Add("uriPriority", fmt.Sprintf("%d", oldRecord.UriPriority))
	}
	if newRecord.UriPriority > 0 {
		formData.Add("newUriPriority", fmt.Sprintf("%d", newRecord.UriPriority))
	}
	if oldRecord.UriWeight > 0 {
		formData.Add("uriWeight", fmt.Sprintf("%d", oldRecord.UriWeight))
	}
	if newRecord.UriWeight > 0 {
		formData.Add("newUriWeight", fmt.Sprintf("%d", newRecord.UriWeight))
	}
	if oldRecord.Uri != "" {
		formData.Add("uri", oldRecord.Uri)
	}
	if newRecord.Uri != "" {
		formData.Add("newUri", newRecord.Uri)
	}
	if oldRecord.Flags != "" {
		formData.Add("flags", oldRecord.Flags)
	}
	if newRecord.Flags != "" {
		formData.Add("newFlags", newRecord.Flags)
	}
	if oldRecord.Tag != "" {
		formData.Add("tag", oldRecord.Tag)
	}
	if newRecord.Tag != "" {
		formData.Add("newTag", newRecord.Tag)
	}
	if oldRecord.Value != "" {
		formData.Add("value", oldRecord.Value)
	}
	if newRecord.Value != "" {
		formData.Add("newValue", newRecord.Value)
	}
	if oldRecord.AName != "" {
		formData.Add("aname", oldRecord.AName)
	}
	if newRecord.AName != "" {
		formData.Add("newAName", newRecord.AName)
	}
	if oldRecord.Forwarder != "" {
		formData.Add("forwarder", oldRecord.Forwarder)
	}
	if newRecord.Forwarder != "" {
		formData.Add("newForwarder", newRecord.Forwarder)
	}
	if oldRecord.ForwarderPriority > 0 {
		formData.Add("forwarderPriority", fmt.Sprintf("%d", oldRecord.ForwarderPriority))
	}
	if newRecord.ForwarderPriority > 0 {
		formData.Add("newForwarderPriority", fmt.Sprintf("%d", newRecord.ForwarderPriority))
	}
	if newRecord.DnssecValidation {
		formData.Add("dnssecValidation", "true")
	}
	if newRecord.ProxyType != "" {
		formData.Add("proxyType", newRecord.ProxyType)
	}
	if newRecord.ProxyAddress != "" {
		formData.Add("proxyAddress", newRecord.ProxyAddress)
	}
	if newRecord.ProxyPort > 0 {
		formData.Add("proxyPort", fmt.Sprintf("%d", newRecord.ProxyPort))
	}
	if newRecord.ProxyUsername != "" {
		formData.Add("proxyUsername", newRecord.ProxyUsername)
	}
	if newRecord.ProxyPassword != "" {
		formData.Add("proxyPassword", newRecord.ProxyPassword)
	}
	if oldRecord.AppName != "" {
		formData.Add("appName", oldRecord.AppName)
	}
	if oldRecord.ClassPath != "" {
		formData.Add("classPath", oldRecord.ClassPath)
	}
	if newRecord.RecordData != "" {
		formData.Add("recordData", newRecord.RecordData)
	}

	// Keep this to force update the record.
	formData.Add("overwrite", "true")

	if err := c.makeRecordsRequest(ctx, "/update", http.MethodPost, nil, formData, nil); err != nil {
		return err
	}

	return nil
}

// DeleteRecord deletes a DNS record.
func (c Client) DeleteRecord(ctx context.Context, record model.DNSRecord) error {
	params := url.Values{}

	if record.Domain != "" {
		params.Add("domain", string(record.Domain))
	}
	params.Add("type", string(record.Type))

	if record.IPAddress != "" {
		params.Add("ipAddress", record.IPAddress)
	}
	if record.Ptr {
		params.Add("ptr", "true")
	}
	if record.NameServer != "" {
		params.Add("nameServer", record.NameServer)
	}
	if record.Glue != "" {
		params.Add("glue", record.Glue)
	}
	if record.CName != "" {
		params.Add("cname", record.CName)
	}
	if record.PtrName != "" {
		params.Add("ptrName", record.PtrName)
	}
	if record.Exchange != "" {
		params.Add("exchange", record.Exchange)
	}
	if record.Preference > 0 {
		params.Add("preference", fmt.Sprintf("%d", record.Preference))
	}
	if record.Text != "" {
		params.Add("text", record.Text)
	}
	if record.SplitText {
		params.Add("splitText", "true")
	}
	if record.Mailbox != "" {
		params.Add("mailbox", record.Mailbox)
	}
	if record.TxtDomain != "" {
		params.Add("txtDomain", record.TxtDomain)
	}
	if record.Priority > 0 {
		params.Add("priority", fmt.Sprintf("%d", record.Priority))
	}
	if record.Weight > 0 {
		params.Add("weight", fmt.Sprintf("%d", record.Weight))
	}
	if record.Port > 0 {
		params.Add("port", fmt.Sprintf("%d", record.Port))
	}
	if record.Target != "" {
		params.Add("target", string(record.Target))
	}
	if record.NaptrOrder > 0 {
		params.Add("naptrOrder", fmt.Sprintf("%d", record.NaptrOrder))
	}
	if record.NaptrPreference > 0 {
		params.Add("naptrPreference", fmt.Sprintf("%d", record.NaptrPreference))
	}
	if record.NaptrFlags != "" {
		params.Add("naptrFlags", record.NaptrFlags)
	}
	if record.NaptrServices != "" {
		params.Add("naptrServices", record.NaptrServices)
	}
	if record.NaptrRegexp != "" {
		params.Add("naptrRegexp", record.NaptrRegexp)
	}
	if record.NaptrReplacement != "" {
		params.Add("naptrReplacement", record.NaptrReplacement)
	}
	if record.DName != "" {
		params.Add("dName", record.DName)
	}
	if record.KeyTag > 0 {
		params.Add("keyTag", fmt.Sprintf("%d", record.KeyTag))
	}
	if record.Algorithm != "" {
		params.Add("algorithm", record.Algorithm)
	}
	if record.DigestType != "" {
		params.Add("digestType", record.DigestType)
	}
	if record.Digest != "" {
		params.Add("digest", record.Digest)
	}
	if record.SshfpAlgorithm != "" {
		params.Add("sshfpAlgorithm", record.SshfpAlgorithm)
	}
	if record.SshfpFingerprintType != "" {
		params.Add("sshfpFingerprintType", record.SshfpFingerprintType)
	}
	if record.SshfpFingerprint != "" {
		params.Add("sshfpFingerprint", record.SshfpFingerprint)
	}
	if record.TlsaCertificateUsage != "" {
		params.Add("tlsaCertificateUsage", record.TlsaCertificateUsage)
	}
	if record.TlsaSelector != "" {
		params.Add("tlsaSelector", record.TlsaSelector)
	}
	if record.TlsaMatchingType != "" {
		params.Add("tlsaMatchingType", record.TlsaMatchingType)
	}
	if record.TlsaCertificateAssociationData != "" {
		params.Add("tlsaCertificateAssociationData", record.TlsaCertificateAssociationData)
	}
	if record.SvcPriority > 0 {
		params.Add("svcPriority", fmt.Sprintf("%d", record.SvcPriority))
	}
	if record.SvcTargetName != "" {
		params.Add("svcTargetName", record.SvcTargetName)
	}
	if record.SvcParams != "" {
		params.Add("svcParams", record.SvcParams)
	}
	if record.UriPriority > 0 {
		params.Add("uriPriority", fmt.Sprintf("%d", record.UriPriority))
	}
	if record.UriWeight > 0 {
		params.Add("uriWeight", fmt.Sprintf("%d", record.UriWeight))
	}
	if record.Uri != "" {
		params.Add("uri", record.Uri)
	}
	if record.Flags != "" {
		params.Add("flags", record.Flags)
	}
	if record.Tag != "" {
		params.Add("tag", record.Tag)
	}
	if record.Value != "" {
		params.Add("value", record.Value)
	}
	if record.AName != "" {
		params.Add("aName", record.AName)
	}
	if record.Forwarder != "" {
		params.Add("forwarder", record.Forwarder)
	}
	if record.ForwarderPriority > 0 {
		params.Add("forwarderPriority", fmt.Sprintf("%d", record.ForwarderPriority))
	}
	if record.ProxyType != "" {
		params.Add("proxyType", record.ProxyType)
	}
	if record.ProxyAddress != "" {
		params.Add("proxyAddress", record.ProxyAddress)
	}
	if record.ProxyPort > 0 {
		params.Add("proxyPort", fmt.Sprintf("%d", record.ProxyPort))
	}
	if record.ProxyUsername != "" {
		params.Add("proxyUsername", record.ProxyUsername)
	}
	if record.ProxyPassword != "" {
		params.Add("proxyPassword", record.ProxyPassword)
	}
	if record.AppName != "" {
		params.Add("appName", record.AppName)
	}
	if record.ClassPath != "" {
		params.Add("classPath", record.ClassPath)
	}
	if record.RecordData != "" {
		params.Add("recordData", record.RecordData)
	}

	return c.makeRecordsRequest(ctx, "/delete", http.MethodGet, params, nil, nil)
}

// ListZones retrieves all DNS zones from the server.
func (c Client) ListZones(ctx context.Context) ([]model.DNSZone, error) {
	var apiResponse struct {
		Response struct {
			Zones []model.DNSZone `json:"zones"`
		} `json:"response"`
		Status string `json:"status"`
	}

	err := c.makeZonesRequest(ctx, "/list", http.MethodGet, nil, nil, &apiResponse)
	if err != nil {
		return nil, err
	}

	return apiResponse.Response.Zones, nil
}

// CreateZone creates a new DNS zone.
func (c Client) CreateZone(ctx context.Context, zone model.DNSZone) error {
	formData := url.Values{
		"zone": {zone.Name},
		"type": {string(zone.Type)},
	}

	// Add optional parameters based on zone type
	if zone.Type == model.ZONE_SECONDARY || zone.Type == model.ZONE_STUB {
		// Add primary name server addresses if needed
		_ = zone // prevent unused variable warning
	}

	return c.makeZonesRequest(ctx, "/create", http.MethodPost, nil, formData, nil)
}

// DeleteZone deletes a DNS zone.
func (c Client) DeleteZone(ctx context.Context, zoneName string) error {
	formData := url.Values{
		"zone": {zoneName},
	}

	return c.makeZonesRequest(ctx, "/delete", http.MethodPost, nil, formData, nil)
}

func constructFullDomain(name, zone string) string {
	if name == "@" || name == "" {
		return zone
	}
	if strings.HasSuffix(name, "."+zone) {
		return name
	}
	if name == zone {
		return name
	}
	return name + "." + zone
}

func mapAPIDNSRecordToDNSRecord(apiRecord apiDNSRecordResponseItem, zone string) model.DNSRecord {
	return model.DNSRecord{
		Type:   model.DNSRecordType(apiRecord.Type),
		Domain: model.DNSRecordName(constructFullDomain(apiRecord.Domain, zone)),

		TTL: model.DNSRecordTTL(apiRecord.TTL),

		Comments:  apiRecord.Comments,
		ExpiryTTL: model.DNSRecordTTL(apiRecord.RData.ExpiryTTL),

		IPAddress:       apiRecord.RData.IPAddress,
		Ptr:             apiRecord.RData.Ptr,
		CreatePtrZone:   apiRecord.RData.CreatePtrZone,
		UpdateSvcbHints: apiRecord.RData.UpdateSvcbHints,

		NameServer: apiRecord.RData.NameServer,
		Glue:       apiRecord.RData.Glue,

		CName: apiRecord.RData.CName,

		PtrName: apiRecord.RData.PtrName,

		Exchange:   apiRecord.RData.Exchange,
		Preference: model.DNSRecordPrio(apiRecord.RData.Preference),

		Text:      apiRecord.RData.Text,
		SplitText: apiRecord.RData.SplitText,

		Mailbox:   apiRecord.RData.Mailbox,
		TxtDomain: apiRecord.RData.TxtDomain,

		Priority: model.DNSRecordPrio(apiRecord.RData.Priority),
		Weight:   model.DNSRecordSRVWeight(apiRecord.RData.Weight),
		Port:     model.DNSRecordSRVPort(apiRecord.RData.Port),
		Target:   model.DNSRecordSRVService(apiRecord.RData.Target),

		NaptrOrder:       apiRecord.RData.NaptrOrder,
		NaptrPreference:  apiRecord.RData.NaptrPreference,
		NaptrFlags:       apiRecord.RData.NaptrFlags,
		NaptrServices:    apiRecord.RData.NaptrServices,
		NaptrRegexp:      apiRecord.RData.NaptrRegexp,
		NaptrReplacement: apiRecord.RData.NaptrReplacement,

		DName: apiRecord.RData.DName,

		KeyTag:     apiRecord.RData.KeyTag,
		Algorithm:  apiRecord.RData.Algorithm,
		DigestType: apiRecord.RData.DigestType,
		Digest:     apiRecord.RData.Digest,

		SshfpAlgorithm:       apiRecord.RData.SshfpAlgorithm,
		SshfpFingerprintType: apiRecord.RData.SshfpFingerprintType,
		SshfpFingerprint:     apiRecord.RData.SshfpFingerprint,

		TlsaCertificateUsage:           apiRecord.RData.TlsaCertificateUsage,
		TlsaSelector:                   apiRecord.RData.TlsaSelector,
		TlsaMatchingType:               apiRecord.RData.TlsaMatchingType,
		TlsaCertificateAssociationData: apiRecord.RData.TlsaCertificateAssociationData,

		SvcPriority:   apiRecord.RData.SvcPriority,
		SvcTargetName: apiRecord.RData.SvcTargetName,
		SvcParams:     apiRecord.RData.SvcParams,

		AutoIpv4Hint: apiRecord.RData.AutoIpv4Hint,
		AutoIpv6Hint: apiRecord.RData.AutoIpv6Hint,

		UriPriority: apiRecord.RData.UriPriority,
		UriWeight:   apiRecord.RData.UriWeight,
		Uri:         apiRecord.RData.Uri,

		Flags: apiRecord.RData.Flags,
		Tag:   apiRecord.RData.Tag,
		Value: apiRecord.RData.Value,

		AName: apiRecord.RData.AName,

		Forwarder:         apiRecord.RData.Forwarder,
		ForwarderPriority: apiRecord.RData.ForwarderPriority,
		DnssecValidation:  apiRecord.RData.DnssecValidation,
		ProxyType:         apiRecord.RData.ProxyType,
		ProxyAddress:      apiRecord.RData.ProxyAddress,
		ProxyPort:         apiRecord.RData.ProxyPort,
		ProxyUsername:     apiRecord.RData.ProxyUsername,
		ProxyPassword:     apiRecord.RData.ProxyPassword,

		AppName:    apiRecord.RData.AppName,
		ClassPath:  apiRecord.RData.ClassPath,
		RecordData: apiRecord.RData.RecordData,
	}
}
