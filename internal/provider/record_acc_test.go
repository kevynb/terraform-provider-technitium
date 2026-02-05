package provider

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/kevynb/terraform-provider-technitium/internal/model"
)

func TestAccRecordResource_basic(t *testing.T) {
	zoneName := acctest.RandomWithPrefix("tfacc") + ".example.local"
	recordName := "a"
	recordDomain := recordName + "." + zoneName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create zone + record and verify attributes.
				Config: testAccRecordConfig(zoneName, recordDomain, "1.2.3.4"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_record.test", "domain", recordDomain),
					resource.TestCheckResourceAttr("technitium_record.test", "type", "A"),
					resource.TestCheckResourceAttr("technitium_record.test", "ip_address", "1.2.3.4"),
				),
			},
			{
				// Update record value and verify.
				Config: testAccRecordConfig(zoneName, recordDomain, "5.6.7.8"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_record.test", "ip_address", "5.6.7.8"),
				),
			},
			{
				// Import existing record into state and verify.
				ResourceName:                         "technitium_record.test",
				ImportState:                          true,
				ImportStateId:                        zoneName + ":" + recordName + ":A:5.6.7.8",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "domain",
			},
			{
				// Drift test: delete the record out-of-band, then expect a non-empty plan.
				PreConfig: func() {
					apiClient := testAccClient(t)
					target := testAccRecordModel(recordDomain, "5.6.7.8")
					if err := apiClient.DeleteRecord(context.Background(), target); err != nil {
						if !strings.Contains(err.Error(), "no such record") {
							t.Fatalf("drift setup failed: %v", err)
						}
					}
					if err := waitForRecordAbsent(apiClient, target, 60*time.Second); err != nil {
						t.Fatalf("drift setup wait failed: %v", err)
					}
				},
				Config:             testAccRecordConfig(zoneName, recordDomain, "5.6.7.8"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// Recreate the record so destroy succeeds cleanly.
				Config: testAccRecordConfig(zoneName, recordDomain, "5.6.7.8"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_record.test", "ip_address", "5.6.7.8"),
				),
			},
		},
	})
}

func TestAccRecordResource_txt(t *testing.T) {
	zoneName := acctest.RandomWithPrefix("tfacc") + ".example.local"
	recordName := "txt"
	recordDomain := recordName + "." + zoneName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create zone + TXT record and verify attributes.
				Config: testAccRecordConfigTXT(zoneName, recordDomain, "hello world"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_record.txt", "domain", recordDomain),
					resource.TestCheckResourceAttr("technitium_record.txt", "type", "TXT"),
					resource.TestCheckResourceAttr("technitium_record.txt", "text", "hello world"),
				),
			},
			{
				// Update TXT record value and verify.
				Config: testAccRecordConfigTXT(zoneName, recordDomain, "updated text"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_record.txt", "text", "updated text"),
				),
			},
			{
				// Import existing TXT record into state and verify.
				ResourceName:                         "technitium_record.txt",
				ImportState:                          true,
				ImportStateId:                        zoneName + ":" + recordName + ":TXT:updated text",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "domain",
			},
			{
				// Drift test: delete the record out-of-band, then expect a non-empty plan.
				PreConfig: func() {
					apiClient := testAccClient(t)
					target := model.DNSRecord{
						Type:   model.REC_TXT,
						Domain: model.DNSRecordName(recordDomain),
						TTL:    3600,
						Text:   "updated text",
					}
					if err := apiClient.DeleteRecord(context.Background(), target); err != nil {
						if !strings.Contains(err.Error(), "no such record") {
							t.Fatalf("drift setup failed: %v", err)
						}
					}
					if err := waitForRecordAbsent(apiClient, target, 60*time.Second); err != nil {
						t.Fatalf("drift setup wait failed: %v", err)
					}
				},
				Config:             testAccRecordConfigTXT(zoneName, recordDomain, "updated text"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// Recreate the record so destroy succeeds cleanly.
				Config: testAccRecordConfigTXT(zoneName, recordDomain, "updated text"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_record.txt", "text", "updated text"),
				),
			},
		},
	})
}

