package progress_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/jaredfolkins/terraform-loading-bar/progress" // Import the package we're testing
)

var mockTerraformOutput = `
{"@level":"info","@message":"Terraform 1.8.0","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:47.831898Z","terraform":"1.8.0","type":"version","ui":"1.2"}
{"@level":"info","@message":"data.google_dns_managed_zone.zone: Refreshing...","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:53.938734Z","hook":{"resource":{"addr":"data.google_dns_managed_zone.zone","module":"","resource":"data.google_dns_managed_zone.zone","implied_provider":"google","resource_type":"google_dns_managed_zone","resource_name":"zone","resource_key":null},"action":"read"},"type":"apply_start"}
{"@level":"info","@message":"data.google_dns_managed_zone.zone: Refresh complete after 0s [id=projects/example-project/managedZones/example.com]","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.220509Z","hook":{"resource":{"addr":"data.google_dns_managed_zone.zone","module":"","resource":"data.google_dns_managed_zone.zone","implied_provider":"google","resource_type":"google_dns_managed_zone","resource_name":"zone","resource_key":null},"action":"read","id_key":"id","id_value":"projects/example-project/managedZones/example.com","elapsed_seconds":0},"type":"apply_complete"}
{"@level":"info","@message":"tls_private_key.ssh: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242635Z","change":{"resource":{"addr":"tls_private_key.ssh","module":"","resource":"tls_private_key.ssh","implied_provider":"tls","resource_type":"tls_private_key","resource_name":"ssh","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_network.vpc: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242741Z","change":{"resource":{"addr":"google_compute_network.vpc","module":"","resource":"google_compute_network.vpc","implied_provider":"google","resource_type":"google_compute_network","resource_name":"vpc","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_managed_ssl_certificate.ssl_certificate: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242756Z","change":{"resource":{"addr":"google_compute_managed_ssl_certificate.ssl_certificate","module":"","resource":"google_compute_managed_ssl_certificate.ssl_certificate","implied_provider":"google","resource_type":"google_compute_managed_ssl_certificate","resource_name":"ssl_certificate","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_global_address.lb_ip: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242765Z","change":{"resource":{"addr":"google_compute_global_address.lb_ip","module":"","resource":"google_compute_global_address.lb_ip","implied_provider":"google","resource_type":"google_compute_global_address","resource_name":"lb_ip","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_address.public_ip: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242772Z","change":{"resource":{"addr":"google_compute_address.public_ip","module":"","resource":"google_compute_address.public_ip","implied_provider":"google","resource_type":"google_compute_address","resource_name":"public_ip","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_firewall.allow_ssh: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242778Z","change":{"resource":{"addr":"google_compute_firewall.allow_ssh","module":"","resource":"google_compute_firewall.allow_ssh","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_ssh","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_health_check.lb_health_check: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242792Z","change":{"resource":{"addr":"google_compute_health_check.lb_health_check","module":"","resource":"google_compute_health_check.lb_health_check","implied_provider":"google","resource_type":"google_compute_health_check","resource_name":"lb_health_check","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_subnetwork.subnet: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242785Z","change":{"resource":{"addr":"google_compute_subnetwork.subnet","module":"","resource":"google_compute_subnetwork.subnet","implied_provider":"google","resource_type":"google_compute_subnetwork","resource_name":"subnet","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_firewall.allow_lb_healthcheck: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242799Z","change":{"resource":{"addr":"google_compute_firewall.allow_lb_healthcheck","module":"","resource":"google_compute_firewall.allow_lb_healthcheck","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_lb_healthcheck","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_instance.vm: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242805Z","change":{"resource":{"addr":"google_compute_instance.vm","module":"","resource":"google_compute_instance.vm","implied_provider":"google","resource_type":"google_compute_instance","resource_name":"vm","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_instance_group.instance_group: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242823Z","change":{"resource":{"addr":"google_compute_instance_group.instance_group","module":"","resource":"google_compute_instance_group.instance_group","implied_provider":"google","resource_type":"google_compute_instance_group","resource_name":"instance_group","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_backend_service.backend_service: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242832Z","change":{"resource":{"addr":"google_compute_backend_service.backend_service","module":"","resource":"google_compute_backend_service.backend_service","implied_provider":"google","resource_type":"google_compute_backend_service","resource_name":"backend_service","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_url_map.url_map: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242839Z","change":{"resource":{"addr":"google_compute_url_map.url_map","module":"","resource":"google_compute_url_map.url_map","implied_provider":"google","resource_type":"google_compute_url_map","resource_name":"url_map","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_target_http_proxy.http_proxy: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242845Z","change":{"resource":{"addr":"google_compute_target_http_proxy.http_proxy","module":"","resource":"google_compute_target_http_proxy.http_proxy","implied_provider":"google","resource_type":"google_compute_target_http_proxy","resource_name":"http_proxy","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_target_https_proxy.https_proxy: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242852Z","change":{"resource":{"addr":"google_compute_target_https_proxy.https_proxy","module":"","resource":"google_compute_target_https_proxy.https_proxy","implied_provider":"google","resource_type":"google_compute_target_https_proxy","resource_name":"https_proxy","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.http_forwarding_rule: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242859Z","change":{"resource":{"addr":"google_compute_global_forwarding_rule.http_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.http_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"http_forwarding_rule","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.https_forwarding_rule: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242867Z","change":{"resource":{"addr":"google_compute_global_forwarding_rule.https_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.https_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"https_forwarding_rule","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_dns_record_set.dns_record: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242875Z","change":{"resource":{"addr":"google_dns_record_set.dns_record","module":"","resource":"google_dns_record_set.dns_record","implied_provider":"google","resource_type":"google_dns_record_set","resource_name":"dns_record","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"Plan: 18 to add, 0 to change, 0 to destroy.","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242882Z","changes":{"add":18,"change":0,"import":0,"remove":0,"operation":"plan"},"type":"change_summary"}
{"@level":"info","@message":"Outputs: 8","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242915Z","outputs":{"dns_name":{"sensitive":false,"action":"create"},"instance_name":{"sensitive":false,"action":"create"},"load_balancer_ip":{"sensitive":false,"action":"create"},"private_ssh_key":{"sensitive":true,"action":"create"},"public_ip":{"sensitive":false,"action":"create"},"public_ssh_key":{"sensitive":false,"action":"create"},"subnet_name":{"sensitive":false,"action":"create"},"vpc_name":{"sensitive":false,"action":"create"}},"type":"outputs"}
{"@level":"info","@message":"tls_private_key.ssh: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:55.664347Z","hook":{"resource":{"addr":"tls_private_key.ssh","module":"","resource":"tls_private_key.ssh","implied_provider":"tls","resource_type":"tls_private_key","resource_name":"ssh","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_network.vpc: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:56.203712Z","hook":{"resource":{"addr":"google_compute_network.vpc","module":"","resource":"google_compute_network.vpc","implied_provider":"google","resource_type":"google_compute_network","resource_name":"vpc","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_address.public_ip: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:56.211378Z","hook":{"resource":{"addr":"google_compute_address.public_ip","module":"","resource":"google_compute_address.public_ip","implied_provider":"google","resource_type":"google_compute_address","resource_name":"public_ip","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_global_address.lb_ip: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:56.214787Z","hook":{"resource":{"addr":"google_compute_global_address.lb_ip","module":"","resource":"google_compute_global_address.lb_ip","implied_provider":"google","resource_type":"google_compute_global_address","resource_name":"lb_ip","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_managed_ssl_certificate.ssl_certificate: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:56.214792Z","hook":{"resource":{"addr":"google_compute_managed_ssl_certificate.ssl_certificate","module":"","resource":"google_compute_managed_ssl_certificate.ssl_certificate","implied_provider":"google","resource_type":"google_compute_managed_ssl_certificate","resource_name":"ssl_certificate","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_health_check.lb_health_check: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:56.215885Z","hook":{"resource":{"addr":"google_compute_health_check.lb_health_check","module":"","resource":"google_compute_health_check.lb_health_check","implied_provider":"google","resource_type":"google_compute_health_check","resource_name":"lb_health_check","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"tls_private_key.ssh: Creation complete after 0s [id=a573ee2b03dcb7d139fdc2a4588b7c413c89e0c9]","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:56.417889Z","hook":{"resource":{"addr":"tls_private_key.ssh","module":"","resource":"tls_private_key.ssh","implied_provider":"tls","resource_type":"tls_private_key","resource_name":"ssh","resource_key":null},"action":"create","id_key":"id","id_value":"a573ee2b03dcb7d139fdc2a4588b7c413c89e0c9","elapsed_seconds":0},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_address.public_ip: Creation complete after 11s [id=projects/example-project/regions/us-central1/addresses/example-instance-public-ip]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:07.193810Z","hook":{"resource":{"addr":"google_compute_address.public_ip","module":"","resource":"google_compute_address.public_ip","implied_provider":"google","resource_type":"google_compute_address","resource_name":"public_ip","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/regions/us-central1/addresses/example-instance-public-ip","elapsed_seconds":11},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_global_address.lb_ip: Creation complete after 12s [id=projects/example-project/global/addresses/example-instance-lb-ip]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:07.599403Z","hook":{"resource":{"addr":"google_compute_global_address.lb_ip","module":"","resource":"google_compute_global_address.lb_ip","implied_provider":"google","resource_type":"google_compute_global_address","resource_name":"lb_ip","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/addresses/example-instance-lb-ip","elapsed_seconds":12},"type":"apply_complete"}
{"@level":"info","@message":"google_dns_record_set.dns_record: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:07.607793Z","hook":{"resource":{"addr":"google_dns_record_set.dns_record","module":"","resource":"google_dns_record_set.dns_record","implied_provider":"google","resource_type":"google_dns_record_set","resource_name":"dns_record","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_health_check.lb_health_check: Creation complete after 12s [id=projects/example-project/global/healthChecks/example-instance-hc]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:07.773153Z","hook":{"resource":{"addr":"google_compute_health_check.lb_health_check","module":"","resource":"google_compute_health_check.lb_health_check","implied_provider":"google","resource_type":"google_compute_health_check","resource_name":"lb_health_check","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/healthChecks/example-instance-hc","elapsed_seconds":12},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_managed_ssl_certificate.ssl_certificate: Creation complete after 12s [id=projects/example-project/global/sslCertificates/example-instance-ssl-cert]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:07.837658Z","hook":{"resource":{"addr":"google_compute_managed_ssl_certificate.ssl_certificate","module":"","resource":"google_compute_managed_ssl_certificate.ssl_certificate","implied_provider":"google","resource_type":"google_compute_managed_ssl_certificate","resource_name":"ssl_certificate","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/sslCertificates/example-instance-ssl-cert","elapsed_seconds":12},"type":"apply_complete"}
{"@level":"info","@message":"google_dns_record_set.dns_record: Creation complete after 2s [id=projects/example-project/managedZones/example.com/rrsets/example-instance.example.com./A]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:10.382040Z","hook":{"resource":{"addr":"google_dns_record_set.dns_record","module":"","resource":"google_dns_record_set.dns_record","implied_provider":"google","resource_type":"google_dns_record_set","resource_name":"dns_record","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/managedZones/example.com/rrsets/example-instance.example.com./A","elapsed_seconds":2},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_network.vpc: Creation complete after 22s [id=projects/example-project/global/networks/example-instance-vpc]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:17.842928Z","hook":{"resource":{"addr":"google_compute_network.vpc","module":"","resource":"google_compute_network.vpc","implied_provider":"google","resource_type":"google_compute_network","resource_name":"vpc","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/networks/example-instance-vpc","elapsed_seconds":22},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_firewall.allow_ssh: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:17.862011Z","hook":{"resource":{"addr":"google_compute_firewall.allow_ssh","module":"","resource":"google_compute_firewall.allow_ssh","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_ssh","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_subnetwork.subnet: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:17.865632Z","hook":{"resource":{"addr":"google_compute_subnetwork.subnet","module":"","resource":"google_compute_subnetwork.subnet","implied_provider":"google","resource_type":"google_compute_subnetwork","resource_name":"subnet","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_firewall.allow_lb_healthcheck: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:17.865930Z","hook":{"resource":{"addr":"google_compute_firewall.allow_lb_healthcheck","module":"","resource":"google_compute_firewall.allow_lb_healthcheck","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_lb_healthcheck","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_firewall.allow_ssh: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:27.865355Z","hook":{"resource":{"addr":"google_compute_firewall.allow_ssh","module":"","resource":"google_compute_firewall.allow_ssh","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_ssh","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_firewall.allow_lb_healthcheck: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:27.867214Z","hook":{"resource":{"addr":"google_compute_firewall.allow_lb_healthcheck","module":"","resource":"google_compute_firewall.allow_lb_healthcheck","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_lb_healthcheck","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_subnetwork.subnet: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:27.867297Z","hook":{"resource":{"addr":"google_compute_subnetwork.subnet","module":"","resource":"google_compute_subnetwork.subnet","implied_provider":"google","resource_type":"google_compute_subnetwork","resource_name":"subnet","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_subnetwork.subnet: Creation complete after 11s [id=projects/example-project/regions/us-central1/subnetworks/example-instance-subnet]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:28.801735Z","hook":{"resource":{"addr":"google_compute_subnetwork.subnet","module":"","resource":"google_compute_subnetwork.subnet","implied_provider":"google","resource_type":"google_compute_subnetwork","resource_name":"subnet","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/regions/us-central1/subnetworks/example-instance-subnet","elapsed_seconds":11},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_instance.vm: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:28.819074Z","hook":{"resource":{"addr":"google_compute_instance.vm","module":"","resource":"google_compute_instance.vm","implied_provider":"google","resource_type":"google_compute_instance","resource_name":"vm","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_instance.vm: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:38.819807Z","hook":{"resource":{"addr":"google_compute_instance.vm","module":"","resource":"google_compute_instance.vm","implied_provider":"google","resource_type":"google_compute_instance","resource_name":"vm","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_instance.vm: Creation complete after 13s [id=projects/example-project/zones/us-central1-a/instances/example-instance]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:53.247084Z","hook":{"resource":{"addr":"google_compute_instance.vm","module":"","resource":"google_compute_instance.vm","implied_provider":"google","resource_type":"google_compute_instance","resource_name":"vm","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/zones/us-central1-a/instances/example-instance","elapsed_seconds":13},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_instance_group.instance_group: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:53.256366Z","hook":{"resource":{"addr":"google_compute_instance_group.instance_group","module":"","resource":"google_compute_instance_group.instance_group","implied_provider":"google","resource_type":"google_compute_instance_group","resource_name":"instance_group","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_instance_group.instance_group: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:03.259284Z","hook":{"resource":{"addr":"google_compute_instance_group.instance_group","module":"","resource":"google_compute_instance_group.instance_group","implied_provider":"google","resource_type":"google_compute_instance_group","resource_name":"instance_group","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_instance_group.instance_group: Creation complete after 11s [id=projects/example-project/zones/us-central1-a/instanceGroups/example-instance-ig]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:04.864269Z","hook":{"resource":{"addr":"google_compute_instance_group.instance_group","module":"","resource":"google_compute_instance_group.instance_group","implied_provider":"google","resource_type":"google_compute_instance_group","resource_name":"instance_group","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/zones/us-central1-a/instanceGroups/example-instance-ig","elapsed_seconds":11},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_backend_service.backend_service: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:04.879821Z","hook":{"resource":{"addr":"google_compute_backend_service.backend_service","module":"","resource":"google_compute_backend_service.backend_service","implied_provider":"google","resource_type":"google_compute_backend_service","resource_name":"backend_service","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_backend_service.backend_service: Still creating... [20s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:14.880179Z","hook":{"resource":{"addr":"google_compute_backend_service.backend_service","module":"","resource":"google_compute_backend_service.backend_service","implied_provider":"google","resource_type":"google_compute_backend_service","resource_name":"backend_service","resource_key":null},"action":"create","elapsed_seconds":20},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_backend_service.backend_service: Still creating... [30s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:24.881280Z","hook":{"resource":{"addr":"google_compute_backend_service.backend_service","module":"","resource":"google_compute_backend_service.backend_service","implied_provider":"google","resource_type":"google_compute_backend_service","resource_name":"backend_service","resource_key":null},"action":"create","elapsed_seconds":30},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_backend_service.backend_service: Still creating... [40s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:34.882970Z","hook":{"resource":{"addr":"google_compute_backend_service.backend_service","module":"","resource":"google_compute_backend_service.backend_service","implied_provider":"google","resource_type":"google_compute_backend_service","resource_name":"backend_service","resource_key":null},"action":"create","elapsed_seconds":40},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_backend_service.backend_service: Still creating... [50s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:44.883973Z","hook":{"resource":{"addr":"google_compute_backend_service.backend_service","module":"","resource":"google_compute_backend_service.backend_service","implied_provider":"google","resource_type":"google_compute_backend_service","resource_name":"backend_service","resource_key":null},"action":"create","elapsed_seconds":50},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_backend_service.backend_service: Creation complete after 54s [id=projects/example-project/global/backendServices/example-instance-bes]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:46.651481Z","hook":{"resource":{"addr":"google_compute_backend_service.backend_service","module":"","resource":"google_compute_backend_service.backend_service","implied_provider":"google","resource_type":"google_compute_backend_service","resource_name":"backend_service","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/backendServices/example-instance-bes","elapsed_seconds":54},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_url_map.url_map: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:46.677563Z","hook":{"resource":{"addr":"google_compute_url_map.url_map","module":"","resource":"google_compute_url_map.url_map","implied_provider":"google","resource_type":"google_compute_url_map","resource_name":"url_map","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_url_map.url_map: Still creating... [20s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:56.681348Z","hook":{"resource":{"addr":"google_compute_url_map.url_map","module":"","resource":"google_compute_url_map.url_map","implied_provider":"google","resource_type":"google_compute_url_map","resource_name":"url_map","resource_key":null},"action":"create","elapsed_seconds":20},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_url_map.url_map: Creation complete after 11s [id=projects/example-project/global/urlMaps/example-instance-urlmap]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:57.908775Z","hook":{"resource":{"addr":"google_compute_url_map.url_map","module":"","resource":"google_compute_url_map.url_map","implied_provider":"google","resource_type":"google_compute_url_map","resource_name":"url_map","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/urlMaps/example-instance-urlmap","elapsed_seconds":11},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_target_https_proxy.https_proxy: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:57.928063Z","hook":{"resource":{"addr":"google_compute_target_https_proxy.https_proxy","module":"","resource":"google_compute_target_https_proxy.https_proxy","implied_provider":"google","resource_type":"google_compute_target_https_proxy","resource_name":"https_proxy","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_target_http_proxy.http_proxy: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:57.929172Z","hook":{"resource":{"addr":"google_compute_target_http_proxy.http_proxy","module":"","resource":"google_compute_target_http_proxy.http_proxy","implied_provider":"google","resource_type":"google_compute_target_http_proxy","resource_name":"http_proxy","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_target_http_proxy.http_proxy: Creation complete after 11s [id=projects/example-project/global/targetHttpProxies/example-instance-http-proxy]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:59.127107Z","hook":{"resource":{"addr":"google_compute_target_http_proxy.http_proxy","module":"","resource":"google_compute_target_http_proxy.http_proxy","implied_provider":"google","resource_type":"google_compute_target_http_proxy","resource_name":"http_proxy","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/targetHttpProxies/example-instance-http-proxy","elapsed_seconds":11},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.http_forwarding_rule: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:59.136779Z","hook":{"resource":{"addr":"google_compute_global_forwarding_rule.http_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.http_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"http_forwarding_rule","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.https_forwarding_rule: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:59.196840Z","hook":{"resource":{"addr":"google_compute_global_forwarding_rule.https_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.https_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"https_forwarding_rule","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.https_forwarding_rule: Creation complete after 22s [id=projects/example-project/global/forwardingRules/example-instance-https-fwd-rule]","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:11.209168Z","hook":{"resource":{"addr":"google_compute_global_forwarding_rule.https_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.https_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"https_forwarding_rule","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/forwardingRules/example-instance-https-fwd-rule","elapsed_seconds":22},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.http_forwarding_rule: Creation complete after 23s [id=projects/example-project/global/forwardingRules/example-instance-http-fwd-rule]","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:11.593142Z","hook":{"resource":{"addr":"google_compute_global_forwarding_rule.http_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.http_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"http_forwarding_rule","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/forwardingRules/example-instance-http-fwd-rule","elapsed_seconds":23},"type":"apply_complete"}
{"@level":"info","@message":"Apply complete! Resources: 18 added, 0 changed, 0 destroyed.","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:11.613225Z","changes":{"add":18,"change":0,"import":0,"remove":0,"operation":"apply"},"type":"change_summary"}
{"@level":"info","@message":"Outputs: 8","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:11.613390Z","outputs":{"dns_name":{"sensitive":false,"type":"string","value":"example-instance.example.com"},"instance_name":{"sensitive":false,"type":"string","value":"example-instance"},"load_balancer_ip":{"sensitive":false,"type":"string","value":"10.0.0.1"},"private_ssh_key":{"sensitive":true,"type":"string"},"public_ip":{"sensitive":false,"type":"string","value":"10.0.0.2"},"public_ssh_key":{"sensitive":false,"type":"string","value":"ssh-rsa EXAMPLE_KEY"},"subnet_name":{"sensitive":false,"type":"string","value":"example-instance-subnet"},"vpc_name":{"sensitive":false,"type":"string","value":"example-instance-vpc"}},"type":"outputs"}
`

