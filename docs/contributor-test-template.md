# Contributor Test Template

Use this as a starting point when adding new provider features. It reflects the
current test structure and acceptance harness.

## Quick Map

- Unit tests: `internal/.../*_test.go`
- Acceptance tests: `internal/provider/*_acc_test.go`
- Acceptance helpers: `internal/provider/acc_utils_test.go`
- Run unit tests: `make test`
- Run acceptance tests: `make acc` or `make acc-test`

## Acceptance Environment

Acceptance tests require a Technitium server. The local harness is:

```
make acc-up
make acc-test
```

`make acc-up` will start Docker and fetch a token into `tools/acceptance/token.env`.

_You can use `make acc` to run both commands in one step._

Set or override:
- `TECHNITIUM_API_URL` (default: `http://localhost:5380`)
- `TECHNITIUM_API_TOKEN` (set by `make acc-up` if empty)
- `TECHNITIUM_ADMIN_USER` / `TECHNITIUM_ADMIN_PASSWORD` (defaults: `admin` / `changeme`)
- `TECHNITIUM_SKIP_TLS_VERIFY` (optional)

## Unit Test Template

Add a unit test for any new model/config mapping, import parsing, or helpers.

Example pattern:

```go
func TestTF2ModelMapping_NewField(t *testing.T) {
    tfData := tfDNSRecord{
        Type:   types.StringValue("TXT"),
        Domain: types.StringValue("example.com"),
        TTL:    types.Int64Value(3600),
        Text:   types.StringValue("value"),
    }

    got := tf2model(tfData)
    if got.Text != "value" {
        t.Fatalf("Text mismatch: got %q", got.Text)
    }
}
```

## Acceptance Test Template (Resource)

Use a full lifecycle test with create, update, import, drift, and recreate.

```go
func TestAccMyResource_basic(t *testing.T) {
    name := acctest.RandomWithPrefix("tfacc") + ".example.local"

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                // Create and verify.
                Config: testAccMyResourceConfig(name, "value1"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("technitium_my_resource.test", "name", name),
                ),
            },
            {
                // Update and verify.
                Config: testAccMyResourceConfig(name, "value2"),
            },
            {
                // Import and verify.
                ResourceName:      "technitium_my_resource.test",
                ImportState:       true,
                ImportStateId:     name,
                ImportStateVerify: true,
                ImportStateVerifyIdentifierAttribute: "name",
            },
            {
                // Drift: mutate out-of-band, then expect non-empty plan.
                PreConfig: func() {
                    apiClient := testAccClient(t)
                    // mutate via apiClient
                },
                Config:             testAccMyResourceConfig(name, "value2"),
                PlanOnly:           true,
                ExpectNonEmptyPlan: true,
            },
            {
                // Recreate so destroy succeeds cleanly.
                Config: testAccMyResourceConfig(name, "value2"),
            },
        },
    })
}
```

### Notes

- Use unique names (`acctest.RandomWithPrefix`) to avoid collisions.
- If you delete a resource for drift, ensure the final step recreates it so
  the acceptance test destroy phase succeeds.
- For import, set `ImportStateVerifyIdentifierAttribute` when the resource
  does not use `id` in state.

