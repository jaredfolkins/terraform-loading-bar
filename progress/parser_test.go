package progress_test

import (
	"context"
	"fmt"
	"io"
	"strconv"
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

var mockTerraformDestroyOutput = `
{"@level":"info","@message":"Terraform 1.11.4","@module":"terraform.ui","@timestamp":"2025-05-10T10:00:57.945980-07:00","terraform":"1.11.4","type":"version","ui":"1.2"}
{"@level":"info","@message":"tls_private_key.ssh: Refreshing state... [id=abcdef1234567890abcdef1234567890abcdef12]","@module":"terraform.ui","@timestamp":"2025-05-10T10:01:00.215590-07:00","hook":{"resource":{"addr":"tls_private_key.ssh","module":"","resource":"tls_private_key.ssh","implied_provider":"tls","resource_type":"tls_private_key","resource_name":"ssh","resource_key":null},"id_key":"id","id_value":"abcdef1234567890abcdef1234567890abcdef12"},"type":"refresh_start"}
{"@level":"info","@message":"tls_private_key.ssh: Refresh complete [id=abcdef1234567890abcdef1234567890abcdef12]","@module":"terraform.ui","@timestamp":"2025-05-10T10:01:00.217654-07:00","hook":{"resource":{"addr":"tls_private_key.ssh","module":"","resource":"tls_private_key.ssh","implied_provider":"tls","resource_type":"tls_private_key","resource_name":"ssh","resource_key":null},"id_key":"id","id_value":"abcdef1234567890abcdef1234567890abcdef12"},"type":"refresh_complete"}
{"@level":"info","@message":"Plan: 0 to add, 0 to change, 18 to destroy.","@module":"terraform.ui","@timestamp":"2025-05-10T10:01:04.960408-07:00","changes":{"add":0,"change":0,"import":0,"remove":18,"operation":"plan"},"type":"change_summary"}
{"@level":"info","@message":"google_dns_record_set.dns_record: Destroying... [id=projects/example-project/managedZones/example.com/rrsets/example-instance.example.com./A]","@module":"terraform.ui","@timestamp":"2025-05-10T10:01:05.865228-07:00","hook":{"resource":{"addr":"google_dns_record_set.dns_record","module":"","resource":"google_dns_record_set.dns_record","implied_provider":"google","resource_type":"google_dns_record_set","resource_name":"dns_record","resource_key":null},"action":"delete","id_key":"id","id_value":"projects/example-project/managedZones/example.com/rrsets/example-instance.example.com./A"},"type":"apply_start"}
{"@level":"info","@message":"google_dns_record_set.dns_record: Destruction complete after 2s","@module":"terraform.ui","@timestamp":"2025-05-10T10:01:08.422318-07:00","hook":{"resource":{"addr":"google_dns_record_set.dns_record","module":"","resource":"google_dns_record_set.dns_record","implied_provider":"google","resource_type":"google_dns_record_set","resource_name":"dns_record","resource_key":null},"action":"delete","elapsed_seconds":2},"type":"apply_complete"}
{"@level":"info","@message":"Destroy complete! Resources: 18 destroyed.","@module":"terraform.ui","@timestamp":"2025-05-10T10:04:06.935983-07:00","changes":{"add":0,"change":0,"import":0,"remove":18,"operation":"destroy"},"type":"change_summary"}
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

	// Check for the final "Apply complete!" message
	expectedApplyComplete := "Apply complete! Resources: 18 added, 0 change..."
	foundApplyComplete := false
	for _, line := range lines {
		if strings.Contains(line, expectedApplyComplete) {
			foundApplyComplete = true
			break
		}
	}
	if !foundApplyComplete {
		t.Errorf("Output does not contain expected final apply complete message '%s'. Output:\n%s", expectedApplyComplete, strings.Join(lines, "\n"))
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

func TestProcessJSONStream_Visual(t *testing.T) {
	// This test prints the progress bar and messages to the terminal for visual inspection.
	reader := strings.NewReader(mockTerraformOutput)
	handler := progress.NewProgressHandler(reader)

	for {
		line, err := handler.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Errorf("Unexpected error: %v", err)
			return
		}
		if line != "" {
			fmt.Println(line)
		}
	}
}

func TestGetProgressOutput(t *testing.T) {
	// Create a reader with the mock Terraform output
	reader := strings.NewReader(mockTerraformOutput)

	// Get the progress output as a string
	output, err := progress.GetProgressOutput(reader)
	if err != nil {
		t.Errorf("GetProgressOutput returned an error: %v", err)
	}

	// Split the output into lines for easier testing
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// --- Assertions ---

	// Check for total steps
	// From the log: "Plan: 18 to add, 0 to change, 0 to destroy."
	expectedTotalStepsStr := "(36)"
	foundTotalSteps := false
	for _, line := range lines {
		if strings.Contains(line, expectedTotalStepsStr) {
			foundTotalSteps = true
			break
		}
	}
	if !foundTotalSteps {
		t.Errorf("Output does not contain expected total steps string '%s'. Output:\n%s", expectedTotalStepsStr, output)
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
		t.Errorf("Output does not contain expected creating message '%s'. Output:\n%s", expectedCreatingMessage, output)
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
		t.Errorf("Output does not contain expected completion message '%s'. Output:\n%s", expectedCompleteMessage, output)
	}

	// Check for the final "Apply complete!" message
	expectedApplyComplete := "Apply complete! Resources: 18 added, 0 change..."
	foundApplyComplete := false
	for _, line := range lines {
		if strings.Contains(line, expectedApplyComplete) {
			foundApplyComplete = true
			break
		}
	}
	if !foundApplyComplete {
		t.Errorf("Output does not contain expected final apply complete message '%s'. Output:\n%s", expectedApplyComplete, output)
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
		t.Errorf("Output does not seem to contain a progress bar structure like '[='. Output:\n%s", output)
	}

	// Check that the output ends with a newline
	if !strings.HasSuffix(output, "\n") {
		t.Errorf("Output does not end with a newline. Output:\n%s", output)
	}

	// Check that the output contains multiple lines
	if len(lines) < 2 {
		t.Errorf("Expected multiple lines of output, got %d. Output:\n%s", len(lines), output)
	}

	// Check that the last line contains either "Outputs: 8" or "Processing outputs..."
	lastLine := lines[len(lines)-1]
	if !strings.Contains(lastLine, "Outputs: 8") && !strings.Contains(lastLine, "Processing outputs...") {
		t.Errorf("Last line does not contain expected output message. Last line: %s", lastLine)
	}
}

func TestGetProgressOutputWithPrint(t *testing.T) {
	// Create a reader with the mock Terraform output
	reader := strings.NewReader(mockTerraformOutput)

	// Get the progress output as a string
	output, err := progress.GetProgressOutput(reader)
	if err != nil {
		t.Errorf("GetProgressOutput returned an error: %v", err)
	}

	// Print the output with a header
	fmt.Println("=== Terraform Progress Output ===")
	fmt.Print(output)
	fmt.Println("=== End of Progress Output ===")

	// Verify the output is not empty
	if len(output) == 0 {
		t.Error("Output is empty")
	}

	// Verify the output contains expected content
	expectedContent := []string{
		"Planning...",
		"Creating...",
		"Creation complete",
		"Apply complete",
	}

	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("Output does not contain expected content: %s", content)
		}
	}

	// Print the output line by line for detailed inspection
	fmt.Println("\n=== Detailed Output Inspection ===")
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if line != "" { // Skip empty lines
			fmt.Printf("Line %d: %s\n", i+1, line)
		}
	}
	fmt.Println("=== End of Detailed Inspection ===")
}

func TestProgressHandler(t *testing.T) {
	// Create a reader with the mock Terraform output
	reader := strings.NewReader(mockTerraformOutput)

	// Create a progress handler
	handler := progress.NewProgressHandler(reader)

	// Read all lines
	var lines []string
	for {
		line, err := handler.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Errorf("Unexpected error: %v", err)
			return
		}
		if line != "" {
			lines = append(lines, line)
		}
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

	// Check for the final "Apply complete!" message
	expectedApplyComplete := "Apply complete! Resources: 18 added, 0 change..."
	foundApplyComplete := false
	for _, line := range lines {
		if strings.Contains(line, expectedApplyComplete) {
			foundApplyComplete = true
			break
		}
	}
	if !foundApplyComplete {
		t.Errorf("Output does not contain expected final apply complete message '%s'. Output:\n%s", expectedApplyComplete, strings.Join(lines, "\n"))
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

// seekableReader is a test helper that wraps a string reader with seeking capability
type seekableReader struct {
	content string
	pos     int64
}

func newSeekableReader(content string) *seekableReader {
	return &seekableReader{content: content}
}

func (r *seekableReader) Read(p []byte) (int, error) {
	if r.pos >= int64(len(r.content)) {
		return 0, io.EOF
	}
	n := copy(p, r.content[r.pos:])
	r.pos += int64(n)
	if r.pos >= int64(len(r.content)) {
		return n, io.EOF
	}
	return n, nil
}

func (r *seekableReader) Seek(offset int64, whence int) (int64, error) {
	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = r.pos + offset
	case io.SeekEnd:
		newPos = int64(len(r.content)) + offset
	default:
		return 0, fmt.Errorf("invalid whence")
	}
	if newPos < 0 {
		return 0, fmt.Errorf("negative position")
	}
	if newPos > int64(len(r.content)) {
		newPos = int64(len(r.content))
	}
	r.pos = newPos
	return newPos, nil
}

func TestProgressHandler_Streaming(t *testing.T) {
	// Create a seekable reader with the mock Terraform output
	reader := newSeekableReader(mockTerraformOutput)

	// Create a progress handler
	handler := progress.NewProgressHandler(reader)

	// Track the lines we receive
	var lines []string
	var errors []error

	// Read all lines with context to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("Test timed out after 5 seconds")
		default:
			line, err := handler.ReadLine()
			if err != nil {
				if err == io.EOF {
					goto done
				}
				errors = append(errors, err)
				goto done
			}
			if line != "" {
				lines = append(lines, line)
			}
		}
	}
done:

	// Verify we got some output
	if len(lines) == 0 {
		t.Error("Expected non-empty output")
		return
	}

	// Verify no errors occurred
	if len(errors) > 0 {
		t.Errorf("Unexpected errors: %v", errors)
		return
	}

	// Test the output format and content
	t.Run("Output Format", func(t *testing.T) {
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

		// Check for the final "Apply complete!" message
		expectedApplyComplete := "Apply complete! Resources: 18 added, 0 change..."
		foundApplyComplete := false
		for _, line := range lines {
			if strings.Contains(line, expectedApplyComplete) {
				foundApplyComplete = true
				break
			}
		}
		if !foundApplyComplete {
			t.Errorf("Output does not contain expected final apply complete message '%s'. Output:\n%s", expectedApplyComplete, strings.Join(lines, "\n"))
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
	})

	// Test the sequence of messages
	t.Run("Message Sequence", func(t *testing.T) {
		// Verify planning phase comes before apply phase
		planningIndex := -1
		applyIndex := -1
		for i, line := range lines {
			if strings.Contains(line, "PLANNING") {
				planningIndex = i
			}
			if strings.Contains(line, "Creating...") {
				applyIndex = i
				break
			}
		}
		if planningIndex == -1 {
			t.Error("Planning phase not found in output")
		}
		if applyIndex == -1 {
			t.Error("Apply phase not found in output")
		}
		if planningIndex > applyIndex {
			t.Error("Planning phase should come before apply phase")
		}

		// Verify progress increases
		lastProgress := -1
		for _, line := range lines {
			if strings.Contains(line, "[=") {
				// Extract current step from line like "(1)[=...](18)"
				parts := strings.Split(line, "[")
				if len(parts) > 0 {
					stepStr := strings.Trim(parts[0], "()")
					if step, err := strconv.Atoi(stepStr); err == nil {
						if lastProgress != -1 && step < lastProgress {
							t.Errorf("Progress decreased from %d to %d", lastProgress, step)
						}
						lastProgress = step
					}
				}
			}
		}
	})

	// Test error handling
	t.Run("Error Handling", func(t *testing.T) {
		// Create a reader that will return an error
		errorReader := &errorReader{err: fmt.Errorf("test error")}
		handler := progress.NewProgressHandler(errorReader)

		// Read the initial progress bar (should always be present)
		_, _ = handler.ReadLine()

		// Read a line and expect an error, with timeout to avoid race
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic occurred during error handling: %v", r)
			}
		}()

		errCh := make(chan error, 1)
		go func() {
			_, err := handler.ReadLine()
			errCh <- err
		}()

		select {
		case err := <-errCh:
			if err == nil {
				t.Error("Expected error from errorReader")
				return
			}
			if err.Error() != "test error" {
				t.Errorf("Expected error 'test error', got '%v'", err)
			}
		case <-time.After(500 * time.Millisecond):
			t.Error("Timed out waiting for error from errorReader")
		}
	})
}

// errorReader is a test helper that always returns an error
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (int, error) {
	return 0, r.err
}

func TestProgressBarTransitions(t *testing.T) {
	// Create a reader with the mock Terraform output
	reader := strings.NewReader(mockTerraformOutput)
	handler := progress.NewProgressHandler(reader)

	// Track the lines we receive
	var lines []string
	var planningLines []string
	var applyLines []string

	// Read all lines
	for {
		line, err := handler.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Errorf("Unexpected error: %v", err)
			return
		}
		if line != "" {
			lines = append(lines, line)

			// Track planning vs apply phase lines
			if strings.Contains(line, "[      PLANNING      ]") {
				planningLines = append(planningLines, line)
			} else if strings.Contains(line, "[=") {
				applyLines = append(applyLines, line)
			}
		}
	}

	// Verify we got some output
	if len(lines) == 0 {
		t.Error("Expected non-empty output")
		return
	}

	// Test 1: Verify we transition from planning to apply phase
	if len(applyLines) == 0 {
		t.Error("Progress bar never transitioned from PLANNING to showing actual progress")
	}

	// Test 2: Verify we don't get stuck in planning
	if len(planningLines) > 20 { // Arbitrary threshold, adjust if needed
		t.Errorf("Progress bar appears to be stuck in planning phase. Got %d planning lines", len(planningLines))
	}

	// Test 3: Verify the transition happens at the right time
	foundApplyStart := false
	for i, line := range lines {
		if strings.Contains(line, "Creating...") {
			foundApplyStart = true
		}
		if foundApplyStart && strings.Contains(line, "[      PLANNING      ]") {
			t.Errorf("Found PLANNING after apply started at line %d: %s", i, line)
		}
	}

	// Test 4: Verify progress increases after transition
	var lastProgress int = -1
	for _, line := range applyLines {
		if strings.Contains(line, "[=") {
			// Extract current step from line like "(1)[=...](18)"
			parts := strings.Split(line, "[")
			if len(parts) > 0 {
				stepStr := strings.Trim(parts[0], "()")
				if step, err := strconv.Atoi(stepStr); err == nil {
					if lastProgress != -1 && step < lastProgress {
						t.Errorf("Progress decreased from %d to %d", lastProgress, step)
					}
					lastProgress = step
				}
			}
		}
	}
}

func TestProgressBarPhaseDetection(t *testing.T) {
	// Create a reader with the mock Terraform output
	reader := strings.NewReader(mockTerraformOutput)
	handler := progress.NewProgressHandler(reader)

	// Track the phases we see
	var phases []string
	var currentPhase string = "initial"

	// Read all lines
	for {
		line, err := handler.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Errorf("Unexpected error: %v", err)
			return
		}
		if line != "" {
			// Detect phase changes
			if strings.Contains(line, "[      PLANNING      ]") {
				if currentPhase != "planning" {
					currentPhase = "planning"
					phases = append(phases, currentPhase)
				}
			} else if strings.Contains(line, "[=") && !strings.Contains(line, "PLANNING") {
				if currentPhase != "apply" {
					currentPhase = "apply"
					phases = append(phases, currentPhase)
				}
			}
		}
	}

	// Verify we saw both phases
	if len(phases) < 2 {
		t.Errorf("Expected at least 2 phases (planning and apply), got %d: %v", len(phases), phases)
	}

	// Verify the correct order of phases
	if phases[0] != "planning" {
		t.Errorf("Expected first phase to be 'planning', got '%s'", phases[0])
	}
	if phases[1] != "apply" {
		t.Errorf("Expected second phase to be 'apply', got '%s'", phases[1])
	}

	// Verify we don't go back to planning after apply starts
	for i := 2; i < len(phases); i++ {
		if phases[i] == "planning" {
			t.Errorf("Found planning phase after apply started at index %d", i)
		}
	}
}