var mockTerraformDestroyOutput = `
{"@level":"info","@message":"Terraform 1.11.4","@module":"terraform.ui","@timestamp":"2025-05-10T10:00:57.945980-07:00","terraform":"1.11.4","type":"version","ui":"1.2"}
{"@level":"info","@message":"tls_private_key.ssh: Refreshing state... [id=abcdef1234567890abcdef1234567890abcdef12]","@module":"terraform.ui","@timestamp":"2025-05-10T10:01:00.215590-07:00","hook":{"resource":{"addr":"tls_private_key.ssh","module":"","resource":"tls_private_key.ssh","implied_provider":"tls","resource_type":"tls_private_key","resource_name":"ssh","resource_key":null},"id_key":"id","id_value":"abcdef1234567890abcdef1234567890abcdef12"},"type":"refresh_start"}
{"@level":"info","@message":"tls_private_key.ssh: Refresh complete [id=abcdef1234567890abcdef1234567890abcdef12]","@module":"terraform.ui","@timestamp":"2025-05-10T10:01:00.217654-07:00","hook":{"resource":{"addr":"tls_private_key.ssh","module":"","resource":"tls_private_key.ssh","implied_provider":"tls","resource_type":"tls_private_key","resource_name":"ssh","resource_key":null},"id_key":"id","id_value":"abcdef1234567890abcdef1234567890abcdef12"},"type":"refresh_complete"}
{"@level":"info","@message":"Plan: 0 to add, 0 to change, 18 to destroy.","@module":"terraform.ui","@timestamp":"2025-05-10T10:01:04.960408-07:00","changes":{"add":0,"change":0,"import":0,"remove":18,"operation":"plan"},"type":"change_summary"}
{"@level":"info","@message":"google_dns_record_set.dns_record: Destroying... [id=projects/example-project/managedZones/example.com/rrsets/example-instance.example.com./A]","@module":"terraform.ui","@timestamp":"2025-05-10T10:01:05.865228-07:00","hook":{"resource":{"addr":"google_dns_record_set.dns_record","module":"","resource":"google_dns_record_set.dns_record","implied_provider":"google","resource_type":"google_dns_record_set","resource_name":"dns_record","resource_key":null},"action":"delete","id_key":"id","id_value":"projects/example-project/managedZones/example.com/rrsets/example-instance.example.com./A"},"type":"apply_start"}
{"@level":"info","@message":"google_dns_record_set.dns_record: Destruction complete after 2s","@module":"terraform.ui","@timestamp":"2025-05-10T10:01:08.422318-07:00","hook":{"resource":{"addr":"google_dns_record_set.dns_record","module":"","resource":"google_dns_record_set.dns_record","implied_provider":"google","resource_type":"google_dns_record_set","resource_name":"dns_record","resource_key":null},"action":"delete","elapsed_seconds":2},"type":"apply_complete"}
{"@level":"info","@message":"Destroy complete! Resources: 18 destroyed.","@module":"terraform.ui","@timestamp":"2025-05-10T10:04:06.935983-07:00","changes":{"add":0,"change":0,"import":0,"remove":18,"operation":"destroy"},"type":"change_summary"}
`

