package psql

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestProtectedEntityTypeManager(t *testing.T) {
	testPETM, err := NewPSQLProtectedEntityTypeManager()
	if err != nil {
		t.Fatalf("Could not get PSQLProtectedEntityTypeManager %v", err)
	}
	ctx := context.TODO()
	ids, err := testPETM.GetProtectedEntities(ctx)
	if err != nil {
		t.Fatalf("Could not get protected entities err %v", err)
	}
	pe, err := testPETM.GetProtectedEntity(ctx, ids[0])
	if err != nil {
		t.Fatalf("Could not get protected entity err %v", err)
	}

	snapshotID, err := pe.Snapshot(context.TODO())
	if err != nil {
		t.Fatalf("Could not create snapshot, err %v", err)
	}
	fmt.Printf("Created snapshot id %s\n", snapshotID.String())

	snapIDs, err := pe.ListSnapshots(ctx)
	if err != nil {
		t.Fatalf("Could not list snapshots err %v", err)
	}
	for _, curSnapID := range snapIDs {
		snapPEID := pe.GetID().IDWithSnapshot(curSnapID)
		snapPE, err := testPETM.GetProtectedEntity(ctx, snapPEID)
		if err != nil {
			t.Fatalf("Could not get snapshot PE err %v", err)
		}
		dataReader, err := snapPE.GetDataReader(ctx)
		if err != nil {
			t.Fatalf("Could not get snapshot data reader err %v", err)
		}
		io.Copy(os.Stdout, dataReader)
	}
}
