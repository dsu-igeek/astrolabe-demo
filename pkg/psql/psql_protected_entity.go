package psql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"io"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os/exec"
)

type PSQLProtectedEntity struct {
	id        astrolabe.ProtectedEntityID
	petm      *PSQLProtectedEntityTypeManager
	namespace string
}

func NewPSQLProtectedEntity(id astrolabe.ProtectedEntityID, petm *PSQLProtectedEntityTypeManager) PSQLProtectedEntity {
	return PSQLProtectedEntity{
		id:   id,
		petm: petm,
	}
}

func (this PSQLProtectedEntity) GetInfo(ctx context.Context) (astrolabe.ProtectedEntityInfo, error) {
	psql, err := this.petm.getPostgresqlForPEID(ctx, this.id)

	if err != nil {
		return nil, errors.Wrapf(err, "getPostgresqlForPEID failed for %s", this.id.String())
	}

	dataS3Transport, err := astrolabe.NewS3DataTransportForPEID(this.id, this.petm.s3Config)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create S3 data transport")
	}

	data := []astrolabe.DataTransport{
		dataS3Transport,
	}

	mdS3Transport, err := astrolabe.NewS3MDTransportForPEID(this.id, this.petm.s3Config)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create S3 md transport")
	}

	md := []astrolabe.DataTransport{
		mdS3Transport,
	}

	combinedS3Transport, err := astrolabe.NewS3CombinedTransportForPEID(this.id, this.petm.s3Config)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create S3 combined transport")
	}

	combined := []astrolabe.DataTransport{
		combinedS3Transport,
	}

	retVal := astrolabe.NewProtectedEntityInfo(
		this.id,
		psql.Name,
		-1,
		data,
		md,
		combined,
		[]astrolabe.ProtectedEntityID{})
	return retVal, nil
}

func (this PSQLProtectedEntity) GetCombinedInfo(ctx context.Context) ([]astrolabe.ProtectedEntityInfo, error) {
	panic("implement me")
}

func (this PSQLProtectedEntity) Snapshot(ctx context.Context, params map[string]map[string]interface {}) (astrolabe.ProtectedEntitySnapshotID, error) {
	if this.id.HasSnapshot() {
		return astrolabe.ProtectedEntitySnapshotID{}, errors.New(fmt.Sprintf("pe %s is a snapshot, cannot snapshot again", this.id.String()))
	}
	snapshotUUID, err := uuid.NewRandom()
	if err != nil {
		return astrolabe.ProtectedEntitySnapshotID{}, errors.Wrap(err, "Failed to create new UUID")
	}
	snapshotID := astrolabe.NewProtectedEntitySnapshotID(snapshotUUID.String())


	err = this.petm.internalRepo.WriteProtectedEntity(ctx, this, snapshotID)
	if err != nil {
		return astrolabe.ProtectedEntitySnapshotID{}, errors.Wrap(err, "Failed to create new snapshot")
	}
	return snapshotID, nil
}

func (this PSQLProtectedEntity) ListSnapshots(ctx context.Context) ([]astrolabe.ProtectedEntitySnapshotID, error) {
	if this.id.HasSnapshot() {
		return nil, errors.New(fmt.Sprintf("pe %s is a snapshot, cannot list snapshots", this.id.String()))
	}

	return this.petm.internalRepo.ListSnapshotsForPEID(this.id)
}

func (this PSQLProtectedEntity) DeleteSnapshot(ctx context.Context, snapshotToDelete astrolabe.ProtectedEntitySnapshotID,
	params map[string]map[string]interface {}) (bool, error) {
	panic("implement me")
}

func (this PSQLProtectedEntity) GetInfoForSnapshot(ctx context.Context, snapshotID astrolabe.ProtectedEntitySnapshotID) (*astrolabe.ProtectedEntityInfo, error) {
	panic("implement me")
}

func (this PSQLProtectedEntity) GetComponents(ctx context.Context) ([]astrolabe.ProtectedEntity, error) {
	return []astrolabe.ProtectedEntity{}, nil
}

func (this PSQLProtectedEntity) GetID() astrolabe.ProtectedEntityID {
	return this.id
}

func (this PSQLProtectedEntity) GetDataReader(ctx context.Context) (io.ReadCloser, error) {
	if !this.id.HasSnapshot() {
		psql, err := this.petm.getPostgresqlForPEID(ctx, this.id)
		if err != nil {
			return nil, errors.Wrapf(err, "could not retrive psql resource for %s", this.id.String())
		}
		namespace := psql.Namespace
		pghost := psql.ObjectMeta.Name
		pgsecret, err := this.petm.KubeClient.Secrets(namespace).Get(ctx, "postgres."+pghost+".credentials", metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "could not retrieve secret")
		}
		pguser := string(pgsecret.Data["username"])
		pgpassword := string(pgsecret.Data["password"])
		fmt.Printf("pguser = %s, pgpassword = %s\n", pguser, pgpassword)
		dumpUUID, err := uuid.NewRandom()
		if err != nil {
			return nil, errors.Wrap(err, "could not create UUID")
		}
		podName := "snapshot-pg-" + dumpUUID.String()
		cmd := exec.Command("/usr/bin/kubectl", "run", "-n", namespace, podName, "--image=dpcpinternal/pg-dump:0.0.5",
			"--env", "PGPASSWORD=" + pgpassword, "--env",
			"PGHOST=" + pghost, "--env", "PGUSER=" +pguser, "-it", "--restart=Never", "--rm")

		cmdStdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, errors.Wrap(err, "Failed to get cmd's stdout")
		}
		err = cmd.Start()
		if err != nil {
			return nil, errors.Wrap(err, "Failed to start command")
		}

		return cmdStdout, nil
	}
	return this.petm.internalRepo.GetDataReaderForSnapshot(this.id)
}

func (this PSQLProtectedEntity) GetMetadataReader(ctx context.Context) (io.ReadCloser, error) {
	if !this.id.HasSnapshot() {
		psql, err := this.petm.getPostgresqlForPEID(ctx, this.id)
		if err != nil {
			return nil, errors.Wrapf(err, "could not retrive psql resource for %s", this.id.String())
		}
		psqlBytes, err := json.Marshal(psql)
		return ioutil.NopCloser(bytes.NewReader(psqlBytes)), nil
	}
	return this.petm.internalRepo.GetMetadataReaderForSnapshot(this.id)
}

func (this PSQLProtectedEntity) Overwrite(ctx context.Context, sourcePE astrolabe.ProtectedEntity, params map[string]map[string]interface{},
overwriteComponents bool) error {
	panic("implement me")
}