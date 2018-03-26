package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
        "io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	class = `#  Licensed under the Apache License, Version 2.0 (the "License"); you may
#  not use this file except in compliance with the License. You may obtain
#  a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
#  WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
#  License for the specific language governing permissions and limitations
#  under the License.

Namespaces:
  =: com.example.pdmi
  std: io.murano
  res: io.murano.resources
  sys: io.murano.system


Name: {{ .Name }}

Extends: std:Application

Properties:
  {{ range .Config -}}
  {{ .Name }}:
    Contract: $.string().notNull()
    Default: {{ .Default }}
  {{ end -}}
  helminstance:
    Contract: $.class(HelmInstance).notNull()
  nodeport:
    Contract: $.string()
    Usage: Out

Methods:
  initialize:
    Body:
      - $._environment: $.find(std:Environment).require()

  deploy:
    Body:
      - If: not $.getAttr(deployed, false)
        Then:
          - $._environment.reporter.report($this, 'starting  Deploy Helm {{ .Name }} release.')
          - $resources: new(sys:Resources)
          - $template: $resources.yaml('DeployHelm.template').bind(dict(
                  {{- range $i, $v := .Config -}}
                  {{- if $i -}}
                  ,
                  {{ end }}
                  {{- .Name }} => $.{{ .Name -}}
                  {{ end -}}
              ))
          - $._environment.reporter.report($this, 'Instance is created. Deploying {{ .Name }} release')
          - $nodeport: $.helminstance.instance.agent.call($template, $resources)
          - $._environment.reporter.report($this, format('Helm release is available, IP 10.18.74.203 Port {0}', $nodeport))
          - $.setAttr(deployed, true)
  destroy:
    Body:
      - If: $.getAttr(deployed, false)
        Then:
          - $._environment.reporter.report($this, 'delete helm release.')
          - $resources: new(sys:Resources)
          - $template: $resources.yaml('DestroyHelm.template').bind(dict({{ (index .Config 0).Name }} => $.{{ (index .Config 0).Name }}))
          - $.helminstance.instance.agent.call($template, $resources)
          - $._environment.reporter.report($this, format('The helm release {0} is destroyed !', $.{{ (index .Config 0).Name }}))

`
	resource_deploy = `#  Licensed under the Apache License, Version 2.0 (the "License"); you may
#  not use this file except in compliance with the License. You may obtain
#  a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
#  WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
#  License for the specific language governing permissions and limitations
#  under the License.

FormatVersion: 2.0.0
Version: 1.0.0
Name: Deploy Helm

Parameters:
  {{ range .Config -}}
  {{ .Name }}: ${{ .Name }}
  {{ end }}
Body: |
  return helmDeploy('{{- range $i, $v := .Config -}} { {{- $i -}} } {{ end -}}'.format({{- range $i, $v := .Config -}} {{if $i}}, {{end}}args.{{$v.Name}} {{- end -}})).stdout
Scripts:
  helmDeploy:
    Type: Application
    Version: 1.0.0
    EntryPoint: runhelmDeploy.sh
    Files: []
    Options:
      captureStdout: true
      captureStderr: true
`
	resource_destroy = `#  Licensed under the Apache License, Version 2.0 (the "License"); you may
#  not use this file except in compliance with the License. You may obtain
#  a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
#  WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
#  License for the specific language governing permissions and limitations
#  under the License.

FormatVersion: 2.0.0
Version: 1.0.0
Name: Destroy Helm

Parameters:
  {{ (index .Config 0).Name }}: ${{ (index .Config 0).Name }}
Body: |
  return helmDestroy('{0}'.format(args.{{ (index .Config 0).Name }})).stdout

Scripts:
  helmDestroy:
    Type: Application
    Version: 1.0.0
    EntryPoint: runhelmlDestroy.sh
    Files: []
    Options:
      captureStdout: true
      captureStderr: true
`
	resource_scripts_deploy = `#!/bin/bash
#  Licensed under the Apache License, Version 2.0 (the "License"); you may
#  not use this file except in compliance with the License. You may obtain
#  a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
#  WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
#  License for the specific language governing permissions and limitations
#  under the License.
{{ range $i, $v:= .Config -}}
#${{ add $i}} {{ .Name}}
{{ end }}
rm -rf /root/.helm
echo "Parameters {{ range $i,$v := .Config -}} ${{ add $i }} {{ end -}} " | tee >>/tmp/test.txt
helm init -c --stable-repo-url http://10.18.74.203:8000 2>&1 | tee >> /tmp/test.txt
echo "helm init successful! " | tee >>/tmp/test.txt
helm repo update 2>&1 | tee >>/tmp/test.txt
echo "helm repo update successful !" | tee>>/tmp/test.txt
helm install {{ range $i,$v := .Config -}} {{- if not $i -}} --name ${{ add $i }} {{ else }} --set {{ replace $v.Name }}=${{ add $i }} {{ end -}} {{ end -}} --wait {{ .Repo }}/{{ .Name }} 2>&1 | tee >> /tmp/test.txt
echo "helm install successful1! " | tee >> /tmp/test.txt
nodeport=$(kubectl get svc $1-{{ .Name }} -o json | jq -r '.spec.ports | .[0].nodePort')
echo $nodeport
`
	resource_scripts_destroy = `#!/bin/bash
