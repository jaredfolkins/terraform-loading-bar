package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jaredfolkins/terraform-loading-bar/progress"
)

// GCPCredentials represents the structure of a Google Cloud Platform service account key
type GCPCredentials struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

// mockTerraformOutput is a sample Terraform JSON output for testing
var mockTerraformOutput = `
{"@level":"info","@message":"Terraform 1.8.0","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:47.831898Z","terraform":"1.8.0","type":"version","ui":"1.2"}
{"@level":"info","@message":"Plan: 2 to add, 0 to change, 0 to destroy.","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242882Z","changes":{"add":2,"change":0,"import":0,"remove":0,"operation":"plan"},"type":"change_summary"}
{"@level":"info","@message":"tls_private_key.ssh: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:55.664347Z","hook":{"resource":{"addr":"tls_private_key.ssh","module":"","resource":"tls_private_key.ssh","implied_provider":"tls","resource_type":"tls_private_key","resource_name":"ssh","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"tls_private_key.ssh: Creation complete after 0s [id=abcdef1234567890abcdef1234567890abcdef12]","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:56.417889Z","hook":{"resource":{"addr":"tls_private_key.ssh","module":"","resource":"tls_private_key.ssh","implied_provider":"tls","resource_type":"tls_private_key","resource_name":"ssh","resource_key":null},"action":"create","id_key":"id","id_value":"abcdef1234567890abcdef1234567890abcdef12","elapsed_seconds":0},"type":"apply_complete"}
{"@level":"info","@message":"Apply complete! Resources: 2 added, 0 changed, 0 destroyed.","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:53.060601Z","changes":{"add":2,"change":0,"import":0,"remove":0,"operation":"apply"},"type":"change_summary"}
`

