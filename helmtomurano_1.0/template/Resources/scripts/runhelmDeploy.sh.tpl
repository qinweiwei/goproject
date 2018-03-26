#!/bin/bash
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
