# Namecheck

<p align="center">
<img src="assets/img/name.logo.png" alt="Namecheck logo" title="Namecheck logo" />
</p>

[![Docs](https://img.shields.io/badge/docs-current-brightgreen.svg)](https://pkg.go.dev/github.com/davidaparicio/namecheck)
[![Go Report Card](https://goreportcard.com/badge/davidaparicio/namecheck)](https://goreportcard.com/report/davidaparicio/namecheck)
[![Github](https://img.shields.io/static/v1?label=github&logo=github&color=E24329&message=main&style=flat-square)](https://github.com/davidaparicio/namecheck)
[![GitLab](https://img.shields.io/static/v1?label=gitlab&logo=gitlab&color=green&message=mirrored&style=flat-square)](https://gitlab.com/davidaparicio/namecheck)
[![froggit](https://img.shields.io/static/v1?label=froggit&logo=froggit&color=yellowgreen&message=mirrored&style=flat-square)](https://lab.frogg.it/davidaparicio/namecheck)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/davidaparicio/namecheck/blob/main/LICENSE.md)
[![Twitter](https://img.shields.io/twitter/follow/dadideo.svg?style=social)](https://twitter.com/intent/follow?screen_name=dadideo)
[![Maintenance](https://img.shields.io/maintenance/yes/2022.svg)]()

## Overview
A simple CLI and server to check a name availability on Twitter and GitHub.

## Remarks
Twitter checker is using a GCP Cloud Function, adding a network latency simulation, configurable through a parameter.

Server handles timeouts as recommended [Filippo Valsorda](https://github.com/FiloSottile) on the Cloudflare blog ([post 1 ](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/)/[ post 2](https://blog.cloudflare.com/exposing-go-on-the-internet/)) and [Ilija Eftimov](https://ieftimov.com/posts/make-resilient-golang-net-http-servers-using-timeouts-deadlines-context-cancellation/).

[ADR](https://github.blog/2020-08-13-why-write-adrs/) (Architecture decision record) about the usage of `gorilla/mux` can be found [here](https://www.alexedwards.net/blog/which-go-router-should-i-use).

## Contribute

Works on my machine - and yours ! Spin up pre-configured, standardized dev environments of this repository, by clicking on the button below.

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#/https://github.com/davidaparicio/namecheck)

## Original project
Fork of the opensource project by [@jub0bs](https://github.com/jub0bs/), available on GitHub at [https://github.com/jub0bs/namecheck](https://github.com/jub0bs/namecheck).