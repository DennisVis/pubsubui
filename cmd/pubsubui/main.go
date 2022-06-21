// Copyright 2022 Dennis Vis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"

	"github.com/DennisVis/pubsubui/internal/pubsubui"
	fe "github.com/DennisVis/pubsubui/web/pubsubui"
)

func main() {
	log.Println(pubsubui.LogPrefix + "running with frontend")

	distFolder, err := fs.Sub(fe.Dist, "dist")
	if err != nil {
		log.Fatalf("%+v", errors.Wrap(err, "could not get handle to dist folder"))
		return
	}

	err = pubsubui.RunApp([]func(chi.Router){
		func(r chi.Router) {
			r.Handle("/*", http.FileServer(http.FS(distFolder)))
		},
	}...)
	if err != nil {
		log.Fatalf(fmt.Sprintf(pubsubui.LogPrefix+"%+v\n", err))
	}
}
