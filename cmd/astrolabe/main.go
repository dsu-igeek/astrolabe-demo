package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/vmware-tanzu/astrolabe/pkg/psql"
	"github.com/vmware-tanzu/astrolabe/pkg/s3repository"
	"strings"

	// restClient is the underlying REST/Swagger client
	restClient "github.com/vmware-tanzu/astrolabe/gen/client"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
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

func setupProtectedEntityManagers(context *cli.Context, allowDual bool) (srcPem astrolabe.ProtectedEntityManager,
	destPem astrolabe.ProtectedEntityManager, err error) {
	confDirStr := context.String("confDir")
	s3RepoStr := context.String("s3Repo")
	if confDirStr != "" && s3RepoStr != "" {
		err = errors.New("Cannot set confDir and s3Repo simultaneously")
		return
	}
	if confDirStr != "" {
		addOnInits := make(map[string]server.InitFunc)
		addOnInits["psql"] = psql.NewPSQLProtectedEntityTypeManager
		srcPem = server.NewProtectedEntityManager(confDirStr, nil, logrus.New())
	}
	if s3RepoStr != "" {

		srcPem, err = createS3Repo(s3RepoStr)
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
		if destConfDirStr != "" && destS3RepoStr != "" {
			err = errors.New("Cannot set destConfDirStr and destS3Repo simultaneously")
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
	var err error
	var srcPEID, destPEID astrolabe.ProtectedEntityID
	var srcFile, destFile string
	srcPEID, err = astrolabe.NewProtectedEntityIDFromString(srcStr)
	if err != nil {
		srcFile = srcStr
	}
	destPEID, err = astrolabe.NewProtectedEntityIDFromString(destStr)
	if err != nil {
		destFile = destStr
	}
	pem, err := setupProtectedEntityManager(c)
	if err != nil {
		log.Fatalf("Could not setup protected entity manager, err = %v", err)
	}

	var reader io.ReadCloser
	var writer io.WriteCloser
	fmt.Printf("cp from ")
	if srcFile != "" {
		fmt.Printf("file %s", srcFile)
		reader, err = os.Open(srcFile)
		if err != nil {
			log.Fatalf("Could not open srcFile %s, err = %v", srcFile, err)
		}
	} else {
		fmt.Printf("pe %s", srcPEID.String())
		srcPE, err := pem.GetProtectedEntity(context.Background(), srcPEID)
		if err != nil {
			log.Fatalf("Could not retrieve protected entity ID %s, err: %v", srcPEID.String(), err)
		}
		var dw io.WriteCloser
		reader, dw = io.Pipe()
		go zipPE(context.Background(), srcPE, dw)
	}
	fmt.Printf(" to ")
	if destFile != "" {
		fmt.Printf("file %s", destFile)
		writer, err = os.Create(destFile)
		if err != nil {
			log.Fatalf("Could not create file %s, err: %v", destFile, err)
		}
	} else {
		fmt.Printf("pe %s", destPEID.String())
	}
	fmt.Printf("\n")

	bytesCopied, err := io.Copy(writer, reader)
	if err != nil {
		log.Fatalf("Error copying %v", err)
	}
	fmt.Printf("Copied %d bytes\n", bytesCopied)
	return nil
}

func zipPE(ctx context.Context, pe astrolabe.ProtectedEntity, writer io.WriteCloser) {
	defer writer.Close()
	err := astrolabe.ZipProtectedEntityToWriter(ctx, pe, writer)
	if err != nil {
		log.Fatalf("Failed to zip protected entity %s, err = %v", pe.GetID().String(), err)
	}
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