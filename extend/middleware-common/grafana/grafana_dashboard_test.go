package grafana

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestGetDashboardName(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	dashboardName := GetDashboardName("MySQL_InnoDB_Details.json")
	g.Expect(dashboardName).Should(gomega.Equal("mysql-innodb-details"))
}