// mockTerraformOutput28Steps is a sample Terraform JSON output with 28 steps for testing
var mockTerraformOutput28Steps = `
{"@level":"info","@message":"Terraform 1.8.0","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:47.831898Z","terraform":"1.8.0","type":"version","ui":"1.2"}
{"@level":"info","@message":"Plan: 28 to add, 0 to change, 0 to destroy.","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242882Z","changes":{"add":28,"change":0,"import":0,"remove":0,"operation":"plan"},"type":"change_summary"}
{"@level":"info","@message":"tls_private_key.ssh: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:55.664347Z","hook":{"resource":{"addr":"tls_private_key.ssh","module":"","resource":"tls_private_key.ssh","implied_provider":"tls","resource_type":"tls_private_key","resource_name":"ssh","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"tls_private_key.ssh: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:65.664347Z","hook":{"resource":{"addr":"tls_private_key.ssh","module":"","resource":"tls_private_key.ssh","implied_provider":"tls","resource_type":"tls_private_key","resource_name":"ssh","resource_key":null},"action":"create"},"type":"apply_progress"}
{"@level":"info","@message":"tls_private_key.ssh: Still creating... [20s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:75.664347Z","hook":{"resource":{"addr":"tls_private_key.ssh","module":"","resource":"tls_private_key.ssh","implied_provider":"tls","resource_type":"tls_private_key","resource_name":"ssh","resource_key":null},"action":"create"},"type":"apply_progress"}
{"@level":"info","@message":"tls_private_key.ssh: Creation complete after 30s [id=abcdef1234567890abcdef1234567890abcdef12]","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:85.417889Z","hook":{"resource":{"addr":"tls_private_key.ssh","module":"","resource":"tls_private_key.ssh","implied_provider":"tls","resource_type":"tls_private_key","resource_name":"ssh","resource_key":null},"action":"create","id_key":"id","id_value":"abcdef1234567890abcdef1234567890abcdef12","elapsed_seconds":30},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_network.main: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:86.664347Z","hook":{"resource":{"addr":"google_compute_network.main","module":"","resource":"google_compute_network.main","implied_provider":"google","resource_type":"google_compute_network","resource_name":"main","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_network.main: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:96.664347Z","hook":{"resource":{"addr":"google_compute_network.main","module":"","resource":"google_compute_network.main","implied_provider":"google","resource_type":"google_compute_network","resource_name":"main","resource_key":null},"action":"create"},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_network.main: Still creating... [20s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:06.664347Z","hook":{"resource":{"addr":"google_compute_network.main","module":"","resource":"google_compute_network.main","implied_provider":"google","resource_type":"google_compute_network","resource_name":"main","resource_key":null},"action":"create"},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_network.main: Creation complete after 30s [id=projects/test-project/global/networks/main]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:16.417889Z","hook":{"resource":{"addr":"google_compute_network.main","module":"","resource":"google_compute_network.main","implied_provider":"google","resource_type":"google_compute_network","resource_name":"main","resource_key":null},"action":"create","id_key":"id","id_value":"projects/test-project/global/networks/main","elapsed_seconds":30},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_subnetwork.main: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:17.664347Z","hook":{"resource":{"addr":"google_compute_subnetwork.main","module":"","resource":"google_compute_subnetwork.main","implied_provider":"google","resource_type":"google_compute_subnetwork","resource_name":"main","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_subnetwork.main: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:27.664347Z","hook":{"resource":{"addr":"google_compute_subnetwork.main","module":"","resource":"google_compute_subnetwork.main","implied_provider":"google","resource_type":"google_compute_subnetwork","resource_name":"main","resource_key":null},"action":"create"},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_subnetwork.main: Still creating... [20s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:37.664347Z","hook":{"resource":{"addr":"google_compute_subnetwork.main","module":"","resource":"google_compute_subnetwork.main","implied_provider":"google","resource_type":"google_compute_subnetwork","resource_name":"main","resource_key":null},"action":"create"},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_subnetwork.main: Creation complete after 30s [id=projects/test-project/regions/us-central1/subnetworks/main]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:47.417889Z","hook":{"resource":{"addr":"google_compute_subnetwork.main","module":"","resource":"google_compute_subnetwork.main","implied_provider":"google","resource_type":"google_compute_subnetwork","resource_name":"main","resource_key":null},"action":"create","id_key":"id","id_value":"projects/test-project/regions/us-central1/subnetworks/main","elapsed_seconds":30},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_firewall.allow_ssh: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:48.664347Z","hook":{"resource":{"addr":"google_compute_firewall.allow_ssh","module":"","resource":"google_compute_firewall.allow_ssh","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_ssh","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_firewall.allow_ssh: Creation complete after 1s [id=projects/test-project/global/firewalls/allow-ssh]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:49.417889Z","hook":{"resource":{"addr":"google_compute_firewall.allow_ssh","module":"","resource":"google_compute_firewall.allow_ssh","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_ssh","resource_key":null},"action":"create","id_key":"id","id_value":"projects/test-project/global/firewalls/allow-ssh","elapsed_seconds":1},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_instance.vm_instance: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:50.664347Z","hook":{"resource":{"addr":"google_compute_instance.vm_instance","module":"","resource":"google_compute_instance.vm_instance","implied_provider":"google","resource_type":"google_compute_instance","resource_name":"vm_instance","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_instance.vm_instance: Creation complete after 3s [id=projects/test-project/zones/us-central1-a/instances/vm-instance]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:53.417889Z","hook":{"resource":{"addr":"google_compute_instance.vm_instance","module":"","resource":"google_compute_instance.vm_instance","implied_provider":"google","resource_type":"google_compute_instance","resource_name":"vm_instance","resource_key":null},"action":"create","id_key":"id","id_value":"projects/test-project/zones/us-central1-a/instances/vm-instance","elapsed_seconds":3},"type":"apply_complete"}
{"@level":"info","@message":"google_dns_managed_zone.main: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:54.664347Z","hook":{"resource":{"addr":"google_dns_managed_zone.main","module":"","resource":"google_dns_managed_zone.main","implied_provider":"google","resource_type":"google_dns_managed_zone","resource_name":"main","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_dns_managed_zone.main: Creation complete after 2s [id=projects/test-project/managedZones/main]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:56.417889Z","hook":{"resource":{"addr":"google_dns_managed_zone.main","module":"","resource":"google_dns_managed_zone.main","implied_provider":"google","resource_type":"google_dns_managed_zone","resource_name":"main","resource_key":null},"action":"create","id_key":"id","id_value":"projects/test-project/managedZones/main","elapsed_seconds":2},"type":"apply_complete"}
{"@level":"info","@message":"google_dns_record_set.a: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:57.664347Z","hook":{"resource":{"addr":"google_dns_record_set.a","module":"","resource":"google_dns_record_set.a","implied_provider":"google","resource_type":"google_dns_record_set","resource_name":"a","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_dns_record_set.a: Creation complete after 1s [id=projects/test-project/managedZones/main/rrsets/test.example.com./A]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:58.417889Z","hook":{"resource":{"addr":"google_dns_record_set.a","module":"","resource":"google_dns_record_set.a","implied_provider":"google","resource_type":"google_dns_record_set","resource_name":"a","resource_key":null},"action":"create","id_key":"id","id_value":"projects/test-project/managedZones/main/rrsets/test.example.com./A","elapsed_seconds":1},"type":"apply_complete"}
{"@level":"info","@message":"google_dns_record_set.www: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:59.664347Z","hook":{"resource":{"addr":"google_dns_record_set.www","module":"","resource":"google_dns_record_set.www","implied_provider":"google","resource_type":"google_dns_record_set","resource_name":"www","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_dns_record_set.www: Creation complete after 1s [id=projects/test-project/managedZones/main/rrsets/www.test.example.com./A]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:00.417889Z","hook":{"resource":{"addr":"google_dns_record_set.www","module":"","resource":"google_dns_record_set.www","implied_provider":"google","resource_type":"google_dns_record_set","resource_name":"www","resource_key":null},"action":"create","id_key":"id","id_value":"projects/test-project/managedZones/main/rrsets/www.test.example.com./A","elapsed_seconds":1},"type":"apply_complete"}
{"@level":"info","@message":"google_storage_bucket.main: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:01.664347Z","hook":{"resource":{"addr":"google_storage_bucket.main","module":"","resource":"google_storage_bucket.main","implied_provider":"google","resource_type":"google_storage_bucket","resource_name":"main","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_storage_bucket.main: Creation complete after 2s [id=test-project-bucket]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:03.417889Z","hook":{"resource":{"addr":"google_storage_bucket.main","module":"","resource":"google_storage_bucket.main","implied_provider":"google","resource_type":"google_storage_bucket","resource_name":"main","resource_key":null},"action":"create","id_key":"id","id_value":"test-project-bucket","elapsed_seconds":2},"type":"apply_complete"}
{"@level":"info","@message":"google_storage_bucket_object.index: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:04.664347Z","hook":{"resource":{"addr":"google_storage_bucket_object.index","module":"","resource":"google_storage_bucket_object.index","implied_provider":"google","resource_type":"google_storage_bucket_object","resource_name":"index","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_storage_bucket_object.index: Creation complete after 1s [id=test-project-bucket/index.html]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:05.417889Z","hook":{"resource":{"addr":"google_storage_bucket_object.index","module":"","resource":"google_storage_bucket_object.index","implied_provider":"google","resource_type":"google_storage_bucket_object","resource_name":"index","resource_key":null},"action":"create","id_key":"id","id_value":"test-project-bucket/index.html","elapsed_seconds":1},"type":"apply_complete"}
{"@level":"info","@message":"google_storage_bucket_iam_member.public: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:06.664347Z","hook":{"resource":{"addr":"google_storage_bucket_iam_member.public","module":"","resource":"google_storage_bucket_iam_member.public","implied_provider":"google","resource_type":"google_storage_bucket_iam_member","resource_name":"public","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_storage_bucket_iam_member.public: Creation complete after 1s [id=test-project-bucket/roles/storage.objectViewer/allUsers]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:07.417889Z","hook":{"resource":{"addr":"google_storage_bucket_iam_member.public","module":"","resource":"google_storage_bucket_iam_member.public","implied_provider":"google","resource_type":"google_storage_bucket_iam_member","resource_name":"public","resource_key":null},"action":"create","id_key":"id","id_value":"test-project-bucket/roles/storage.objectViewer/allUsers","elapsed_seconds":1},"type":"apply_complete"}
{"@level":"info","@message":"google_cloud_run_service.main: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:08.664347Z","hook":{"resource":{"addr":"google_cloud_run_service.main","module":"","resource":"google_cloud_run_service.main","implied_provider":"google","resource_type":"google_cloud_run_service","resource_name":"main","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_cloud_run_service.main: Creation complete after 3s [id=projects/test-project/locations/us-central1/services/main]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:11.417889Z","hook":{"resource":{"addr":"google_cloud_run_service.main","module":"","resource":"google_cloud_run_service.main","implied_provider":"google","resource_type":"google_cloud_run_service","resource_name":"main","resource_key":null},"action":"create","id_key":"id","id_value":"projects/test-project/locations/us-central1/services/main","elapsed_seconds":3},"type":"apply_complete"}
{"@level":"info","@message":"google_cloud_run_service_iam_member.public: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:12.664347Z","hook":{"resource":{"addr":"google_cloud_run_service_iam_member.public","module":"","resource":"google_cloud_run_service_iam_member.public","implied_provider":"google","resource_type":"google_cloud_run_service_iam_member","resource_name":"public","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_cloud_run_service_iam_member.public: Creation complete after 1s [id=projects/test-project/locations/us-central1/services/main/roles/run.invoker/allUsers]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:13.417889Z","hook":{"resource":{"addr":"google_cloud_run_service_iam_member.public","module":"","resource":"google_cloud_run_service_iam_member.public","implied_provider":"google","resource_type":"google_cloud_run_service_iam_member","resource_name":"public","resource_key":null},"action":"create","id_key":"id","id_value":"projects/test-project/locations/us-central1/services/main/roles/run.invoker/allUsers","elapsed_seconds":1},"type":"apply_complete"}
{"@level":"info","@message":"Apply complete! Resources: 28 added, 0 changed, 0 destroyed.","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:53.060601Z","changes":{"add":28,"change":0,"import":0,"remove":0,"operation":"apply"},"type":"change_summary"}
`

