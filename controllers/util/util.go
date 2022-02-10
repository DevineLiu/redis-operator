package util

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/DevineLiu/redis-operator/extend/middleware-common/grafana"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ParseRedisMemConf(p string) (string, error) {
	var mul int64 = 1
	u := strings.ToLower(p)
	digits := u

	if strings.HasSuffix(u, "k") {
		digits = u[:len(u)-len("k")]
		mul = 1000
	} else if strings.HasSuffix(u, "kb") {
		digits = u[:len(u)-len("kb")]
		mul = 1024
	} else if strings.HasSuffix(u, "m") {
		digits = u[:len(u)-len("m")]
		mul = 1000 * 1000
	} else if strings.HasSuffix(u, "mb") {
		digits = u[:len(u)-len("mb")]
		mul = 1024 * 1024
	} else if strings.HasSuffix(u, "g") {
		digits = u[:len(u)-len("g")]
		mul = 1000 * 1000 * 1000
	} else if strings.HasSuffix(u, "gb") {
		digits = u[:len(u)-len("gb")]
		mul = 1024 * 1024 * 1024
	} else if strings.HasSuffix(u, "b") {
		digits = u[:len(u)-len("b")]
		mul = 1
	}

	val, err := strconv.ParseInt(digits, 10, 64)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(val*mul, 10), nil
}

var dashboardFolder = "/etc/config/grafana_dashboards"

func CreateGrafanaDashboard(Client client.Client) {
	if os.Getenv("DEBUG") == "true" {
		dashboardFolder = "./grafana_dashboard"
	}
	dashboardManager := grafana.NewGrafanaDashboardMgr(Client).DashboardJsonFolder(dashboardFolder).
		DashboardFolder("Redis").Namespace("operators")
	err := dashboardManager.CreateGrafanaDashboard()
	if err != nil {
		fmt.Println("create grafana err:", err)
		return
	}

	fmt.Println("add grafana dashboard finish")
}