#  Licensed under the Apache License, Version 2.0 (the "License"); you may
#  not use this file except in compliance with the License. You may obtain
#  a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
#  WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
#  License for the specific language governing permissions and limitations
#  under the License.
#$1 name
echo "Parameters $1" | tee >>/tmp/test.txt
helm delete $1 --purge | tee >> /tmp/test.txt
echo "helm delete {{ .Name }}successful1! " | tee >> /tmp/test.txt

`
	ui = `#  Licensed under the Apache License, Version 2.0 (the "License"); you may
#  not use this file except in compliance with the License. You may obtain
#  a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
#  WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
#  License for the specific language governing permissions and limitations
#  under the License.

Version: 2.0
Application:
  ?:
    type: com.example.pdmi.{{ .Name }}
  {{ range .Config -}}
  {{ .Name }}: $.appConfiguration.{{ .Name }}
  {{ end -}}
  helminstance: $.appConfiguration.instance

Forms:
  - appConfiguration:
      fields:
        - name: license
          type: string
          description: Apache License, Version 2.0
          hidden: true
          required: false
        {{ range .Config -}}
        - name: {{ .Name }}
          type: string
          label: {{ .Describe }}
          description: >-
            {{ .Describe }}
          required: true
          initial: {{ .Default }}
        {{ end -}}
        - name: instance
          type: com.example.pdmi.HelmInstance
          label: Helm instance to deploy helm application
          description: >-
             Select to deploy helm application
          required: true

`
	manifest = `#  Licensed under the Apache License, Version 2.0 (the "License"); you may
#  not use this file except in compliance with the License. You may obtain
#  a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
#  WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
#  License for the specific language governing permissions and limitations
#  under the License.

Format: 1.0
Type: Application
FullName: com.example.pdmi.{{ .Name }}
Name: Helm {{ .Name }}
Description: |
 The Apache HTTP Server Project is an effort to develop and maintain an
 open-source HTTP server for modern operating systems including UNIX and
 Windows NT. The goal of this project is to provide a secure, efficient and
 extensible server that provides HTTP services in sync with the current HTTP
 standards.
 Apache httpd has been the most popular web server on the Internet since
 April 1996, and celebrated its 17th birthday as a project this February.
Author: 'PDMI, Inc'
Tags: [Helm, {{ .Name }}]
Classes:
 com.example.pdmi.{{ .Name }}: HelmTemplate.yaml
Require:
  com.example.pdmi.HelmInstance:
`
)

type ChartConfig struct {
	Name     string `json:"name"`
	Describe string `json:"describe"`
	Default  string `json:"default"`
}

type HelmChart struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Repo    string `json:"repo"`
	Config  []ChartConfig
}

func CompressZip(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}
        if !strings.HasSuffix(source, "/"){
        	source += "/"
        }

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}
	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			//header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			header.Name = strings.TrimPrefix(path, source)
                        fmt.Println(header.Name)
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

func handler(outdir string, config string) error {
	var helm HelmChart

	data, err := ioutil.ReadFile(config)
	if err != nil {
		log.Println("Open config file err: ", err)
		return err
	}

	err = json.Unmarshal(data, &helm)
	if err != nil {
		log.Println("json Unmarshal err: ", err)
		return err
	}

	funcMap := template.FuncMap{
		"add": func(i int) int {
			return i + 1
		},
		"replace": func(i string) string {
			j := strings.Replace(i, "_", ".", -1)
			return j
		},
	}
	dir, err := ioutil.TempDir("/tmp", "tmp") //在DIR目录下创建tmp为目录名前缀的目录，DIR必须存在，否则创建不成功
	if err != nil {
		log.Println("mkdir temp Dir fail: ", err)
		return err
	}

	templates := map[string]string{
		"Classes/HelmTemplate.yaml":           class,
		"Resources/DeployHelm.template":       resource_deploy,
		"Resources/DestroyHelm.template":          resource_destroy,
		"Resources/scripts/runhelmDeploy.sh":  resource_scripts_deploy,
		"Resources/scripts/runhelmDestroy.sh": resource_scripts_destroy,
		"UI/ui.yaml":                          ui,
		"manifest.yaml":                       manifest,
	}
	for index, value := range templates {
		path := dir + "/" + index
                fmt.Println(path)
		os.MkdirAll(filepath.Dir(path), 0755)
		f, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
		defer f.Close()
		t := template.Must(template.New("").Funcs(funcMap).Parse(value))
		err = t.Execute(f, helm)
		if err != nil {
			log.Println("executing template:", err)
			return err
		}
	}
        if !strings.HasSuffix(outdir, "/"){
		outdir += "/"
	}
        
        name := outdir + helm.Name + ".zip"
	CompressZip(dir, name)

	return nil

}

func main() {
	var outdir, config string
	flag.StringVar(&outdir, "o", "", "The directory of murano package ")
	flag.StringVar(&config, "c", "", "The config file of helm ")

	flag.Parse()
	err := handler(outdir, config)
	if err != nil {
		log.Println("handler murano package fail:", err)
	}
}
