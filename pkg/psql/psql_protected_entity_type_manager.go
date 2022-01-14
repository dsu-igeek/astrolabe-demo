package psql

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"github.com/vmware-tanzu/astrolabe/pkg/localsnap"
	v1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	"github.com/zalando/postgres-operator/pkg/util/k8sutil"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
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

const (
	timeout     = 3 * time.Minute
	poll        = 6 * time.Second
	podWaitTime = 10 * time.Second
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
		if curPSQL.Name == id.GetID() {
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

func (this PSQLProtectedEntityTypeManager) Copy(ctx context.Context, pe astrolabe.ProtectedEntity, params map[string]map[string]interface{},
	options astrolabe.CopyCreateOptions) (astrolabe.ProtectedEntity, error) {
	panic("implement me")

}

func (this PSQLProtectedEntityTypeManager) CopyFromInfo(ctx context.Context, info astrolabe.ProtectedEntityInfo, params map[string]map[string]interface{},
	options astrolabe.CopyCreateOptions) (astrolabe.ProtectedEntity, error) {
	reader, err := this.internalRepo.GetDataReaderForSnapshot(info.GetID())
	sourcePSQL, err := this.getPostgresqlForPEID(context.TODO(), info.GetID())
	destinationPSQL := sourcePSQL

	y, m, d := time.Now().Date()
	date := fmt.Sprintf("%d%d%d", y, m, d)
	destinationPSQL.Name = fmt.Sprintf("%s-%s-%s", sourcePSQL.Spec.TeamID, strings.Split(sourcePSQL.Name, "-")[1], date)
	destinationPSQL.ResourceVersion = ""
	if destinationPSQL.Spec.ResourceLimits.CPU == "" {
		destinationPSQL.Spec.ResourceLimits.CPU = "1"
		destinationPSQL.Spec.ResourceLimits.Memory = "500Mi"
	}

	if destinationPSQL.Spec.ResourceRequests.CPU == "" {
		destinationPSQL.Spec.ResourceRequests.CPU = "100m"
		destinationPSQL.Spec.ResourceRequests.Memory = "100Mi"
	}

	destinationpghost := destinationPSQL.Name
	namespace := destinationPSQL.Namespace

	sourcepghost := sourcePSQL.Name
	sourcePostgresSecret, err := this.KubeClient.Secrets(namespace).Get(ctx, "postgres."+sourcepghost+".credentials", metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve secret from source postgres")
	}

	destinationPostgresSecret := sourcePostgresSecret
	destinationPostgresSecret.Name = "postgres." + destinationpghost + ".credentials"
	destinationPostgresSecret.ResourceVersion = ""

	_, err = this.KubeClient.Secrets(namespace).Create(context.TODO(), destinationPostgresSecret, metav1.CreateOptions{})
	if err != nil && !k8serr.IsAlreadyExists(err) {
		return nil, err
	}

	for user := range destinationPSQL.Spec.Users {
		secretName := user + "." + sourcepghost + ".credentials"
		sourceSecret, err := this.KubeClient.Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			if k8serr.IsNotFound(err) {
				continue
			} else if !k8serr.IsAlreadyExists(err) {
				return nil, errors.Wrap(err, "error while retrieving secrets from source cluster")
			}
		}
		destinationSecret := sourceSecret
		destinationSecret.Name = user + "." + destinationpghost + ".credentials"
		destinationSecret.ResourceVersion = ""
		_, err = this.KubeClient.Secrets(namespace).Create(context.TODO(), destinationSecret, metav1.CreateOptions{})
		if err != nil && !k8serr.IsAlreadyExists(err) {
			return nil, errors.Wrap(err, "error creating secrets")
		}

	}

	_, err = this.KubeClient.AcidV1ClientSet.AcidV1().Postgresqls(sourcePSQL.Namespace).Create(context.TODO(), &destinationPSQL, metav1.CreateOptions{})
	if err != nil && !k8serr.IsAlreadyExists(err) {
		return nil, err
	}
	podName := destinationPSQL.Name + "-0"

	pod, err := this.KubeClient.Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil && !k8serr.IsAlreadyExists(err) {
		if k8serr.IsNotFound(err) {
			for start := time.Now(); time.Since(start) < timeout; time.Sleep(poll) {
				demoPod, err := this.KubeClient.Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
				if demoPod != nil {
					fmt.Println("Pod Status", demoPod.Status.Phase)
				}
				if !k8serr.IsNotFound(err) && demoPod.Status.Phase == corev1.PodRunning {
					fmt.Println("Pod found")
					// Wait before exec
					time.Sleep(podWaitTime)
					break
				}
			}
		} else {
			return nil, err
		}
	}

	postgresUser := string(sourcePostgresSecret.Data["username"])

	var dbname string
	for keyname := range destinationPSQL.Spec.Databases {
		dbname = keyname
	}

	fmt.Println("dbname is ", dbname)

	execPod := exec.Command("kubectl", "exec", "-i", pod.Name, "-n", pod.Namespace, "--", "psql", "-U", postgresUser, "-d", dbname)
	execPod.Stdin = reader
	res, err := execPod.Output()
	fmt.Println(string(res))
	if err != nil {
		return nil, err
	}

	checkSecret, err := this.KubeClient.Secrets(namespace).Get(context.TODO(), destinationPostgresSecret.Name, metav1.GetOptions{})
	if err != nil && !k8serr.IsAlreadyExists(err) {
		return nil, err
	}

	if string(checkSecret.Data["password"]) == string(sourcePostgresSecret.Data["password"]) {
		fmt.Println("Secrets are in sync")
	}

	destinationID := astrolabe.NewProtectedEntityID(this.GetTypeName(), string(destinationPSQL.UID))
	newProtectedEntity, err := this.GetProtectedEntity(context.TODO(), destinationID)
	if err != nil {
		return nil, err
	}

	return newProtectedEntity, nil

}

func (this PSQLProtectedEntityTypeManager) Delete(ctx context.Context, id astrolabe.ProtectedEntityID) error {
	panic("implement me")
}