// mockGCPCredentials is a sample GCP credentials JSON for testing
var mockGCPCredentials = GCPCredentials{
	Type:                    "service_account",
	ProjectID:               "test-project",
	PrivateKeyID:            "test-key-id",
	PrivateKey:              "test-private-key",
	ClientEmail:             "test@test-project.iam.gserviceaccount.com",
	ClientID:                "test-client-id",
	AuthURI:                 "https://accounts.google.com/o/oauth2/auth",
	TokenURI:                "https://oauth2.googleapis.com/token",
	AuthProviderX509CertURL: "https://www.googleapis.com/oauth2/v1/certs",
	ClientX509CertURL:       "https://www.googleapis.com/robot/v1/metadata/x509/test%40test-project.iam.gserviceaccount.com",
}

// copyFile copies a file from src to dst. If dst exists, it will be overwritten.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Sync()
}

// setupTestEnvironment creates a temporary test environment
func setupTestEnvironment(t *testing.T) (string, func()) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "lemc-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create source directory
	sourceDir := filepath.Join(tempDir, "terraform-config")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	// Create LEMC directories
	lemcPrivateDir := filepath.Join(tempDir, "lemc", "private")
	lemcPublicDir := filepath.Join(tempDir, "lemc", "public")
	if err := os.MkdirAll(lemcPrivateDir, 0755); err != nil {
		t.Fatalf("Failed to create LEMC private dir: %v", err)
	}
	if err := os.MkdirAll(lemcPublicDir, 0755); err != nil {
		t.Fatalf("Failed to create LEMC public dir: %v", err)
	}

	// Create test Terraform files in source directory
	files := map[string]string{
		"main.tf":      `resource "tls_private_key" "ssh" {}`,
		"outputs.tf":   `output "private_key" { value = tls_private_key.ssh.private_key_pem }`,
		"variables.tf": `variable "gcp_project_id" { type = string }`,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(sourceDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", name, err)
		}
	}

	// Copy files to LEMC private directory
	for name := range files {
		srcPath := filepath.Join(sourceDir, name)
		dstPath := filepath.Join(lemcPrivateDir, name)
		if err := copyFile(srcPath, dstPath); err != nil {
			t.Fatalf("Failed to copy %s: %v", name, err)
		}
	}

	// Create terraform.tfvars in LEMC private directory
	tfVarsContent := fmt.Sprintf(`gcp_project_id = "test-project"
lemc_uuid = "test-uuid"
lemc_username = "test-user"
lemc_scope = "individual"
lemc_user_id = "test-user-id"
domain_name = "test.example.com"
dns_zone_name = "test-zone"
credentials_file = "%s"
`, filepath.Join(lemcPrivateDir, "gcp-credentials.json"))

	if err := os.WriteFile(filepath.Join(lemcPrivateDir, "terraform.tfvars"), []byte(tfVarsContent), 0644); err != nil {
		t.Fatalf("Failed to write terraform.tfvars: %v", err)
	}

	// Set environment variables
	os.Setenv("GCP_KEY_JSON_CONTENT", string(mustMarshalJSON(mockGCPCredentials)))
	os.Setenv("LEMC_UUID", "test-uuid")
	os.Setenv("LEMC_USERNAME", "test-user")
	os.Setenv("LEMC_SCOPE", "individual")
	os.Setenv("LEMC_USER_ID", "test-user-id")
	os.Setenv("ROOT_DOMAIN", "test.example.com")
	os.Setenv("ROOT_ZONE", "test-zone")

	// Return cleanup function
	return tempDir, func() {
		os.RemoveAll(tempDir)
		os.Unsetenv("GCP_KEY_JSON_CONTENT")
		os.Unsetenv("LEMC_UUID")
		os.Unsetenv("LEMC_USERNAME")
		os.Unsetenv("LEMC_SCOPE")
		os.Unsetenv("LEMC_USER_ID")
		os.Unsetenv("ROOT_DOMAIN")
		os.Unsetenv("ROOT_ZONE")
	}
}