var mockTerraformApplyOutputNoPlanSummary = `
{"@level":"info","@message":"Terraform 1.8.0","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:31.599035Z","terraform":"1.8.0","type":"version","ui":"1.2"}
{"@level":"info","@message":"tls_private_key.ssh: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352520Z","change":{"resource":{"addr":"tls_private_key.ssh","module":"","resource":"tls_private_key.ssh","implied_provider":"tls","resource_type":"tls_private_key","resource_name":"ssh","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_network.vpc: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352619Z","change":{"resource":{"addr":"google_compute_network.vpc","module":"","resource":"google_compute_network.vpc","implied_provider":"google","resource_type":"google_compute_network","resource_name":"vpc","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_global_address.lb_ip: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352632Z","change":{"resource":{"addr":"google_compute_global_address.lb_ip","module":"","resource":"google_compute_global_address.lb_ip","implied_provider":"google","resource_type":"google_compute_global_address","resource_name":"lb_ip","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_managed_ssl_certificate.ssl_certificate: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352639Z","change":{"resource":{"addr":"google_compute_managed_ssl_certificate.ssl_certificate","module":"","resource":"google_compute_managed_ssl_certificate.ssl_certificate","implied_provider":"google","resource_type":"google_compute_managed_ssl_certificate","resource_name":"ssl_certificate","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_address.public_ip: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352646Z","change":{"resource":{"addr":"google_compute_address.public_ip","module":"","resource":"google_compute_address.public_ip","implied_provider":"google","resource_type":"google_compute_address","resource_name":"public_ip","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_firewall.allow_ssh: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352653Z","change":{"resource":{"addr":"google_compute_firewall.allow_ssh","module":"","resource":"google_compute_firewall.allow_ssh","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_ssh","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_health_check.lb_health_check: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352659Z","change":{"resource":{"addr":"google_compute_health_check.lb_health_check","module":"","resource":"google_compute_health_check.lb_health_check","implied_provider":"google","resource_type":"google_compute_health_check","resource_name":"lb_health_check","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_subnetwork.subnet: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352664Z","change":{"resource":{"addr":"google_compute_subnetwork.subnet","module":"","resource":"google_compute_subnetwork.subnet","implied_provider":"google","resource_type":"google_compute_subnetwork","resource_name":"subnet","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_firewall.allow_lb_healthcheck: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352670Z","change":{"resource":{"addr":"google_compute_firewall.allow_lb_healthcheck","module":"","resource":"google_compute_firewall.allow_lb_healthcheck","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_lb_healthcheck","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_instance.vm: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352675Z","change":{"resource":{"addr":"google_compute_instance.vm","module":"","resource":"google_compute_instance.vm","implied_provider":"google","resource_type":"google_compute_instance","resource_name":"vm","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_instance_group.instance_group: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352699Z","change":{"resource":{"addr":"google_compute_instance_group.instance_group","module":"","resource":"google_compute_instance_group.instance_group","implied_provider":"google","resource_type":"google_compute_instance_group","resource_name":"instance_group","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_backend_service.backend_service: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352704Z","change":{"resource":{"addr":"google_compute_backend_service.backend_service","module":"","resource":"google_compute_backend_service.backend_service","implied_provider":"google","resource_type":"google_compute_backend_service","resource_name":"backend_service","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_url_map.url_map: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352710Z","change":{"resource":{"addr":"google_compute_url_map.url_map","module":"","resource":"google_compute_url_map.url_map","implied_provider":"google","resource_type":"google_compute_url_map","resource_name":"url_map","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_target_http_proxy.http_proxy: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352715Z","change":{"resource":{"addr":"google_compute_target_http_proxy.http_proxy","module":"","resource":"google_compute_target_http_proxy.http_proxy","implied_provider":"google","resource_type":"google_compute_target_http_proxy","resource_name":"http_proxy","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_target_https_proxy.https_proxy: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352721Z","change":{"resource":{"addr":"google_compute_target_https_proxy.https_proxy","module":"","resource":"google_compute_target_https_proxy.https_proxy","implied_provider":"google","resource_type":"google_compute_target_https_proxy","resource_name":"https_proxy","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.http_forwarding_rule: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352726Z","change":{"resource":{"addr":"google_compute_global_forwarding_rule.http_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.http_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"http_forwarding_rule","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.https_forwarding_rule: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352732Z","change":{"resource":{"addr":"google_compute_global_forwarding_rule.https_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.https_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"https_forwarding_rule","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_dns_record_set.dns_record: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:34.352738Z","change":{"resource":{"addr":"google_dns_record_set.dns_record","module":"","resource":"google_dns_record_set.dns_record","implied_provider":"google","resource_type":"google_dns_record_set","resource_name":"dns_record","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"tls_private_key.ssh: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:35.751910Z","hook":{"resource":{"addr":"tls_private_key.ssh","module":"","resource":"tls_private_key.ssh","implied_provider":"tls","resource_type":"tls_private_key","resource_name":"ssh","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_network.vpc: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:36.129747Z","hook":{"resource":{"addr":"google_compute_network.vpc","module":"","resource":"google_compute_network.vpc","implied_provider":"google","resource_type":"google_compute_network","resource_name":"vpc","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_managed_ssl_certificate.ssl_certificate: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:36.130899Z","hook":{"resource":{"addr":"google_compute_managed_ssl_certificate.ssl_certificate","module":"","resource":"google_compute_managed_ssl_certificate.ssl_certificate","implied_provider":"google","resource_type":"google_compute_managed_ssl_certificate","resource_name":"ssl_certificate","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_health_check.lb_health_check: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:36.133273Z","hook":{"resource":{"addr":"google_compute_health_check.lb_health_check","module":"","resource":"google_compute_health_check.lb_health_check","implied_provider":"google","resource_type":"google_compute_health_check","resource_name":"lb_health_check","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_global_address.lb_ip: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:36.133359Z","hook":{"resource":{"addr":"google_compute_global_address.lb_ip","module":"","resource":"google_compute_global_address.lb_ip","implied_provider":"google","resource_type":"google_compute_global_address","resource_name":"lb_ip","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_address.public_ip: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:36.133832Z","hook":{"resource":{"addr":"google_compute_address.public_ip","module":"","resource":"google_compute_address.public_ip","implied_provider":"google","resource_type":"google_compute_address","resource_name":"public_ip","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"tls_private_key.ssh: Creation complete after 3s [id=525020afa1595971e6caf04f70b19893a28f3211]","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:38.650469Z","hook":{"resource":{"addr":"tls_private_key.ssh","module":"","resource":"tls_private_key.ssh","implied_provider":"tls","resource_type":"tls_private_key","resource_name":"ssh","resource_key":null},"action":"create","id_key":"id","id_value":"525020afa1595971e6caf04f70b19893a28f3211","elapsed_seconds":3},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_network.vpc: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:46.131243Z","hook":{"resource":{"addr":"google_compute_network.vpc","module":"","resource":"google_compute_network.vpc","implied_provider":"google","resource_type":"google_compute_network","resource_name":"vpc","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_managed_ssl_certificate.ssl_certificate: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:46.133262Z","hook":{"resource":{"addr":"google_compute_managed_ssl_certificate.ssl_certificate","module":"","resource":"google_compute_managed_ssl_certificate.ssl_certificate","implied_provider":"google","resource_type":"google_compute_managed_ssl_certificate","resource_name":"ssl_certificate","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_health_check.lb_health_check: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:46.135293Z","hook":{"resource":{"addr":"google_compute_health_check.lb_health_check","module":"","resource":"google_compute_health_check.lb_health_check","implied_provider":"google","resource_type":"google_compute_health_check","resource_name":"lb_health_check","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_address.public_ip: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:46.135489Z","hook":{"resource":{"addr":"google_compute_address.public_ip","module":"","resource":"google_compute_address.public_ip","implied_provider":"google","resource_type":"google_compute_address","resource_name":"public_ip","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_global_address.lb_ip: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:46.135542Z","hook":{"resource":{"addr":"google_compute_global_address.lb_ip","module":"","resource":"google_compute_global_address.lb_ip","implied_provider":"google","resource_type":"google_compute_global_address","resource_name":"lb_ip","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_address.public_ip: Creation complete after 11s [id=projects/hon3y-356719/regions/us-central1/addresses/lemc-abcdef01-testuser-42-ind-public-ip]","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:47.254997Z","hook":{"resource":{"addr":"google_compute_address.public_ip","module":"","resource":"google_compute_address.public_ip","implied_provider":"google","resource_type":"google_compute_address","resource_name":"public_ip","resource_key":null},"action":"create","id_key":"id","id_value":"projects/hon3y-356719/regions/us-central1/addresses/lemc-abcdef01-testuser-42-ind-public-ip","elapsed_seconds":11},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_global_address.lb_ip: Creation complete after 12s [id=projects/hon3y-356719/global/addresses/lemc-abcdef01-testuser-42-ind-lb-ip]","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:47.563534Z","hook":{"resource":{"addr":"google_compute_global_address.lb_ip","module":"","resource":"google_compute_global_address.lb_ip","implied_provider":"google","resource_type":"google_compute_global_address","resource_name":"lb_ip","resource_key":null},"action":"create","id_key":"id","id_value":"projects/hon3y-356719/global/addresses/lemc-abcdef01-testuser-42-ind-lb-ip","elapsed_seconds":12},"type":"apply_complete"}
{"@level":"info","@message":"google_dns_record_set.dns_record: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:47.615263Z","hook":{"resource":{"addr":"google_dns_record_set.dns_record","module":"","resource":"google_dns_record_set.dns_record","implied_provider":"google","resource_type":"google_dns_record_set","resource_name":"dns_record","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_health_check.lb_health_check: Creation complete after 12s [id=projects/hon3y-356719/global/healthChecks/lemc-abcdef01-testuser-42-ind-hc]","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:47.773153Z","hook":{"resource":{"addr":"google_compute_health_check.lb_health_check","module":"","resource":"google_compute_health_check.lb_health_check","implied_provider":"google","resource_type":"google_compute_health_check","resource_name":"lb_health_check","resource_key":null},"action":"create","id_key":"id","id_value":"projects/hon3y-356719/global/healthChecks/lemc-abcdef01-testuser-42-ind-hc","elapsed_seconds":12},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_managed_ssl_certificate.ssl_certificate: Creation complete after 12s [id=projects/hon3y-356719/global/sslCertificates/lemc-abcdef01-testuser-42-ind-ssl-cert]","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:47.837658Z","hook":{"resource":{"addr":"google_compute_managed_ssl_certificate.ssl_certificate","module":"","resource":"google_compute_managed_ssl_certificate.ssl_certificate","implied_provider":"google","resource_type":"google_compute_managed_ssl_certificate","resource_name":"ssl_certificate","resource_key":null},"action":"create","id_key":"id","id_value":"projects/hon3y-356719/global/sslCertificates/lemc-abcdef01-testuser-42-ind-ssl-cert","elapsed_seconds":12},"type":"apply_complete"}
{"@level":"info","@message":"google_dns_record_set.dns_record: Creation complete after 2s [id=projects/hon3y-356719/managedZones/w-a-s-d-com/rrsets/lemc-abcdef01-testuser-42-ind.w-a-s-d.com./A]","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:50.283336Z","hook":{"resource":{"addr":"google_dns_record_set.dns_record","module":"","resource":"google_dns_record_set.dns_record","implied_provider":"google","resource_type":"google_dns_record_set","resource_name":"dns_record","resource_key":null},"action":"create","id_key":"id","id_value":"projects/hon3y-356719/managedZones/w-a-s-d-com/rrsets/lemc-abcdef01-testuser-42-ind.w-a-s-d.com./A","elapsed_seconds":2},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_network.vpc: Still creating... [20s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:56.133240Z","hook":{"resource":{"addr":"google_compute_network.vpc","module":"","resource":"google_compute_network.vpc","implied_provider":"google","resource_type":"google_compute_network","resource_name":"vpc","resource_key":null},"action":"create","elapsed_seconds":20},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_network.vpc: Creation complete after 22s [id=projects/hon3y-356719/global/networks/lemc-abcdef01-testuser-42-ind-vpc]","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:57.729396Z","hook":{"resource":{"addr":"google_compute_network.vpc","module":"","resource":"google_compute_network.vpc","implied_provider":"google","resource_type":"google_compute_network","resource_name":"vpc","resource_key":null},"action":"create","id_key":"id","id_value":"projects/hon3y-356719/global/networks/lemc-abcdef01-testuser-42-ind-vpc","elapsed_seconds":22},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_firewall.allow_ssh: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:57.749673Z","hook":{"resource":{"addr":"google_compute_firewall.allow_ssh","module":"","resource":"google_compute_firewall.allow_ssh","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_ssh","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_subnetwork.subnet: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:57.749940Z","hook":{"resource":{"addr":"google_compute_subnetwork.subnet","module":"","resource":"google_compute_subnetwork.subnet","implied_provider":"google","resource_type":"google_compute_subnetwork","resource_name":"subnet","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_firewall.allow_lb_healthcheck: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T18:59:57.751062Z","hook":{"resource":{"addr":"google_compute_firewall.allow_lb_healthcheck","module":"","resource":"google_compute_firewall.allow_lb_healthcheck","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_lb_healthcheck","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_firewall.allow_ssh: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:00:07.750355Z","hook":{"resource":{"addr":"google_compute_firewall.allow_ssh","module":"","resource":"google_compute_firewall.allow_ssh","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_ssh","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_firewall.allow_lb_healthcheck: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:00:07.752214Z","hook":{"resource":{"addr":"google_compute_firewall.allow_lb_healthcheck","module":"","resource":"google_compute_firewall.allow_lb_healthcheck","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_lb_healthcheck","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_subnetwork.subnet: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:00:07.752297Z","hook":{"resource":{"addr":"google_compute_subnetwork.subnet","module":"","resource":"google_compute_subnetwork.subnet","implied_provider":"google","resource_type":"google_compute_subnetwork","resource_name":"subnet","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_subnetwork.subnet: Creation complete after 11s [id=projects/hon3y-356719/regions/us-central1/subnetworks/lemc-abcdef01-testuser-42-ind-subnet]","@module":"terraform.ui","@timestamp":"2025-05-10T19:00:08.801735Z","hook":{"resource":{"addr":"google_compute_subnetwork.subnet","module":"","resource":"google_compute_subnetwork.subnet","implied_provider":"google","resource_type":"google_compute_subnetwork","resource_name":"subnet","resource_key":null},"action":"create","id_key":"id","id_value":"projects/hon3y-356719/regions/us-central1/subnetworks/lemc-abcdef01-testuser-42-ind-subnet","elapsed_seconds":11},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_instance.vm: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T19:00:08.819074Z","hook":{"resource":{"addr":"google_compute_instance.vm","module":"","resource":"google_compute_instance.vm","implied_provider":"google","resource_type":"google_compute_instance","resource_name":"vm","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_instance.vm: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:00:18.819807Z","hook":{"resource":{"addr":"google_compute_instance.vm","module":"","resource":"google_compute_instance.vm","implied_provider":"google","resource_type":"google_compute_instance","resource_name":"vm","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_instance.vm: Creation complete after 13s [id=projects/hon3y-356719/zones/us-central1-a/instances/lemc-abcdef01-testuser-42-ind]","@module":"terraform.ui","@timestamp":"2025-05-10T19:00:21.531236Z","hook":{"resource":{"addr":"google_compute_instance.vm","module":"","resource":"google_compute_instance.vm","implied_provider":"google","resource_type":"google_compute_instance","resource_name":"vm","resource_key":null},"action":"create","id_key":"id","id_value":"projects/hon3y-356719/zones/us-central1-a/instances/lemc-abcdef01-testuser-42-ind","elapsed_seconds":13},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_instance_group.instance_group: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T19:00:21.540005Z","hook":{"resource":{"addr":"google_compute_instance_group.instance_group","module":"","resource":"google_compute_instance_group.instance_group","implied_provider":"google","resource_type":"google_compute_instance_group","resource_name":"instance_group","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_instance_group.instance_group: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:00:31.542974Z","hook":{"resource":{"addr":"google_compute_instance_group.instance_group","module":"","resource":"google_compute_instance_group.instance_group","implied_provider":"google","resource_type":"google_compute_instance_group","resource_name":"instance_group","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_instance_group.instance_group: Creation complete after 11s [id=projects/hon3y-356719/zones/us-central1-a/instanceGroups/lemc-abcdef01-testuser-42-ind-ig]","@module":"terraform.ui","@timestamp":"2025-05-10T19:00:33.400148Z","hook":{"resource":{"addr":"google_compute_instance_group.instance_group","module":"","resource":"google_compute_instance_group.instance_group","implied_provider":"google","resource_type":"google_compute_instance_group","resource_name":"instance_group","resource_key":null},"action":"create","id_key":"id","id_value":"projects/hon3y-356719/zones/us-central1-a/instanceGroups/lemc-abcdef01-testuser-42-ind-ig","elapsed_seconds":11},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_backend_service.backend_service: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:00:33.415495Z","hook":{"resource":{"addr":"google_compute_backend_service.backend_service","module":"","resource":"google_compute_backend_service.backend_service","implied_provider":"google","resource_type":"google_compute_backend_service","resource_name":"backend_service","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_backend_service.backend_service: Still creating... [20s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:00:43.419284Z","hook":{"resource":{"addr":"google_compute_backend_service.backend_service","module":"","resource":"google_compute_backend_service.backend_service","implied_provider":"google","resource_type":"google_compute_backend_service","resource_name":"backend_service","resource_key":null},"action":"create","elapsed_seconds":20},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_backend_service.backend_service: Still creating... [30s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:00:53.420179Z","hook":{"resource":{"addr":"google_compute_backend_service.backend_service","module":"","resource":"google_compute_backend_service.backend_service","implied_provider":"google","resource_type":"google_compute_backend_service","resource_name":"backend_service","resource_key":null},"action":"create","elapsed_seconds":30},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_backend_service.backend_service: Still creating... [40s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:01:03.421280Z","hook":{"resource":{"addr":"google_compute_backend_service.backend_service","module":"","resource":"google_compute_backend_service.backend_service","implied_provider":"google","resource_type":"google_compute_backend_service","resource_name":"backend_service","resource_key":null},"action":"create","elapsed_seconds":40},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_backend_service.backend_service: Still creating... [50s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:01:13.422970Z","hook":{"resource":{"addr":"google_compute_backend_service.backend_service","module":"","resource":"google_compute_backend_service.backend_service","implied_provider":"google","resource_type":"google_compute_backend_service","resource_name":"backend_service","resource_key":null},"action":"create","elapsed_seconds":50},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_backend_service.backend_service: Creation complete after 54s [id=projects/hon3y-356719/global/backendServices/lemc-abcdef01-testuser-42-ind-bes]","@module":"terraform.ui","@timestamp":"2025-05-10T19:01:26.651481Z","hook":{"resource":{"addr":"google_compute_backend_service.backend_service","module":"","resource":"google_compute_backend_service.backend_service","implied_provider":"google","resource_type":"google_compute_backend_service","resource_name":"backend_service","resource_key":null},"action":"create","id_key":"id","id_value":"projects/hon3y-356719/global/backendServices/lemc-abcdef01-testuser-42-ind-bes","elapsed_seconds":54},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_url_map.url_map: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:01:13.423852Z","hook":{"resource":{"addr":"google_compute_url_map.url_map","module":"","resource":"google_compute_url_map.url_map","implied_provider":"google","resource_type":"google_compute_url_map","resource_name":"url_map","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_url_map.url_map: Still creating... [20s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:01:23.424970Z","hook":{"resource":{"addr":"google_compute_url_map.url_map","module":"","resource":"google_compute_url_map.url_map","implied_provider":"google","resource_type":"google_compute_url_map","resource_name":"url_map","resource_key":null},"action":"create","elapsed_seconds":20},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_url_map.url_map: Creation complete after 11s [id=projects/hon3y-356719/global/urlMaps/lemc-abcdef01-testuser-42-ind-urlmap]","@module":"terraform.ui","@timestamp":"2025-05-10T19:01:24.908775Z","hook":{"resource":{"addr":"google_compute_url_map.url_map","module":"","resource":"google_compute_url_map.url_map","implied_provider":"google","resource_type":"google_compute_url_map","resource_name":"url_map","resource_key":null},"action":"create","id_key":"id","id_value":"projects/hon3y-356719/global/urlMaps/lemc-abcdef01-testuser-42-ind-urlmap","elapsed_seconds":11},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_target_https_proxy.https_proxy: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:01:24.928063Z","hook":{"resource":{"addr":"google_compute_target_https_proxy.https_proxy","module":"","resource":"google_compute_target_https_proxy.https_proxy","implied_provider":"google","resource_type":"google_compute_target_https_proxy","resource_name":"https_proxy","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_target_http_proxy.http_proxy: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:01:24.929172Z","hook":{"resource":{"addr":"google_compute_target_http_proxy.http_proxy","module":"","resource":"google_compute_target_http_proxy.http_proxy","implied_provider":"google","resource_type":"google_compute_target_http_proxy","resource_name":"http_proxy","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_target_http_proxy.http_proxy: Creation complete after 11s [id=projects/hon3y-356719/global/targetHttpProxies/lemc-abcdef01-testuser-42-ind-http-proxy]","@module":"terraform.ui","@timestamp":"2025-05-10T19:01:25.127107Z","hook":{"resource":{"addr":"google_compute_target_http_proxy.http_proxy","module":"","resource":"google_compute_target_http_proxy.http_proxy","implied_provider":"google","resource_type":"google_compute_target_http_proxy","resource_name":"http_proxy","resource_key":null},"action":"create","id_key":"id","id_value":"projects/hon3y-356719/global/targetHttpProxies/lemc-abcdef01-testuser-42-ind-http-proxy","elapsed_seconds":11},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.http_forwarding_rule: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:01:25.136779Z","hook":{"resource":{"addr":"google_compute_global_forwarding_rule.http_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.http_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"http_forwarding_rule","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.https_forwarding_rule: Still creating... [10s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:01:35.137909Z","hook":{"resource":{"addr":"google_compute_global_forwarding_rule.http_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.http_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"http_forwarding_rule","resource_key":null},"action":"create","elapsed_seconds":10},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.https_forwarding_rule: Still creating... [20s elapsed]","@module":"terraform.ui","@timestamp":"2025-05-10T19:01:45.139746Z","hook":{"resource":{"addr":"google_compute_global_forwarding_rule.https_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.https_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"https_forwarding_rule","resource_key":null},"action":"create","elapsed_seconds":20},"type":"apply_progress"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.https_forwarding_rule: Creation complete after 22s [id=projects/hon3y-356719/global/forwardingRules/lemc-abcdef01-testuser-42-ind-https-fwd-rule]","@module":"terraform.ui","@timestamp":"2025-05-10T19:01:52.567767Z","hook":{"resource":{"addr":"google_compute_global_forwarding_rule.https_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.https_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"https_forwarding_rule","resource_key":null},"action":"create","id_key":"id","id_value":"projects/hon3y-356719/global/forwardingRules/lemc-abcdef01-testuser-42-ind-https-fwd-rule","elapsed_seconds":22},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.http_forwarding_rule: Creation complete after 23s [id=projects/hon3y-356719/global/forwardingRules/lemc-abcdef01-testuser-42-ind-http-fwd-rule]","@module":"terraform.ui","@timestamp":"2025-05-10T19:01:53.045954Z","hook":{"resource":{"addr":"google_compute_global_forwarding_rule.http_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.http_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"http_forwarding_rule","resource_key":null},"action":"create","id_key":"id","id_value":"projects/hon3y-356719/global/forwardingRules/lemc-abcdef01-testuser-42-ind-http-fwd-rule","elapsed_seconds":23},"type":"apply_complete"}
{"@level":"info","@message":"Apply complete! Resources: 18 added, 0 changed, 0 destroyed.","@module":"terraform.ui","@timestamp":"2025-05-10T19:01:53.060601Z","changes":{"add":18,"change":0,"import":0,"remove":0,"operation":"apply"},"type":"change_summary"}
{"@level":"info","@message":"Outputs: 8","@module":"terraform.ui","@timestamp":"2025-05-10T19:01:53.060662Z","outputs":{"dns_name":{"sensitive":false,"type":"string","value":"lemc-abcdef01-testuser-42-ind.w-a-s-d.com"},"instance_name":{"sensitive":false,"type":"string","value":"lemc-abcdef01-testuser-42-ind"},"load_balancer_ip":{"sensitive":false,"type":"string","value":"34.54.77.204"},"private_ssh_key":{"sensitive":true,"type":"string"},"public_ip":{"sensitive":false,"type":"string","value":"34.122.141.130"},"public_ssh_key":{"sensitive":false,"type":"string","value":"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC0Lz2VxLSGpa4MROY0EkPpQS5IRx7frqvZolQYSH0hEF8Nc3mKWl37y/VCJabPlxr3/tmGg+yKAr7fdG0P75n9E0vRrm5R7SKfxN7LCAcLGZJM++OHYFoQ43aO2yeui3NVlxCwC0i1JNgZMrIl17uG9tLul/Om0luIf7oBh9t4sxrHvX4j2Vj3a68rnNwlSBuZ02GldOwKVlDYF2rJU02ndR+B+n3wgw0geaTtgYWLE2dFOu3hAwB5n+DQy+WXlcz+3OpCfzguEHPtirI9p9aC+BFNdknpVqj/KR2FMFSt7xAtZ5JFyv0QO0BWdx0mOBeouiKWG+853l/6fRCJpNYLr+iCg7rmB0Jf4oFvgyW1C+WuZJEFIybO9tjWty0hqAiyMTz4ld1xXP6M1JpwaKztu7nLCIHpCW/dqDuAH2yiqod9/eLcU8rc6uQQ6b1tFuUdMQeiU9HlxtwV/RRahi562glishNkrZoaagZL+htB503SaB4LZqKZc6y/8e8kErIJF/0Wkt9PI9uWzT1BNfEI4+/Y1UWYMAevDYHcEOG0SfVVHb/uO1JMzJjvziwhyr6vYx1zNfEsqVx9qIia1a7I/8YJGxfFnROlx7uEaDRRXprCc1twhj0vEJfwi6/EPjzjlsQdUAPAfzoUA9ZaQ0K4GRhoCTFHI7fz//RbPVDlbw==\\n"},"subnet_name":{"sensitive":false,"type":"string","value":"lemc-abcdef01-testuser-42-ind-subnet"},"vpc_name":{"sensitive":false,"type":"string","value":"lemc-abcdef01-testuser-42-ind-vpc"}},"type":"outputs"}
`

