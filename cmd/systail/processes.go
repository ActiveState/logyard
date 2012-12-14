// TODO: load this into doozer?
package main

// PROCESS is a map of process name to log file. if log file is empty,
// assume /s/logs/${name}.log.
var PROCESSES = map[string]string{
	"stackato-lxc": "",
	"kato":         "",
	"supervisord":  "",
	"dmesg":        "/var/log/dmesg",
	"kernel":       "/var/log/kern.log",
	"auth":         "/var/log/auth.log",
	"dpkg":         "/var/log/dpkg.log",
	"nginx_error":  "/var/log/nginx/error.log",

	// Autogenerated from the following command:
	// for name in ` grep name ~/as/stackato/etc/kato/processes.yml | \
	//    cut -d " " -f 6`; do echo "\"$name\","; done
	"doozerd":            "",
	"cloudevents":        "",
	"postgresql":         "",
	"avahi_daemon":       "",
	"avahi_announcer":    "",
	"nginx":              "",
	"cc_nginx":           "",
	"cc_nginx_error":     "",
	"nats_server":        "",
	"redis_server":       "",
	"applog_redis":       "",
	"logyard":            "",
	"apptail":            "",
	"systail":            "",
	"app_store":          "",
	"cloud_controller":   "",
	"health_manager":     "",
	"prealloc":           "",
	"stager":             "",
	"router":             "",
	"dea":                "",
	"mysql":              "",
	"filesystem_node":    "",
	"filesystem_gateway": "",
	"memcached_node":     "",
	"memcached_gateway":  "",
	"mongodb_node":       "",
	"mongodb_gateway":    "",
	"mysql_node":         "",
	"mysql_gateway":      "",
	"postgresql_node":    "",
	"postgresql_gateway": "",
	"rabbit_node":        "",
	"rabbit_gateway":     "",
	"redis_node":         "",
	"redis_gateway":      "",
	"router2g":           "",
}
