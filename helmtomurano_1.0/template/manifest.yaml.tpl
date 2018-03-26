﻿#  Licensed under the Apache License, Version 2.0 (the "License"); you may
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
