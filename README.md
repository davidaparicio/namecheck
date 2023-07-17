# Namecheck

<p align="center">
<img src="assets/img/name.logo.png" alt="Namecheck logo" title="Namecheck logo" />
</p>

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://pkg.go.dev/github.com/davidaparicio/namecheck)
[![Go Report Card](https://goreportcard.com/badge/davidaparicio/namecheck)](https://goreportcard.com/report/davidaparicio/namecheck)
[![codecov](https://codecov.io/gh/davidaparicio/namecheck/branch/main/graph/badge.svg?token=VYP4LAODQ6)](https://codecov.io/gh/davidaparicio/namecheck)
[![build](https://github.com/davidaparicio/namecheck/actions/workflows/goreleaser.yml/badge.svg)](https://github.com/davidaparicio/namecheck/actions/workflows/goreleaser.yml)
[![Github](https://img.shields.io/static/v1?label=github&logo=github&color=E24329&message=main&style=flat-square)](https://github.com/davidaparicio/namecheck)
[![GitLab](https://img.shields.io/static/v1?label=gitlab&logo=gitlab&color=green&message=mirrored&style=flat-square)](https://gitlab.com/davidaparicio/namecheck)
[![Froggit](https://img.shields.io/static/v1?label=froggit&logo=froggit&color=red&message=no&style=flat-square)](https://lab.frogg.it/davidaparicio/namecheck)

[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=davidaparicio_namecheck&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=davidaparicio_namecheck)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=davidaparicio_namecheck&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=davidaparicio_namecheck)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/davidaparicio/namecheck/blob/main/LICENSE.md)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fdavidaparicio%2Fnamecheck.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fdavidaparicio%2Fnamecheck?ref=badge_shield)
[![Maintenance](https://img.shields.io/maintenance/yes/2023.svg)]()
[![Twitter](https://img.shields.io/twitter/follow/dadideo.svg?style=social)](https://twitter.com/intent/follow?screen_name=dadideo)


## Overview
A simple CLI and server to check a name availability on Twitter and GitHub.

## How to use it

If you have already ```Docker``` installed on your laptop

```docker run davidaparicio/namecheck:<TAG/VERSION_LIKE_v0.0.5> <PSEUDO_TO_CHECK>```

If not, you need ```Go``` and all dependencies

```go run cmd/cli/main.go <PSEUDO_TO_CHECK>```

or the server with 

```go run cmd/server/main.go```

and check with a curl command ```curl http://localhost:8080/check?username=<PSEUDO_TO_CHECK>```

For more information, you can see [examples here](EXAMPLES.md)


## Remarks
Twitter checker is using a GCP Cloud Function, adding a network latency simulation, configurable through a parameter.

Server handles timeouts as recommended [Filippo Valsorda](https://github.com/FiloSottile) on the Cloudflare blog ([post 1 ](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/)/[ post 2](https://blog.cloudflare.com/exposing-go-on-the-internet/)) and [Ilija Eftimov](https://ieftimov.com/posts/make-resilient-golang-net-http-servers-using-timeouts-deadlines-context-cancellation/).

[ADR](https://github.blog/2020-08-13-why-write-adrs/) (Architecture decision record) about the usage of `gorilla/mux` can be found [here](https://www.alexedwards.net/blog/which-go-router-should-i-use).

## Contribute

Works on my machine - and yours ! Spin up pre-configured, standardized dev environments of this repository, by clicking on the button below.

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#/https://github.com/davidaparicio/namecheck)

## Code coverage

![coverage](https://codecov.io/gh/davidaparicio/namecheck/branch/main/graphs/sunburst.svg?token=VYP4LAODQ6)

## Original project
Fork of the opensource project by [@jub0bs](https://github.com/jub0bs/), available on GitHub at [https://github.com/jub0bs/namecheck](https://github.com/jub0bs/namecheck).

## License
Licensed under the MIT License, Version 2.0 (the "License"). You may not use this file except in compliance with the License.
You may obtain a copy of the License [here](https://choosealicense.com/licenses/mit/).

If needed some help,  there are a ["Licenses 101" by FOSSA](https://fossa.com/blog/open-source-licenses-101-mit-license/), a [Snyk explanation](https://snyk.io/learn/what-is-mit-license/)
of MIT license and a [French conference talk](https://www.youtube.com/watch?v=8WwTe0vLhgc) by [Jean-Michael Legait](https://twitter.com/jmlegait) about licenses.

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fdavidaparicio%2Fnamecheck.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fdavidaparicio%2Fnamecheck?ref=badge_large)