var mockTerraformSingleDigitOutput = `
{"@level":"info","@message":"Terraform 1.8.0","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:47.831898Z","terraform":"1.8.0","type":"version","ui":"1.2"}
{"@level":"info","@message":"Plan: 1 to add, 0 to change, 0 to destroy.","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242882Z","changes":{"add":1,"change":0,"import":0,"remove":0,"operation":"plan"},"type":"change_summary"}
{"@level":"info","@message":"mock_resource.example: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:55.664347Z","hook":{"resource":{"addr":"mock_resource.example","module":"","resource":"mock_resource.example","implied_provider":"mock","resource_type":"mock_resource","resource_name":"example","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"mock_resource.example: Creation complete after 0s [id=mock_id]","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:56.417889Z","hook":{"resource":{"addr":"mock_resource.example","module":"","resource":"mock_resource.example","implied_provider":"mock","resource_type":"mock_resource","resource_name":"example","resource_key":null},"action":"create","id_key":"id","id_value":"mock_id","elapsed_seconds":0},"type":"apply_complete"}
{"@level":"info","@message":"Apply complete! Resources: 1 added, 0 changed, 0 destroyed.","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:11.613225Z","changes":{"add":1,"change":0,"import":0,"remove":0,"operation":"apply"},"type":"change_summary"}
`