func TestAccRecordResource_cname(t *testing.T) {
	zoneName := acctest.RandomWithPrefix("tfacc") + ".example.local"
	recordName := "cname"
	recordDomain := recordName + "." + zoneName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create zone + CNAME record and verify attributes.
				Config: testAccRecordConfigCNAME(zoneName, recordDomain, "target.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_record.cname", "domain", recordDomain),
					resource.TestCheckResourceAttr("technitium_record.cname", "type", "CNAME"),
					resource.TestCheckResourceAttr("technitium_record.cname", "cname", "target.example.com"),
				),
			},
			{
				// Update CNAME target and verify.
				Config: testAccRecordConfigCNAME(zoneName, recordDomain, "target2.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_record.cname", "cname", "target2.example.com"),
				),
			},
			{
				// Import existing CNAME record into state and verify.
				ResourceName:                         "technitium_record.cname",
				ImportState:                          true,
				ImportStateId:                        zoneName + ":" + recordName + ":CNAME:target2.example.com",
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "domain",
			},
			{
				// Drift test: delete the record out-of-band, then expect a non-empty plan.
				PreConfig: func() {
					apiClient := testAccClient(t)
					target := model.DNSRecord{
						Type:   model.REC_CNAME,
						Domain: model.DNSRecordName(recordDomain),
						TTL:    3600,
						CName:  "target2.example.com",
					}
					if err := apiClient.DeleteRecord(context.Background(), target); err != nil {
						if !strings.Contains(err.Error(), "no such record") {
							t.Fatalf("drift setup failed: %v", err)
						}
					}
					if err := waitForRecordAbsent(apiClient, target, 60*time.Second); err != nil {
						t.Fatalf("drift setup wait failed: %v", err)
					}
				},
				Config:             testAccRecordConfigCNAME(zoneName, recordDomain, "target2.example.com"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// Recreate the record so destroy succeeds cleanly.
				Config: testAccRecordConfigCNAME(zoneName, recordDomain, "target2.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_record.cname", "cname", "target2.example.com"),
				),
			},
		},
	})
}

func testAccRecordConfig(zoneName, recordDomain, ip string) string {
	apiURL := testAccAPIURL()
	return `
provider "technitium" {
  url = "` + apiURL + `"
}

resource "technitium_zone" "test" {
  name = "` + zoneName + `"
  type = "Primary"
}

resource "technitium_record" "test" {
  domain     = "` + recordDomain + `"
  type       = "A"
  ttl        = 3600
  ip_address = "` + ip + `"
  depends_on = [technitium_zone.test]
}
`
}

func testAccRecordConfigTXT(zoneName, recordDomain, text string) string {
	apiURL := testAccAPIURL()
	return `
provider "technitium" {
  url = "` + apiURL + `"
}

resource "technitium_zone" "test_txt" {
  name = "` + zoneName + `"
  type = "Primary"
}

resource "technitium_record" "txt" {
  domain     = "` + recordDomain + `"
  type       = "TXT"
  ttl        = 3600
  text       = "` + text + `"
  depends_on = [technitium_zone.test_txt]
}
`
}

func testAccRecordConfigCNAME(zoneName, recordDomain, target string) string {
	apiURL := testAccAPIURL()
	return `
provider "technitium" {
  url = "` + apiURL + `"
}

resource "technitium_zone" "test_cname" {
  name = "` + zoneName + `"
  type = "Primary"
}

resource "technitium_record" "cname" {
  domain     = "` + recordDomain + `"
  type       = "CNAME"
  ttl        = 3600
  cname      = "` + target + `"
  depends_on = [technitium_zone.test_cname]
}
`
}

func testAccRecordModel(domain, ip string) model.DNSRecord {
	return model.DNSRecord{
		Type:      model.REC_A,
		Domain:    model.DNSRecordName(domain),
		TTL:       3600,
		IPAddress: ip,
	}
}
