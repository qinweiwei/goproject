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
