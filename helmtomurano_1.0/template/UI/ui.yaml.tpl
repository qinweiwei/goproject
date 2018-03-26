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