// mustMarshalJSON is a helper to marshal JSON without error handling
func mustMarshalJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// mockTerraformCommand is a helper to create a mock terraform command
func mockTerraformCommand(t *testing.T, output string) *exec.Cmd {
	cmd := exec.Command("echo", output)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func TestLEMCProgressBar(t *testing.T) {
	// Setup test environment
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a buffer to capture output
	var output bytes.Buffer

	// Create a progress handler with the mock output
	progressHandler := progress.NewProgressHandler(strings.NewReader(mockTerraformOutput))

	// Read and process the output
	var lines []string
	for {
		line, err := progressHandler.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("Error reading progress: %v", err)
		}
		if line != "" {
			// Format as LEMC output
			formattedLine := fmt.Sprintf("lemc.html.trunc; %s\n", strings.TrimSpace(line))
			output.WriteString(formattedLine)
			lines = append(lines, line)
			// Print to stdout for validation
			fmt.Print(formattedLine)
		}
	}

	// Verify the output
	outputStr := output.String()

	// Check for progress bar structure
	if !strings.Contains(outputStr, "[=") && !strings.Contains(outputStr, "[-/-]") {
		t.Error("Output does not contain progress bar structure")
	}

	// Check for LEMC formatting
	if !strings.Contains(outputStr, "lemc.html.trunc;") {
		t.Error("Output does not contain LEMC formatting")
	}

	// Check for specific messages
	expectedMessages := []string{
		"Plan: 2 to add",
		"Creating...",
		"Creation complete",
		"Apply complete!",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(outputStr, msg) {
			t.Errorf("Output does not contain expected message: %s", msg)
		}
	}

	// Verify progress bar transitions
	foundPlanning := false
	foundApply := false
	for _, line := range lines {
		if strings.Contains(line, "PLANNING") {
			foundPlanning = true
		}
		if strings.Contains(line, "[=") && !strings.Contains(line, "PLANNING") {
			foundApply = true
		}
	}

	if !foundPlanning {
		t.Error("Did not find planning phase in output")
	}
	if !foundApply {
		t.Error("Did not find apply phase in output")
	}
}

