package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/kevynb/terraform-provider-technitium/internal/client"
	"github.com/kevynb/terraform-provider-technitium/internal/model"
)

func testAccPreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC is not set")
	}
	if os.Getenv("TECHNITIUM_API_URL") == "" {
		t.Fatal("TECHNITIUM_API_URL must be set for acceptance tests")
	}
	if os.Getenv("TECHNITIUM_API_TOKEN") == "" {
		t.Fatal("TECHNITIUM_API_TOKEN must be set for acceptance tests")
	}
	if err := waitForZoneList(); err != nil {
		t.Fatalf("Technitium server not ready: %v", err)
	}
}

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"technitium": providerserver.NewProtocol6WithError(New(
		"test",
		func(apiURL, token string, skipCertificateVerification bool) (model.DNSApiClient, error) {
			return client.NewClient(apiURL, token, skipCertificateVerification)
		},
	)()),
}

func testAccAPIURL() string {
	apiURL := os.Getenv("TECHNITIUM_API_URL")
	if strings.HasSuffix(apiURL, "/api") {
		return strings.TrimSuffix(apiURL, "/api")
	}
	return apiURL
}

func testAccClient(t *testing.T) *client.Client {
	t.Helper()

	apiURL := os.Getenv("TECHNITIUM_API_URL")
	apiToken := os.Getenv("TECHNITIUM_API_TOKEN")
	skipVerify := parseEnvBool(os.Getenv("TECHNITIUM_SKIP_TLS_VERIFY"))

	c, err := client.NewClient(apiURL, apiToken, skipVerify)
	if err != nil {
		t.Fatalf("failed to create API client: %v", err)
	}
	return c
}

func parseEnvBool(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func waitForZoneList() error {
	const retryTimeout = 120 * time.Second
	const retryInterval = 1 * time.Second

	deadline := time.Now().Add(retryTimeout)
	var lastErr error
	for time.Now().Before(deadline) {
		apiClient, err := client.NewClient(
			os.Getenv("TECHNITIUM_API_URL"),
			os.Getenv("TECHNITIUM_API_TOKEN"),
			parseEnvBool(os.Getenv("TECHNITIUM_SKIP_TLS_VERIFY")),
		)
		if err != nil {
			lastErr = err
			time.Sleep(retryInterval)
			continue
		}
		_, err = apiClient.ListZones(context.Background())
		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(retryInterval)
	}

	return lastErr
}

func waitForZoneAbsent(apiClient *client.Client, zoneName string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		zones, err := apiClient.ListZones(context.Background())
		if err != nil {
			lastErr = err
			time.Sleep(1 * time.Second)
			continue
		}
		found := false
		for _, z := range zones {
			if z.Name == zoneName {
				found = true
				break
			}
		}
		if !found {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("zone %s still present after timeout", zoneName)
}

func waitForRecordAbsent(apiClient *client.Client, target model.DNSRecord, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		records, err := apiClient.GetRecords(context.Background(), target.Domain)
		if err != nil {
			lastErr = err
			time.Sleep(1 * time.Second)
			continue
		}
		found := false
		for _, rec := range records {
			if rec.SameKey(target) {
				found = true
				break
			}
		}
		if !found {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("record %s still present after timeout", target.Domain)
}
