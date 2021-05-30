package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/dsu-igeek/astrolabe-kopia/pkg/kopiarepo"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	kubernetes "github.com/vmware-tanzu/astrolabe-velero/pkg/k8sns"
	"github.com/vmware-tanzu/astrolabe/pkg/s3repository"
	ebs_astrolabe "github.com/vmware-tanzu/velero-plugin-for-aws/pkg/ebs-astrolabe"
	"strings"

	// restClient is the underlying REST/Swagger client
	restClient "github.com/vmware-tanzu/astrolabe/gen/client"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"github.com/dsu-igeek/astrolabe-demo/pkg/psql"
	// astrolabeClient is the Astrolabe API on top of the REST client
	astrolabeClient "github.com/vmware-tanzu/astrolabe/pkg/client"
	"github.com/vmware-tanzu/astrolabe/pkg/server"
	"io"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Usage: "Astrolabe server",
			},
			&cli.StringFlag{
				Name:     "destHost",
				Usage:    "Optional different destination Astrolabe server for cp",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "insecure",
				Usage:    "Only use HTTP",
				Required: false,
				Hidden:   false,
				Value:    false,
			},
			&cli.StringFlag{
				Name:  "confDir",
				Usage: "Configuration directory",
			},
			&cli.StringFlag{
				Name:     "destConfDir",
				Usage:    "Optional different destination Astrolabe configuration directory for cp",
				Required: false,
			},
			&cli.StringFlag{
				Name: "s3Repo",
				Usage: "Configuration to use an S3 repo as the primary repository.  Formatted as '<region>:<bucket>:<prefix>'",
				Required: false,
			},
			&cli.StringFlag{
				Name: "destS3Repo",
				Usage: "Configuration to use an S3 repo as the optional destination repository.  Formatted as '<region>:<bucket>:<prefix>'",
				Required: false,
			},
			&cli.StringFlag{
				Name: "kopiaRepo",
				Usage: "Kopia repo directory",
				Required: false,
			},
			&cli.StringFlag{
				Name: "destKopiaRepo",
				Usage: "Kopia repo directory to use as the optional destination repository",
				Required: false,
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "types",
				Usage:  "shows Protected Entity Types",
				Action: types,
			},
			{
				Name:   "ls",
				Usage:  "lists entities for a type",
				Action: ls,
			},
			{
				Name:      "show",
				Usage:     "shows info for a protected entity",
				ArgsUsage: "<protected entity id>",
				Action:    show,
			},
			{
				Name:      "lssn",
				Usage:     "lists snapshots for a Protected Entity",
				Action:    lssn,
				ArgsUsage: "<protected entity id>",
			},
			{
				Name:      "snap",
				Usage:     "snapshots a Protected Entity",
				Action:    snap,
				ArgsUsage: "<protected entity id>",
			},
			{
				Name:      "rmsn",
				Usage:     "removes a Protected Entity snapshot",
				Action:    rmsn,
				ArgsUsage: "<protected entity snapshot id>",
			},
			{
				Name:      "cp",
				Usage:     "copies a Protected Entity snapshot",
				Action:    cp,
				ArgsUsage: "<src> <dest>",
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func setupProtectedEntityManager(context *cli.Context) (pem astrolabe.ProtectedEntityManager, err error) {
	pem, _, err = setupProtectedEntityManagers(context, false)
	return
}

func countNonEmptyStrings(strings []string) (count int) {
	for _, checkStr := range strings {
		if checkStr != "" {
			count++
		}
	}
	return
}

func setupProtectedEntityManagers(context *cli.Context, allowDual bool) (srcPem astrolabe.ProtectedEntityManager,
	destPem astrolabe.ProtectedEntityManager, err error) {
	confDirStr := context.String("confDir")
	s3RepoStr := context.String("s3Repo")
	kopiaRepoDir := context.String("kopiaRepo")
	if countNonEmptyStrings([]string{confDirStr, s3RepoStr, kopiaRepoDir}) > 1 {
		err = errors.New("Only set one of confDir, kopiaRepo and s3Repo")
		return
	}
	if confDirStr != "" {
		addOnInits := make(map[string]server.InitFunc)
		addOnInits["psql"] = psql.NewPSQLProtectedEntityTypeManager
		addOnInits["ebs"] = ebs_astrolabe.NewEBSProtectedEntityTypeManager
		addOnInits["k8sns"] = kubernetes.NewKubernetesNamespaceProtectedEntityTypeManagerFromConfig
		srcPem = server.NewProtectedEntityManager(confDirStr, addOnInits, logrus.New())
	}
	if s3RepoStr != "" {
		srcPem, err = createS3Repo(s3RepoStr)
		if err != nil {
			return
		}
	}

	if kopiaRepoDir != "" {
		srcPem, err = createKopiaRepo(kopiaRepoDir)
		if err != nil {
			return
		}
	}
	if srcPem == nil {
		host := context.String("host")
		if host != "" {
			srcPem, err = setupHostPEM(host, context)
			if err != nil {
				return
			}
		}
	}

	if allowDual {
		destConfDirStr := context.String("destConfDir")
		destS3RepoStr := context.String("destS3Repo")
		destKopiaRepoStr := context.String("destKopiaRepo")
		if countNonEmptyStrings([]string{destConfDirStr, destS3RepoStr, destKopiaRepoStr}) > 1 {
			err = errors.New("Only set one of destConfDir, destS3Repo and destKopiaRepo")
			return
		}
		if destConfDirStr != "" {
			destPem = server.NewProtectedEntityManager(confDirStr, nil, logrus.New())
		}
		if destS3RepoStr != "" {
			destPem, err = createS3Repo(destS3RepoStr)
			if err != nil {
				return
			}
		}
		if destKopiaRepoStr != "" {
			destPem, err = createKopiaRepo(destKopiaRepoStr)
			if err != nil {
				return
			}
		}
		if destPem == nil {
			destHost := context.String("destHost")
			if destHost != "" {
				destPem, err = setupHostPEM(destHost, context)
				if err != nil {
					return
				}
			}
		}
	}
	if srcPem == nil {
		err = errors.New("No source Protected Entity Manager config info specified")
	}
	if destPem == nil {
		destPem = srcPem
	}
	return
}

func setupHostPEM(host string, c *cli.Context) (hostPem astrolabe.ProtectedEntityManager, err error) {
	insecure := c.Bool("insecure")
	transport := restClient.DefaultTransportConfig()
	transport.Host = host
	if insecure {
		transport.Schemes = []string{"http"}
	}

	restClient := restClient.NewHTTPClientWithConfig(nil, transport)
	hostPem, err = astrolabeClient.NewClientProtectedEntityManager(restClient)
	if err != nil {
		err = errors.Wrap(err, "Failed to create new ClientProtectedEntityManager")
		return
	}
	return
}

func types(c *cli.Context) error {
	pem, err := setupProtectedEntityManager(c)
	if err != nil {
		log.Fatalf("Could not setup protected entity manager, err =%v", err)
	}
	for _, curPETM := range pem.ListEntityTypeManagers() {
		fmt.Println(curPETM.GetTypeName())
	}
	return nil
}

func ls(c *cli.Context) error {
	pem, err := setupProtectedEntityManager(c)
	if err != nil {
		log.Fatalf("Could not setup protected entity manager, err =%v", err)
	}
	peType := c.Args().First()
	petm := pem.GetProtectedEntityTypeManager(peType)
	if petm == nil {
		log.Fatalf("Could not find type named %s", peType)
	}
	peIDs, err := petm.GetProtectedEntities(context.Background())
	if err != nil {
		log.Fatalf("Could not retrieve protected entities for type %s err:%b", peType, err)
	}

	for _, curPEID := range peIDs {
		fmt.Println(curPEID.String())
	}
	return nil
}

func lssn(c *cli.Context) error {
	peIDStr := c.Args().First()
	peID, err := astrolabe.NewProtectedEntityIDFromString(peIDStr)
	if err != nil {
		log.Fatalf("Could not parse protected entity ID %s, err: %v", peIDStr, err)
	}

	pem, err := setupProtectedEntityManager(c)
	if err != nil {
		log.Fatalf("Could not setup protected entity manager, err =%v", err)
	}

	pe, err := pem.GetProtectedEntity(context.Background(), peID)
	if err != nil {
		log.Fatalf("Could not retrieve protected entity ID %s, err: %v", peIDStr, err)
	}

	snaps, err := pe.ListSnapshots(context.Background())
	if err != nil {
		log.Fatalf("Could not get snapshots for protected entity ID %s, err: %v", peIDStr, err)
	}

	for _, curSnapshotID := range snaps {
		curPESnapshotID := peID.IDWithSnapshot(curSnapshotID)
		fmt.Println(curPESnapshotID.String())
	}
	return nil
}

func show(c *cli.Context) error {
	peIDStr := c.Args().First()
	peID, err := astrolabe.NewProtectedEntityIDFromString(peIDStr)
	if err != nil {
		log.Fatalf("Could not parse protected entity ID %s, err: %v", peIDStr, err)
	}

	pem, err := setupProtectedEntityManager(c)
	if err != nil {
		log.Fatalf("Could not setup protected entity manager, err =%v", err)
	}

	pe, err := pem.GetProtectedEntity(context.Background(), peID)
	if err != nil {
		log.Fatalf("Could not retrieve protected entity ID %s, err: %v", peIDStr, err)
	}

	info, err := pe.GetInfo(context.Background())
	if err != nil {
		log.Fatalf("Could not retrieve info for %s, err: %v", peIDStr, err)
	}
	fmt.Printf("%v\n", info)
	fmt.Println("Components:")
	components, err := pe.GetComponents(context.Background())
	for component := range components {
		fmt.Printf("%v\n", component)
	}
	return nil
}

func snap(c *cli.Context) error {
	peIDStr := c.Args().First()
	peID, err := astrolabe.NewProtectedEntityIDFromString(peIDStr)
	if err != nil {
		log.Fatalf("Could not parse protected entity ID %s, err: %v", peIDStr, err)
	}

	pem, err := setupProtectedEntityManager(c)
	if err != nil {
		log.Fatalf("Could not setup protected entity manager, err =%v", err)
	}

	pe, err := pem.GetProtectedEntity(context.Background(), peID)
	if err != nil {
		log.Fatalf("Could not retrieve protected entity ID %s, err: %v", peIDStr, err)
	}
	snap, err := pe.Snapshot(context.Background(), make(map[string]map[string]interface{}))
	if err != nil {
		log.Fatalf("Could not snapshot protected entity ID %s, err: %v", peIDStr, err)
	}
	fmt.Println(snap.String())
	return nil
}

func rmsn(c *cli.Context) error {
	peIDStr := c.Args().First()
	peID, err := astrolabe.NewProtectedEntityIDFromString(peIDStr)
	if err != nil {
		log.Fatalf("Could not parse protected entity ID %s, err: %v", peIDStr, err)
	}
	if !peID.HasSnapshot() {
		log.Fatalf("Protected entity ID %s does not have a snapshot ID", peIDStr)
	}

	pem, err := setupProtectedEntityManager(c)
	if err != nil {
		log.Fatalf("Could not setup protected entity manager, err =%v", err)
	}

	pe, err := pem.GetProtectedEntity(context.Background(), peID)
	if err != nil {
		log.Fatalf("Could not retrieve protected entity ID %s, err: %v", peIDStr, err)
	}
	success, err := pe.DeleteSnapshot(context.Background(), peID.GetSnapshotID(), make(map[string]map[string]interface{}))
	if err != nil {
		log.Fatalf("Could not remove snapshot ID %s, err: %v", peIDStr, err)
	}
	if success {
		log.Printf("Removed snapshot %s\n", peIDStr)
	}
	return nil
}

func cp(c *cli.Context) error {
	if c.NArg() != 2 {
		log.Fatalf("Expected two arguments for cp, got %d", c.NArg())
	}
	srcStr := c.Args().First()
	destStr := c.Args().Get(1)
	ctx := context.Background()
	var err error
	var srcPE, destPE astrolabe.ProtectedEntity
	var srcPEID, destPEID astrolabe.ProtectedEntityID
	var srcFile, destFile string
	srcPEID, err = astrolabe.NewProtectedEntityIDFromString(srcStr)
	// TODO - come up with a better way to separate PEs from files.
	if err != nil || strings.HasPrefix(srcStr, "/") {
		srcFile = srcStr
	}
	destPEID, err = astrolabe.NewProtectedEntityIDFromString(destStr)
	if err != nil || strings.HasPrefix(destStr, "/") {
		destFile = destStr
	}
	srcPEM, destPEM, err := setupProtectedEntityManagers(c, true)
	if err != nil {
		log.Fatalf("Could not setup protected entity manager, err = %v", err)
	}

	var reader io.ReadCloser
	var writer io.WriteCloser
	fmt.Printf("cp from ")
	if srcFile != "" {
		fmt.Printf("file %s", srcFile)
	} else {
		if srcPEM == nil {
			log.Fatal("No source Protected Entity Manager specified\n")
		}
		fmt.Printf("pe %s", srcPEID.String())
		srcPE, err = srcPEM.GetProtectedEntity(ctx, srcPEID)
		if err != nil {
			log.Fatalf("Could not retrieve protected entity ID %s, err: %v", srcPEID.String(), err)
		}
	}
	fmt.Printf(" to ")
	if destFile != "" {
		fmt.Printf("file %s", destFile)
	} else {
		fmt.Printf("pe %s", destPEID.String())
	}
	fmt.Printf("\n")

	var bytesCopied int64
	if srcPE != nil && destFile != "" {
		// copy a src pe snapshot to a dest file
		var zipFileWriter io.WriteCloser
		zipFileWriter, err = os.Create(destFile)
		if err != nil {
			log.Fatalf("Could not create file %s, err: %v", destFile, err)
		}
		defer func() {
			if err := zipFileWriter.Close(); err != nil {
				log.Fatalf("Could not close file %s, err: %v", destFile, err)
			}
		}()

		bytesCopied, err = astrolabe.ZipProtectedEntityToFile(ctx, srcPE, zipFileWriter)
	} else if srcFile != "" && destPEM != nil{
		// copy a src file to an existing or new dest pe. BTW, the dest pe should be unconsumed.
		srcFileReader, err := os.Open(srcFile)
		if err != nil {
			log.Fatalf("Could not open srcFile %s, err = %v", srcFile, err)
		}
		defer srcFileReader.Close()

		fileInfo, err := srcFileReader.Stat()
		if err != nil {
			log.Fatalf("Got err %v getting stat for source file %v", err, srcFile)
		}

		srcPE, err := astrolabe.GetPEFromZipStream(ctx, srcFileReader, fileInfo.Size())
		if err != nil {
			return errors.Errorf("Got err %v when unzipping PE from the source zip file", err)
		}

		destPE, err = destPEM.GetProtectedEntity(ctx, destPEID)
		if err != nil {
			log.Printf("Could not retrieve protected entity ID %s, err: %v, will try to create object", destPEID.String(), err)
		}

		var params map[string]map[string]interface{}
		if destPE != nil {
			if err := destPE.Overwrite(ctx, srcPE, params, true); err != nil {
				return errors.Errorf("Got err %v overwriting the dest PE %v with source PE %v", err, destPE.GetID(), srcPE.GetID())
			}
		} else {
			destPETM := destPEM.GetProtectedEntityTypeManager(destPEID.GetPeType())
			if destPETM == nil {
				return errors.Errorf("Could not get destination PETM for type %s", destPEID.GetPeType())
			}
			_, err := destPETM.Copy(ctx, srcPE, params, astrolabe.AllocateNewObject)
			if err != nil {
				return errors.WithMessagef(err, "Could not copy %s", destPEID.String())
			}
		}
		bytesCopied = fileInfo.Size()
	} else {
		reader, err = os.Open(srcFile)
		if err != nil {
			log.Fatalf("Could not open srcFile %s, err = %v", srcFile, err)
		}
		defer reader.Close()

		bytesCopied, err = io.Copy(writer, reader)
	}

	if err != nil {
		log.Fatalf("Error copying %v", err)
	}

	fmt.Printf("Copied %d bytes\n", bytesCopied)

	return nil
}

func createS3Repo(s3RepoStr string) (s3Pem astrolabe.ProtectedEntityManager, err error) {
	s3Components := strings.Split(s3RepoStr, ":")
	if len(s3Components) != 3 {
		err = errors.Errorf("s3Repo arguments should be '<region>:<bucket>:<prefix>', received '%s'", s3RepoStr)
		return
	}
	s3Region := s3Components[0]
	s3Bucket := s3Components[1]
	s3Prefix := s3Components[2]
	var sess *session.Session
	sess, err = session.NewSession(&aws.Config{
		Region: aws.String(s3Region)},
	)
	if err != nil {
		return
	}
	s3Pem, err = s3repository.NewS3RepositoryProtectedEntityManager(*sess, s3Bucket, s3Prefix, logrus.New())
	if err != nil {
		return
	}
	return
}

func createKopiaRepo(kopiaRepoStr string) (kopiaPEM astrolabe.ProtectedEntityManager, err error) {
	return kopiarepo.NewKopiaRepositoryProtectedEntityManager(kopiaRepoStr, logrus.StandardLogger())
}