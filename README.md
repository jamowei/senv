# Senv

[![License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)](https://github.com/jamowei/senv/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/jamowei/senv)](https://goreportcard.com/report/github.com/jamowei/senv)
[![Build Status](https://travis-ci.org/jamowei/senv.svg?branch=master)](https://travis-ci.org/jamowei/senv)
[![Coverage Status](https://coveralls.io/repos/github/jamowei/senv/badge.svg?branch=master)](https://coveralls.io/github/jamowei/senv?branch=master)

A fast Spring cloud-config-client written in Go.
 
It fetches properties from a Spring cloud-config-server
and make them available via system environment variables
For more information on spring cloud config take a look [here](https://cloud.spring.io/spring-cloud-config/).

# ToDo's

* https support with own ca
* customizable http header for vault and basic auth
* 100% code coverage
* ...

# Contributing

1. Fork it
2. Download your fork to your PC (`git clone https://github.com/your_username/cobra && cd cobra`)
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Make changes and add them (`git add .`)
5. Commit your changes (`git commit -m 'Add some feature'`)
6. Push to the branch (`git push origin my-new-feature`)
7. Create new pull request

# License

Senv is released under the MIT license. See [LICENSE](https://github.com/jamowei/senv/blob/master/LICENSE)