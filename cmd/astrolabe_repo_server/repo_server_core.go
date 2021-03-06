/*
 * Copyright 2019 the Astrolabe contributors
 * SPDX-License-Identifier: Apache-2.0
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"github.com/dsu-igeek/astrolabe-kopia/pkg/kopiarepo"
	"github.com/go-openapi/loads"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/gen/restapi"
	"github.com/vmware-tanzu/astrolabe/gen/restapi/operations"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"github.com/vmware-tanzu/astrolabe/pkg/server"
	"os"
	"strconv"

	//"github.com/labstack/gommon/log"
	"log"
)

func ServerMain(addonInits map[string]astrolabe.InitFunc) {
	server, _, err := ServerInit(addonInits)
	if err != nil {
		log.Println("Error initializing server = %v\n", err)
		return
	}
	defer server.Shutdown()

	// serve API
	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
}

func ServerInit(addonInits map[string]astrolabe.InitFunc) (*restapi.Server, astrolabe.ProtectedEntityManager, error) {
	confDirStr := flag.String("confDir", "", "Configuration directory")
	apiPortStr := flag.String("apiPort", "1323", "REST API port")
	insecure := flag.Bool("insecure", false, "Only use HTTP")
	flag.Parse()
	if *confDirStr == "" {
		flag.Usage()
		return nil, nil, errors.New("confDir is not defined")
	}
	apiPort, err := strconv.Atoi(*apiPortStr)
	if err != nil {
		log.Printf("apiPort %s is not an integer\n", *apiPortStr)
		os.Exit(1)
	}
	//pem := server.NewProtectedEntityManager(*confDirStr, addonInits, logrus.New())
	pem, err := kopiarepo.NewKopiaRepositoryProtectedEntityManager(*confDirStr, logrus.New())
	if err != nil {
		log.Printf("Could not initialize Kopia repo, err =%v\n", err)
		os.Exit(1)
	}
	tm := server.NewTaskManager()
	apiHandler := server.NewOpenAPIAstrolabeHandler(pem, tm)
	// load embedded swagger file
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		log.Fatalln(err)
	}

	// create new service API
	api := operations.NewAstrolabeAPI(swaggerSpec)
	server := restapi.NewServer(api)
	if *insecure {
		server.EnabledListeners = []string{"http"}
	}

	// parse flags
	flag.Parse()
	// set the port this service will be run on
	server.Port = apiPort

	apiHandler.AttachHandlers(api)
	return server, pem, nil
}