func TestLEMCEnvironmentSetup(t *testing.T) {
	// Setup test environment
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Verify environment variables
	envVars := map[string]string{
		"GCP_KEY_JSON_CONTENT": string(mustMarshalJSON(mockGCPCredentials)),
		"LEMC_UUID":            "test-uuid",
		"LEMC_USERNAME":        "test-user",
		"LEMC_SCOPE":           "individual",
		"LEMC_USER_ID":         "test-user-id",
		"ROOT_DOMAIN":          "test.example.com",
		"ROOT_ZONE":            "test-zone",
	}

	for key, expected := range envVars {
		if actual := os.Getenv(key); actual != expected {
			t.Errorf("Environment variable %s = %q, want %q", key, actual, expected)
		}
	}

	// Verify directory structure
	expectedDirs := []string{
		filepath.Join(tempDir, "terraform-config"),
		filepath.Join(tempDir, "lemc", "private"),
		filepath.Join(tempDir, "lemc", "public"),
	}

	for _, dir := range expectedDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory %s does not exist", dir)
		}
	}

	// Verify Terraform files
	expectedFiles := []string{
		"main.tf",
		"outputs.tf",
		"variables.tf",
		"terraform.tfvars",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tempDir, "lemc", "private", file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("File %s does not exist", path)
		}
	}
}

