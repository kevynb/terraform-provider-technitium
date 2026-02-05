package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccZoneResource_basic(t *testing.T) {
	zoneName := acctest.RandomWithPrefix("tfacc") + ".example.local"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create zone and verify basic attributes.
				Config: testAccZoneConfig(zoneName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_zone.test", "name", zoneName),
					resource.TestCheckResourceAttr("technitium_zone.test", "type", "Primary"),
				),
			},
			{
				// Import existing zone into state and verify.
				ResourceName:                         "technitium_zone.test",
				ImportState:                          true,
				ImportStateId:                        zoneName,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
			},
			{
				// Drift test: delete the zone out-of-band, then expect a non-empty plan.
				PreConfig: func() {
					apiClient := testAccClient(t)
					if err := apiClient.DeleteZone(context.Background(), zoneName); err != nil {
						t.Fatalf("drift setup failed: %v", err)
					}
					if err := waitForZoneAbsent(apiClient, zoneName, 60*time.Second); err != nil {
						t.Fatalf("drift setup wait failed: %v", err)
					}
				},
				Config:             testAccZoneConfig(zoneName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// Recreate the zone so destroy succeeds cleanly.
				Config: testAccZoneConfig(zoneName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_zone.test", "name", zoneName),
				),
			},
		},
	})
}

func TestAccZoneResource_forwarder(t *testing.T) {
	zoneName := acctest.RandomWithPrefix("tfacc") + ".example.local"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create forwarder zone with DNSSEC validation enabled.
				Config: testAccZoneForwarderConfig(zoneName, "8.8.8.8", "Udp", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_zone.forwarder", "name", zoneName),
					resource.TestCheckResourceAttr("technitium_zone.forwarder", "type", "Forwarder"),
					resource.TestCheckResourceAttr("technitium_zone.forwarder", "forwarder", "8.8.8.8"),
					resource.TestCheckResourceAttr("technitium_zone.forwarder", "protocol", "Udp"),
					resource.TestCheckResourceAttr("technitium_zone.forwarder", "dnssec_validation", "true"),
				),
			},
			{
				// Update forwarder and protocol and verify.
				Config: testAccZoneForwarderConfig(zoneName, "1.1.1.1", "Tcp", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_zone.forwarder", "forwarder", "1.1.1.1"),
					resource.TestCheckResourceAttr("technitium_zone.forwarder", "protocol", "Tcp"),
				),
			},
			{
				// Import existing forwarder zone into state and verify.
				ResourceName:                         "technitium_zone.forwarder",
				ImportState:                          true,
				ImportStateId:                        zoneName,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
			},
			{
				// Drift test: delete the zone out-of-band, then expect a non-empty plan.
				PreConfig: func() {
					apiClient := testAccClient(t)
					if err := apiClient.DeleteZone(context.Background(), zoneName); err != nil {
						t.Fatalf("drift setup failed: %v", err)
					}
					if err := waitForZoneAbsent(apiClient, zoneName, 60*time.Second); err != nil {
						t.Fatalf("drift setup wait failed: %v", err)
					}
				},
				Config:             testAccZoneForwarderConfig(zoneName, "1.1.1.1", "Tcp", true),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// Recreate the zone so destroy succeeds cleanly.
				Config: testAccZoneForwarderConfig(zoneName, "1.1.1.1", "Tcp", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("technitium_zone.forwarder", "name", zoneName),
				),
			},
		},
	})
}

func testAccZoneConfig(name string) string {
	apiURL := testAccAPIURL()
	return `
provider "technitium" {
  url = "` + apiURL + `"
}

resource "technitium_zone" "test" {
  name = "` + name + `"
  type = "Primary"
}
`
}

func testAccZoneForwarderConfig(name, forwarder, protocol string, dnssecValidation bool) string {
	apiURL := testAccAPIURL()
	dnssec := "false"
	if dnssecValidation {
		dnssec = "true"
	}
	return `
provider "technitium" {
  url = "` + apiURL + `"
}

resource "technitium_zone" "forwarder" {
  name              = "` + name + `"
  type              = "Forwarder"
  forwarder         = "` + forwarder + `"
  protocol          = "` + protocol + `"
  dnssec_validation = ` + dnssec + `
}
`
}
