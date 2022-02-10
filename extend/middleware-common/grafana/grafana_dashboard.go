package grafana

import (
	"context"
	"fmt"
	"gomod.alauda.cn/ait-apis/grafanadashboard/v1beta1"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

var defaultJsonFolder = "/etc/config/grafana_dashboards"
var defaultDashboardFolder = "Middleware"

type DashboardManager struct {
	client.Client
	dashboardJsonFolder string
	dashboardFolder     string
	namespace           string
}

func NewGrafanaDashboardMgr(Client client.Client) *DashboardManager {
	return &DashboardManager{
		Client:              Client,
		dashboardJsonFolder: defaultJsonFolder,
		dashboardFolder:     defaultDashboardFolder,
		namespace:           "operators",
	}
}

func (mgr *DashboardManager) Namespace(namespace string) *DashboardManager {
	mgr.namespace = namespace
	return mgr
}

func (mgr *DashboardManager) DashboardJsonFolder(dashboardJsonFolder string) *DashboardManager {
	mgr.dashboardJsonFolder = dashboardJsonFolder
	return mgr
}

func (mgr *DashboardManager) DashboardFolder(dashboardFolder string) *DashboardManager {
	mgr.dashboardFolder = dashboardFolder
	return mgr
}

func (mgr *DashboardManager) CreateGrafanaDashboard() error {
	files, err := ioutil.ReadDir(mgr.dashboardJsonFolder)
	if err != nil {
		return err
	}
	for _, file := range files {
		err = mgr.ProcessFile(mgr.Client,
			fmt.Sprintf("%s//%s", mgr.dashboardJsonFolder, file.Name()), file.Name(), mgr.dashboardFolder)
		if err != nil {
			fmt.Print(err)
		}
	}
	return nil
}

func (mgr *DashboardManager) ProcessFile(Client client.Client, path string, file string, dashboardFolder string) error {
	grafanaDashboard, err := mgr.GetDashboard(path, file, dashboardFolder)
	if err != nil {
		return err
	}
	existed := &v1beta1.GrafanaDashboard{}
	err = Client.Get(context.Background(), types.NamespacedName{
		Name:      GetDashboardName(file),
		Namespace: mgr.namespace,
	}, existed)
	if err != nil {
		if errors.IsNotFound(err) {
			return Client.Create(context.Background(), grafanaDashboard)
		}
		return err
	}
	if existed.Spec.Json != grafanaDashboard.Spec.Json ||
		existed.Spec.Folder != grafanaDashboard.Spec.Folder {
		existed.Spec.Json = grafanaDashboard.Spec.Json
		existed.Spec.Folder = grafanaDashboard.Spec.Folder
		return Client.Update(context.Background(), existed)
	}
	return nil
}

func (mgr *DashboardManager) GetDashboard(path string, file string, dashboardFolder string) (*v1beta1.GrafanaDashboard, error) {
	grafanaJson, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return &v1beta1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: mgr.namespace,
			Name:      GetDashboardName(file),
		},
		Spec: v1beta1.GrafanaDashboardSpec{
			Folder: dashboardFolder,
			Json:   string(grafanaJson),
		},
	}, nil
}

func GetDashboardName(fileName string) string {
	result := fileName[:len(fileName)-5]
	result = strings.ReplaceAll(result, "_", "-")
	return strings.ToLower(result)
}
