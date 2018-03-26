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