func TestProcessJSONStream(t *testing.T) {
	// Create a reader with the mock Terraform output
	reader := strings.NewReader(mockTerraformOutput)

	// Create a progress handler
	handler := progress.NewProgressHandler(reader)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Read all lines with context
	var lines []string
	var err error

	for {
		select {
		case <-ctx.Done():
			t.Fatal("Test timed out after 5 seconds")
		default:
			line, readErr := handler.ReadLine()
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

	// Check for total steps
	expectedTotalStepsStr := "(36)"
	foundTotalSteps := false
	for _, line := range lines {
		if strings.Contains(line, expectedTotalStepsStr) {
			foundTotalSteps = true
			break
		}
	}
	if !foundTotalSteps {
		t.Errorf("Output does not contain expected total steps string '%s'. Output:\n%s", expectedTotalStepsStr, strings.Join(lines, "\n"))
	}

	// Check for a specific "Creating..." message
	expectedCreatingMessage := "tls_private_key.ssh: Creating..."
	foundCreatingMessage := false
	for _, line := range lines {
		if strings.Contains(line, expectedCreatingMessage) {
			foundCreatingMessage = true
			break
		}
	}
	if !foundCreatingMessage {
		t.Errorf("Output does not contain expected creating message '%s'. Output:\n%s", expectedCreatingMessage, strings.Join(lines, "\n"))
	}

	// Check for a specific "Creation complete..." message
	expectedCompleteMessage := "Creation complete"
	foundCompleteMessage := false
	for _, line := range lines {
		if strings.Contains(line, expectedCompleteMessage) {
			foundCompleteMessage = true
			break
		}
	}
	if !foundCompleteMessage {
		t.Errorf("Output does not contain expected completion message '%s'. Output:\n%s", expectedCompleteMessage, strings.Join(lines, "\n"))
	}

	// Check for progress bar structure (e.g., presence of '[=')
	foundProgressBar := false
	for _, line := range lines {
		if strings.Contains(line, "[=") || strings.Contains(line, "[-/-]") {
			foundProgressBar = true
			break
		}
	}
	if !foundProgressBar {
		t.Errorf("Output does not seem to contain a progress bar structure like '[='. Output:\n%s", strings.Join(lines, "\n"))
	}
}

func TestProcessJSONStream_ApplyFromPlanFile(t *testing.T) {
	reader := strings.NewReader(mockTerraformApplyOutputNoPlanSummary)
	handler := progress.NewProgressHandler(reader)

	var lines []string
	var err error

	// Read all lines with context to prevent test hanging
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Increased timeout for potentially long stream
	defer cancel()

	firstApplyStartProcessed := false
	var planningPhaseLine string
	var applyPhaseLine string
	var applyCompleteLine string

	// First line is always initial "Planning..."
	initialLine, readErr := handler.ReadLine()
	if readErr != nil {
		t.Fatalf("Failed to read initial line: %v", readErr)
	}
	if !strings.Contains(initialLine, "[      PLANNING      ] Planning...") {
		t.Errorf("Expected initial line to be planning, got: %s", initialLine)
	}

	for {
		var line string
		var readErr error
		select {
		case <-ctx.Done():
			t.Logf("Collected lines before timeout:\n%s", strings.Join(lines, "\n"))
			t.Fatal("Test timed out processing stream")
			return
		default:
			line, readErr = handler.ReadLine()
			if readErr != nil {
				if readErr == io.EOF {
					goto done
				}
				err = readErr
				goto done
			}
			if line != "" {
				lines = append(lines, line)
				// Capture a line from the "planned_change" phase (still considered planning by current logic)
				if strings.Contains(line, "Plan to create") && planningPhaseLine == "" {
					planningPhaseLine = line
				}
				// Capture the first line from "apply_start" phase
				if strings.Contains(line, ": Creating...") && !firstApplyStartProcessed {
					applyPhaseLine = line
					firstApplyStartProcessed = true
				}
				// Capture the "Apply complete" line
				if strings.Contains(line, "Apply complete!") {
					applyCompleteLine = line
				}
			}
		}
	}
done:

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if len(lines) == 0 {
		t.Error("Expected non-empty output from handler")
		return
	}

	// Verify planning phase line shows "[      PLANNING      ]"
	if planningPhaseLine == "" {
		t.Error("Did not capture a planning phase line (e.g., 'Plan to create')")
	} else if !strings.Contains(planningPhaseLine, "[      PLANNING      ]") {
		t.Errorf("Expected planning phase line to contain '[      PLANNING      ]', got: %s", planningPhaseLine)
	}

	// Verify apply phase line now shows a proper progress bar, not the static one.
	// Total steps should be 18 resources * 2 = 36.
	// The first apply step will be (1)[=...](36)
	expectedApplyBarPattern := "(01)[=" // Check for the start of a correct progress bar
	expectedTotalInBar := "(36)"        // Check that the total is correctly identified

	if applyPhaseLine == "" {
		t.Error("Did not capture an apply phase line (e.g., 'Creating...')")
	} else {
		if !strings.Contains(applyPhaseLine, expectedApplyBarPattern) {
			t.Errorf("Expected apply phase line to start with pattern like '%s', got: %s", expectedApplyBarPattern, applyPhaseLine)
		}
		if !strings.Contains(applyPhaseLine, expectedTotalInBar) {
			t.Errorf("Expected apply phase line to contain total steps '%s', got: %s", expectedTotalInBar, applyPhaseLine)
		}
	}

	// Verify "Apply complete!" line also shows the correct bar, fully progressed
	expectedApplyCompletePattern := "(36)[=" // Fully progressed
	if applyCompleteLine == "" {
		t.Error("Did not capture 'Apply complete!' line")
	} else {
		if !strings.Contains(applyCompleteLine, expectedApplyCompletePattern) {
			t.Errorf("Expected 'Apply complete!' line to start with pattern like '%s', got: %s", expectedApplyCompletePattern, applyCompleteLine)
		}
		if !strings.Contains(applyCompleteLine, expectedTotalInBar) {
			t.Errorf("Expected 'Apply complete!' line to contain total steps '%s', got: %s", expectedTotalInBar, applyCompleteLine)
		}
	}

	// For debugging, print a few lines if asserts fail
	if t.Failed() {
		t.Log("Sample of processed lines for ApplyFromPlanFile test:")
		sampleSize := 10
		if len(lines) < sampleSize {
			sampleSize = len(lines)
		}
		for i := 0; i < sampleSize; i++ {
			t.Logf("Line %d: %s", i, lines[i])
		}
		if len(lines) > sampleSize {
			t.Log("...")
			for i := len(lines) - sampleSize; i < len(lines); i++ {
				if i >= 0 { // ensure i is not negative
					t.Logf("Line %d: %s", i, lines[i])
				}
			}
		}
	}
}

func TestProcessJSONStream_Destroy(t *testing.T) {
	// Create a reader with the mock Terraform destroy output
	reader := strings.NewReader(mockTerraformDestroyOutput)

	// Create a progress handler
	handler := progress.NewProgressHandler(reader)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Read all lines with context
	var lines []string
	var err error

	for {
		select {
		case <-ctx.Done():
			t.Fatal("Test timed out after 5 seconds")
		default:
			line, readErr := handler.ReadLine()
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

	// Check for total steps
	expectedTotalStepsStr := "(36)" // Destroy operation has 36 total steps
	foundTotalSteps := false
	for _, line := range lines {
		if strings.Contains(line, expectedTotalStepsStr) {
			foundTotalSteps = true
			break
		}
	}
	if !foundTotalSteps {
		t.Errorf("Output does not contain expected total steps string '%s'. Output:\n%s", expectedTotalStepsStr, strings.Join(lines, "\n"))
	}

	// Check for a specific "Destroying..." message
	expectedDestroyingMessage := "google_dns_record_set.dns_record: Destroying..."
	foundDestroyingMessage := false
	for _, line := range lines {
		if strings.Contains(line, expectedDestroyingMessage) {
			foundDestroyingMessage = true
			break
		}
	}
	if !foundDestroyingMessage {
		t.Errorf("Output does not contain expected destroying message '%s'. Output:\n%s", expectedDestroyingMessage, strings.Join(lines, "\n"))
	}

	// Check for a specific "Destruction complete..." message
	expectedCompleteMessage := "Destruction..."
	foundCompleteMessage := false
	for _, line := range lines {
		if strings.Contains(line, expectedCompleteMessage) {
			foundCompleteMessage = true
			break
		}
	}
	if !foundCompleteMessage {
		t.Errorf("Output does not contain expected completion message '%s'. Output:\n%s", expectedCompleteMessage, strings.Join(lines, "\n"))
	}

	// Check for the final "Destroy complete!" message
	expectedDestroyComplete := "Destroy complete! Resources: 18 destroyed"
	foundDestroyComplete := false
	for _, line := range lines {
		if strings.Contains(line, expectedDestroyComplete) {
			foundDestroyComplete = true
			break
		}
	}
	if !foundDestroyComplete {
		t.Errorf("Output does not contain expected final destroy complete message '%s'. Output:\n%s", expectedDestroyComplete, strings.Join(lines, "\n"))
	}

	// Check for progress bar structure (e.g., presence of '[=')
	foundProgressBar := false
	for _, line := range lines {
		if strings.Contains(line, "[=") || strings.Contains(line, "[-/-]") {
			foundProgressBar = true
			break
		}
	}
	if !foundProgressBar {
		t.Errorf("Output does not seem to contain a progress bar structure like '[='. Output:\n%s", strings.Join(lines, "\n"))
	}
}

func TestGetProgressString_SingleDigitPadding(t *testing.T) {
	reader := strings.NewReader(mockTerraformSingleDigitOutput)
	handler := progress.NewProgressHandler(reader)

	var lines []string
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Read initial planning line
	initialLine, readErr := handler.ReadLine()
	if readErr != nil {
		t.Fatalf("Failed to read initial line: %v", readErr)
	}
	lines = append(lines, initialLine)
	if !strings.Contains(initialLine, "[      PLANNING      ] Planning...") {
		t.Errorf("Expected initial line to be planning, got: %s", initialLine)
	}

	for {
		select {
		case <-ctx.Done():
			err = fmt.Errorf("test timed out: %w", ctx.Err())
			goto done
		default:
			line, readErr := handler.ReadLine()
			if readErr != nil {
				if readErr == io.EOF {
					goto done
				}
				err = readErr
				goto done
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

	if len(lines) < 4 { // Initial, Plan Summary, Creating, Complete, Apply Summary
		t.Fatalf("Expected at least 4 lines of output, got %d. Output:\\n%s", len(lines), strings.Join(lines, "\\n"))
	}

	// Example: Plan summary message should show (02) for total
	// (00)[PLANNING](02) Plan: 1 to add...
	planSummaryLine := ""
	for _, l := range lines {
		if strings.Contains(l, "Plan: 1 to add") {
			planSummaryLine = l
			break
		}
	}
	if planSummaryLine == "" {
		t.Fatal("Could not find 'Plan: 1 to add' line in output")
	}
	expectedPlanSummaryPattern := "(00)[  ](02) Plan: 1 to add"
	if !strings.HasPrefix(planSummaryLine, expectedPlanSummaryPattern) {
		t.Errorf("Expected plan summary line to start with '%s', got '%s'", expectedPlanSummaryPattern, planSummaryLine)
	}

	// Example: Creating message (current=1, total=2)
	// (01)[=         ](02) mock_resource.example: Creating...
	creatingLine := ""
	for _, l := range lines {
		if strings.Contains(l, "mock_resource.example: Creating...") {
			creatingLine = l
			break
		}
	}
	if creatingLine == "" {
		t.Fatal("Could not find 'mock_resource.example: Creating...' line in output")
	}
	// Bar width will be totalSteps (2 in this case)
	// (01)[= ](02) mock_resource.example: Creating...
	expectedCreatingPattern := "(01)[= ](02) mock_resource.example: Creating..."
	if !strings.HasPrefix(creatingLine, expectedCreatingPattern) {
		t.Errorf("Expected creating line to start with '%s', got '%s'", expectedCreatingPattern, creatingLine)
	}

	// Example: Creation complete message (current=2, total=2)
	// (02)[==](02) mock_resource.example: Creation complete...
	completeLine := ""
	for _, l := range lines {
		if strings.Contains(l, "mock_resource.example: Creation complete") {
			completeLine = l
			break
		}
	}
	if completeLine == "" {
		t.Fatal("Could not find 'mock_resource.example: Creation complete' line in output")
	}
	expectedCompletePattern := "(02)[==](02) mock_resource.example: Creation complete"
	if !strings.HasPrefix(completeLine, expectedCompletePattern) {
		t.Errorf("Expected complete line to start with '%s', got '%s'", expectedCompletePattern, completeLine)
	}

	// Example: Apply complete message
	// (02)[==](02) Apply complete! Resources: 1 added
	applySummaryLine := ""
	for _, l := range lines {
		if strings.Contains(l, "Apply complete! Resources: 1 added") {
			applySummaryLine = l
			break
		}
	}
	if applySummaryLine == "" {
		t.Fatal("Could not find 'Apply complete! Resources: 1 added' line in output")
	}
	expectedApplySummaryPattern := "(02)[==](02) Apply complete! Resources: 1 added"
	if !strings.HasPrefix(applySummaryLine, expectedApplySummaryPattern) {
		t.Errorf("Expected apply summary line to start with '%s', got '%s'", expectedApplySummaryPattern, applySummaryLine)
	}

	if t.Failed() {
		t.Logf("Full output for TestGetProgressString_SingleDigitPadding:\\n%s", strings.Join(lines, "\\n"))
	}
}
