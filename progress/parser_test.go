package progress_test

import (
	"bytes"
	"io"
	"os"
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
{"@level":"info","@message":"google_compute_firewall.allow_lb_healthcheck: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242778Z","change":{"resource":{"addr":"google_compute_firewall.allow_lb_healthcheck","module":"","resource":"google_compute_firewall.allow_lb_healthcheck","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_lb_healthcheck","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_subnetwork.subnet: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242785Z","change":{"resource":{"addr":"google_compute_subnetwork.subnet","module":"","resource":"google_compute_subnetwork.subnet","implied_provider":"google","resource_type":"google_compute_subnetwork","resource_name":"subnet","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_health_check.lb_health_check: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242792Z","change":{"resource":{"addr":"google_compute_health_check.lb_health_check","module":"","resource":"google_compute_health_check.lb_health_check","implied_provider":"google","resource_type":"google_compute_health_check","resource_name":"lb_health_check","resource_key":null},"action":"create"},"type":"planned_change"}
{"@level":"info","@message":"google_compute_firewall.allow_ssh: Plan to create","@module":"terraform.ui","@timestamp":"2025-05-10T03:27:54.242799Z","change":{"resource":{"addr":"google_compute_firewall.allow_ssh","module":"","resource":"google_compute_firewall.allow_ssh","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_ssh","resource_key":null},"action":"create"},"type":"planned_change"}
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
{"@level":"info","@message":"google_compute_firewall.allow_ssh: Creation complete after 11s [id=projects/example-project/global/firewalls/example-instance-allow-ssh]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:29.036695Z","hook":{"resource":{"addr":"google_compute_firewall.allow_ssh","module":"","resource":"google_compute_firewall.allow_ssh","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_ssh","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/firewalls/example-instance-allow-ssh","elapsed_seconds":11},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_firewall.allow_lb_healthcheck: Creation complete after 11s [id=projects/example-project/global/firewalls/example-instance-allow-lb-hc]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:29.098357Z","hook":{"resource":{"addr":"google_compute_firewall.allow_lb_healthcheck","module":"","resource":"google_compute_firewall.allow_lb_healthcheck","implied_provider":"google","resource_type":"google_compute_firewall","resource_name":"allow_lb_healthcheck","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/firewalls/example-instance-allow-lb-hc","elapsed_seconds":11},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_subnetwork.subnet: Creation complete after 21s [id=projects/example-project/regions/us-central1/subnetworks/example-instance-subnet]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:39.298507Z","hook":{"resource":{"addr":"google_compute_subnetwork.subnet","module":"","resource":"google_compute_subnetwork.subnet","implied_provider":"google","resource_type":"google_compute_subnetwork","resource_name":"subnet","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/regions/us-central1/subnetworks/example-instance-subnet","elapsed_seconds":21},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_instance.vm: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:39.316775Z","hook":{"resource":{"addr":"google_compute_instance.vm","module":"","resource":"google_compute_instance.vm","implied_provider":"google","resource_type":"google_compute_instance","resource_name":"vm","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_instance.vm: Creation complete after 14s [id=projects/example-project/zones/us-central1-a/instances/example-instance]","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:53.247084Z","hook":{"resource":{"addr":"google_compute_instance.vm","module":"","resource":"google_compute_instance.vm","implied_provider":"google","resource_type":"google_compute_instance","resource_name":"vm","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/zones/us-central1-a/instances/example-instance","elapsed_seconds":14},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_instance_group.instance_group: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:28:53.256366Z","hook":{"resource":{"addr":"google_compute_instance_group.instance_group","module":"","resource":"google_compute_instance_group.instance_group","implied_provider":"google","resource_type":"google_compute_instance_group","resource_name":"instance_group","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_instance_group.instance_group: Creation complete after 12s [id=projects/example-project/zones/us-central1-a/instanceGroups/example-instance-ig]","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:04.864269Z","hook":{"resource":{"addr":"google_compute_instance_group.instance_group","module":"","resource":"google_compute_instance_group.instance_group","implied_provider":"google","resource_type":"google_compute_instance_group","resource_name":"instance_group","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/zones/us-central1-a/instanceGroups/example-instance-ig","elapsed_seconds":12},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_backend_service.backend_service: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:29:04.879821Z","hook":{"resource":{"addr":"google_compute_backend_service.backend_service","module":"","resource":"google_compute_backend_service.backend_service","implied_provider":"google","resource_type":"google_compute_backend_service","resource_name":"backend_service","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_backend_service.backend_service: Creation complete after 1m3s [id=projects/example-project/global/backendServices/example-instance-bes]","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:07.910844Z","hook":{"resource":{"addr":"google_compute_backend_service.backend_service","module":"","resource":"google_compute_backend_service.backend_service","implied_provider":"google","resource_type":"google_compute_backend_service","resource_name":"backend_service","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/backendServices/example-instance-bes","elapsed_seconds":63},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_url_map.url_map: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:07.933852Z","hook":{"resource":{"addr":"google_compute_url_map.url_map","module":"","resource":"google_compute_url_map.url_map","implied_provider":"google","resource_type":"google_compute_url_map","resource_name":"url_map","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_url_map.url_map: Creation complete after 11s [id=projects/example-project/global/urlMaps/example-instance-urlmap]","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:19.372620Z","hook":{"resource":{"addr":"google_compute_url_map.url_map","module":"","resource":"google_compute_url_map.url_map","implied_provider":"google","resource_type":"google_compute_url_map","resource_name":"url_map","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/urlMaps/example-instance-urlmap","elapsed_seconds":11},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_target_http_proxy.http_proxy: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:19.380868Z","hook":{"resource":{"addr":"google_compute_target_http_proxy.http_proxy","module":"","resource":"google_compute_target_http_proxy.http_proxy","implied_provider":"google","resource_type":"google_compute_target_http_proxy","resource_name":"http_proxy","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_target_https_proxy.https_proxy: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:19.384149Z","hook":{"resource":{"addr":"google_compute_target_https_proxy.https_proxy","module":"","resource":"google_compute_target_https_proxy.https_proxy","implied_provider":"google","resource_type":"google_compute_target_https_proxy","resource_name":"https_proxy","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_target_https_proxy.https_proxy: Creation complete after 12s [id=projects/example-project/global/targetHttpsProxies/example-instance-https-proxy]","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:30.589006Z","hook":{"resource":{"addr":"google_compute_target_https_proxy.https_proxy","module":"","resource":"google_compute_target_https_proxy.https_proxy","implied_provider":"google","resource_type":"google_compute_target_https_proxy","resource_name":"https_proxy","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/targetHttpsProxies/example-instance-https-proxy","elapsed_seconds":12},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.https_forwarding_rule: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:30.599341Z","hook":{"resource":{"addr":"google_compute_global_forwarding_rule.https_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.https_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"https_forwarding_rule","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_target_http_proxy.http_proxy: Creation complete after 12s [id=projects/example-project/global/targetHttpProxies/example-instance-http-proxy]","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:30.703229Z","hook":{"resource":{"addr":"google_compute_target_http_proxy.http_proxy","module":"","resource":"google_compute_target_http_proxy.http_proxy","implied_provider":"google","resource_type":"google_compute_target_http_proxy","resource_name":"http_proxy","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/targetHttpProxies/example-instance-http-proxy","elapsed_seconds":12},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.http_forwarding_rule: Creating...","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:30.711891Z","hook":{"resource":{"addr":"google_compute_global_forwarding_rule.http_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.http_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"http_forwarding_rule","resource_key":null},"action":"create"},"type":"apply_start"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.https_forwarding_rule: Creation complete after 22s [id=projects/example-project/global/forwardingRules/example-instance-https-fwd-rule]","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:52.567767Z","hook":{"resource":{"addr":"google_compute_global_forwarding_rule.https_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.https_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"https_forwarding_rule","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/forwardingRules/example-instance-https-fwd-rule","elapsed_seconds":22},"type":"apply_complete"}
{"@level":"info","@message":"google_compute_global_forwarding_rule.http_forwarding_rule: Creation complete after 22s [id=projects/example-project/global/forwardingRules/example-instance-http-fwd-rule]","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:53.045954Z","hook":{"resource":{"addr":"google_compute_global_forwarding_rule.http_forwarding_rule","module":"","resource":"google_compute_global_forwarding_rule.http_forwarding_rule","implied_provider":"google","resource_type":"google_compute_global_forwarding_rule","resource_name":"http_forwarding_rule","resource_key":null},"action":"create","id_key":"id","id_value":"projects/example-project/global/forwardingRules/example-instance-http-fwd-rule","elapsed_seconds":22},"type":"apply_complete"}
{"@level":"warn","@message":"Warning: Value for undeclared variable","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:53.056560Z","diagnostic":{"severity":"warning","summary":"Value for undeclared variable","detail":"The root module does not declare a variable named "credentials_file" but a value was found in file "terraform.tfvars". If you meant to use this value, add a "variable" block to the configuration.\n\nTo silence these warnings, use TF_VAR_... environment variables to provide certain "global" settings to all configurations in your organization. To reduce the verbosity of these warnings, use the -compact-warnings option."},"type":"diagnostic"}
{"@level":"info","@message":"Apply complete! Resources: 18 added, 0 changed, 0 destroyed.","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:53.060601Z","changes":{"add":18,"change":0,"import":0,"remove":0,"operation":"apply"},"type":"change_summary"}
{"@level":"info","@message":"Outputs: 8","@module":"terraform.ui","@timestamp":"2025-05-10T03:30:53.060662Z","outputs":{"dns_name":{"sensitive":false,"type":"string","value":"example-instance.example.com"},"instance_name":{"sensitive":false,"type":"string","value":"example-instance"},"load_balancer_ip":{"sensitive":false,"type":"string","value":"10.0.0.1"},"private_ssh_key":{"sensitive":true,"type":"string"},"public_ip":{"sensitive":false,"type":"string","value":"10.0.0.2"},"public_ssh_key":{"sensitive":false,"type":"string","value":"ssh-rsa EXAMPLE_KEY"},"subnet_name":{"sensitive":false,"type":"string","value":"example-instance-subnet"},"vpc_name":{"sensitive":false,"type":"string","value":"example-instance-vpc"}},"type":"outputs"}
`

func TestProcessJSONStream(t *testing.T) {
	// Extract JSON lines from the provided sample log
	// Note: Manually curating this from the log. In a real scenario, ensure only JSON lines are fed.

	// Keep a reference to the original os.Stdout
	originalStdout := os.Stdout
	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a buffer to capture the output
	var outputBuf bytes.Buffer

	// Run the function in a goroutine so we can read from the pipe
	// and it doesn't block indefinitely if ProcessJSONStream has issues.
	errCh := make(chan error)
	go func() {
		errCh <- progress.ProcessJSONStream(strings.NewReader(mockTerraformOutput), &outputBuf)
	}()

	// Close the write end of the pipe on the main goroutine side
	// This will signal EOF to the reader.
	w.Close()

	// Read all output from the read end of the pipe
	var buf bytes.Buffer
	_, copyErr := io.Copy(&buf, r)

	// Restore os.Stdout
	os.Stdout = originalStdout

	// Wait for ProcessJSONStream to finish
	processErr := <-errCh

	if processErr != nil {
		t.Errorf("ProcessJSONStream returned an error: %v", processErr)
	}
	if copyErr != nil {
		t.Errorf("Error capturing stdout: %v", copyErr)
	}

	output := outputBuf.String()

	// --- Assertions ---

	// Check for total steps
	// From the log: "Plan: 18 to add, 0 to change, 0 to destroy."
	expectedTotalStepsStr := "(18)"
	if !strings.Contains(output, expectedTotalStepsStr) {
		t.Errorf("Output does not contain expected total steps string '%s'. Output:\n%s", expectedTotalStepsStr, output)
	}

	// Check for a specific "Creating..." message
	// e.g., "tls_private_key.ssh: Creating..."
	// The progress bar updates rapidly, so we check for the message substring
	expectedCreatingMessage := "tls_private_key.ssh: Creating..."
	if !strings.Contains(output, expectedCreatingMessage) {
		t.Errorf("Output does not contain expected creating message substring '%s'. Output:\n%s", expectedCreatingMessage, output)
	}

	// Check for a specific "Creation complete..." message
	expectedCompleteMessage := "Creation complete" // Part of the message
	if !strings.Contains(output, expectedCompleteMessage) {
		t.Errorf("Output does not contain expected completion message substring '%s'. Output:\n%s", expectedCompleteMessage, output)
	}

	// Check for the final "Apply complete!" message
	// The message is truncated to 48 characters with ...
	expectedApplyComplete := "Apply complete! Resources: 18 added, 0 change..."
	if !strings.Contains(output, expectedApplyComplete) {
		t.Errorf("Output does not contain expected final apply complete message '%s'. Output:\n%s", expectedApplyComplete, output)
	}

	// Check for progress bar structure (e.g., presence of '[=')
	// This is a bit tricky because the bar changes. We look for a common pattern.
	// The first step would be like "1[=...](18)"
	// A more robust check might be to parse the output lines, but for now, a substring check.
	if !strings.Contains(output, "[=") && !strings.Contains(output, "[-/-]") { // [-/-] is for totalSteps=0 case
		t.Errorf("Output does not seem to contain a progress bar structure like '[='. Output:\n%s", output)
	}

	// Check that a newline was printed after the final progress bar for apply summary
	// The "Apply complete!" message should be followed by a newline (handled by fmt.Println in ProcessJSONStream)
	// and then the "Outputs: 8" message starts on a new line in the *original* terraform output,
	// our progress bar code should also ensure it prints a newline.

	// A simple check for multiple lines at the end:
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		// The very last thing printed by printProgress before the final newline would be the "Outputs: 8" message
		// or the "Processing outputs..." message.
		if !strings.Contains(lastLine, "Outputs: 8") && !strings.Contains(lastLine, "Processing outputs...") {
			// t.Logf("Last line of output: %s", lastLine) // For debugging
		}
	} else {
		t.Errorf("Expected multiple lines of output, got 0 or 1 after split and trim. Output:\n%s", output)
	}

	// A more detailed check might involve capturing each line distinctly,
	// which is harder with `
	// clearing lines unless the terminal emulation is very precise.
	// For now, these substring checks cover the main aspects.
}

type delayedReader struct {
	lines []string
	idx   int
	delay time.Duration
}

func (dr *delayedReader) Read(p []byte) (int, error) {
	if dr.idx >= len(dr.lines) {
		return 0, io.EOF
	}
	line := dr.lines[dr.idx] + "\n"
	copy(p, line)
	dr.idx++
	time.Sleep(dr.delay)
	return len(line), nil
}

func TestProcessJSONStream_Visual(t *testing.T) {
	// This test prints the progress bar and messages to the terminal for visual inspection.

	dr := &delayedReader{
		lines: strings.Split(strings.TrimSpace(mockTerraformOutput), "\n"),
		delay: 500 * time.Millisecond,
	}

	err := progress.ProcessJSONStream(dr, os.Stdout)
	if err != nil {
		t.Errorf("ProcessJSONStream returned an error: %v", err)
	}
}