func TestLEMCProgressBarWithTimeout(t *testing.T) {
	// Setup test environment
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a progress handler with the mock output
	progressHandler := progress.NewProgressHandler(strings.NewReader(mockTerraformOutput))

	// Read and process the output with timeout
	var lines []string
	var err error

	for {
		select {
		case <-ctx.Done():
			t.Fatal("Test timed out after 5 seconds")
		default:
			line, readErr := progressHandler.ReadLine()
			if readErr != nil {
				if readErr == io.EOF {
					goto done
				}
				err = readErr
				return
			}
			if line != "" {
				lines = append(lines, line)
			}
		}
	}
done:

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	// Verify we got some output
	if len(lines) == 0 {
		t.Error("Expected non-empty output")
		return
	}

	// Verify progress bar structure
	foundProgressBar := false
	for _, line := range lines {
		if strings.Contains(line, "[=") || strings.Contains(line, "[-/-]") {
			foundProgressBar = true
			break
		}
	}
	if !foundProgressBar {
		t.Error("Output does not contain progress bar structure")
	}
}

func TestLEMCProgressBar28Steps(t *testing.T) {
	// Setup test environment
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a buffer to capture output
	var output bytes.Buffer

	// Create a progress handler with the mock output
	progressHandler := progress.NewProgressHandler(strings.NewReader(mockTerraformOutput28Steps))

	// Read and process the output
	var lines []string
	for {
		line, err := progressHandler.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("Error reading progress: %v", err)
		}
		if line != "" {
			// Format as LEMC output
			formattedLine := fmt.Sprintf("lemc.html.trunc; %s\n", strings.TrimSpace(line))
			output.WriteString(formattedLine)
			lines = append(lines, line)
			// Print to stdout for validation
			fmt.Print(formattedLine)
		}
	}

	// Verify the output
	outputStr := output.String()

	// Check for progress bar structure
	if !strings.Contains(outputStr, "[=") && !strings.Contains(outputStr, "[-/-]") {
		t.Error("Output does not contain progress bar structure")
	}

	// Check for LEMC formatting
	if !strings.Contains(outputStr, "lemc.html.trunc;") {
		t.Error("Output does not contain LEMC formatting")
	}

	// Check for specific messages
	expectedMessages := []string{
		"Plan: 28 to add",
		"Creating...",
		"Still creating...",
		"Creation complete",
		"Apply complete!",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(outputStr, msg) {
			t.Errorf("Output does not contain expected message: %s", msg)
		}
	}

	// Verify progress bar transitions
	foundPlanning := false
	foundApply := false
	progressSteps := 0
	for _, line := range lines {
		if strings.Contains(line, "PLANNING") {
			foundPlanning = true
		}
		if strings.Contains(line, "[=") && !strings.Contains(line, "PLANNING") {
			foundApply = true
			progressSteps++
		}
	}

	if !foundPlanning {
		t.Error("Did not find planning phase in output")
	}
	if !foundApply {
		t.Error("Did not find apply phase in output")
	}
	if progressSteps < 28 { // We expect at least 28 progress steps (one for each resource creation)
		t.Errorf("Expected at least 28 progress steps, got %d", progressSteps)
	}
}
