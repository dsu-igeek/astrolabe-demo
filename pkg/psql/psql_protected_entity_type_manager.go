package psql

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"github.com/vmware-tanzu/astrolabe/pkg/localsnap"
	v1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	"github.com/zalando/postgres-operator/pkg/util/k8sutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type PSQLProtectedEntityTypeManager struct {
	KubeClient   k8sutil.KubernetesClient
	snapshotsDir string
	s3Config     astrolabe.S3Config
	logger       logrus.FieldLogger
	internalRepo localsnap.LocalSnapshotRepo
}

const (
	KubeConfigKey   = "kubeconfig"
	SnapshotsDirKey = "snapshotsDir"
)

func NewPSQLProtectedEntityTypeManager(params map[string]interface{}, s3Config astrolabe.S3Config,
	logger logrus.FieldLogger) (astrolabe.ProtectedEntityTypeManager, error) {
	kubeconfgPathObj := params["KubeConfigKey"]
	kubeconfigPath := ""

	if kubeconfgPathObj != nil {
		kubeconfigPath = kubeconfgPathObj.(string)
	} else {
		kubeconfigPath = os.Getenv("KUBECONFIG")
	}
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)

	if err != nil {
		return PSQLProtectedEntityTypeManager{}, errors.Wrap(err, "could not create restConfig")
	}
	kubeClient, err := k8sutil.NewFromConfig(restConfig)

	if err != nil {
		return PSQLProtectedEntityTypeManager{}, errors.Wrap(err, "could not create kubeClient")
	}

	snapshotsDir, hasSnapshotsDir := params[SnapshotsDirKey].(string)
	if !hasSnapshotsDir {
		return PSQLProtectedEntityTypeManager{}, errors.New("no " + SnapshotsDirKey + " param found")
	}

	localSnapshotRepo, err := localsnap.NewLocalSnapshotRepo(Typename, snapshotsDir)
	if err != nil {
		return PSQLProtectedEntityTypeManager{}, err
	}
	returnPETM := PSQLProtectedEntityTypeManager{
		KubeClient:   kubeClient,
		snapshotsDir: snapshotsDir,
		logger:       logger,
		s3Config:     s3Config,
		internalRepo: localSnapshotRepo,
	}

	return returnPETM, nil
}

const Typename = "psql"
func (this PSQLProtectedEntityTypeManager) GetTypeName() string {
	return Typename
}

func (this PSQLProtectedEntityTypeManager) GetProtectedEntity(ctx context.Context, id astrolabe.ProtectedEntityID) (astrolabe.ProtectedEntity, error) {
	_, err := this.getPostgresqlForPEID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve instance info")
	}
	return NewPSQLProtectedEntity(id, &this), nil

}

func (this PSQLProtectedEntityTypeManager) getPostgresqlForPEID(ctx context.Context, id astrolabe.ProtectedEntityID) (v1.Postgresql, error) {
	list, err := this.listAllPSQLs(ctx)
	if err != nil {
		return v1.Postgresql{}, errors.Wrap(err, "could not retrieve psqls")
	}
	for _, curPSQL := range list {
		if string(curPSQL.UID) == id.GetID() {
			return curPSQL, nil
		}
	}
	return v1.Postgresql{}, errors.New("Not found")
}

func (this PSQLProtectedEntityTypeManager) listAllPSQLs(ctx context.Context) ([]v1.Postgresql, error) {
	var options metav1.ListOptions
	namespaces, err := this.KubeClient.Namespaces().List(ctx, options)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve namespaces")
	}
	var list []v1.Postgresql
	for _, curNamespace := range namespaces.Items {
		curList, err := this.KubeClient.AcidV1ClientSet.AcidV1().Postgresqls(curNamespace.Name).List(ctx, options)
		if err != nil {
			return nil, errors.Wrap(err, "could not retrieve postgres instances")
		}
		list = append(list, curList.Items...)
	}
	return list, nil
}

func (this PSQLProtectedEntityTypeManager) GetProtectedEntities(ctx context.Context) ([]astrolabe.ProtectedEntityID, error) {
	list, err := this.listAllPSQLs(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve postgres instances")
	}

	var returnIDs []astrolabe.ProtectedEntityID
	for _, curPSQL := range list {
		fmt.Printf("%v\n", curPSQL)
		id := astrolabe.NewProtectedEntityID(this.GetTypeName(), string(curPSQL.UID))
		returnIDs = append(returnIDs, id)
	}
	return returnIDs, nil
}

func (this PSQLProtectedEntityTypeManager) Copy(ctx context.Context, pe astrolabe.ProtectedEntity, params map[string]map[string]interface {},
    options astrolabe.CopyCreateOptions) (astrolabe.ProtectedEntity, error) {
	panic("implement me")
}

func (this PSQLProtectedEntityTypeManager) CopyFromInfo(ctx context.Context, info astrolabe.ProtectedEntityInfo, params map[string]map[string]interface {},
    options astrolabe.CopyCreateOptions) (astrolabe.ProtectedEntity, error) {
	panic("implement me")
}

func (this PSQLProtectedEntityTypeManager) Delete(ctx context.Context, id astrolabe.ProtectedEntityID) error {
	panic("implement me")
}
