# Senv

[![License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)](https://github.com/jamowei/senv/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/jamowei/senv)](https://goreportcard.com/report/github.com/jamowei/senv)
[![Build Status](https://travis-ci.org/jamowei/senv.svg?branch=master)](https://travis-ci.org/jamowei/senv)
[![Coverage Status](https://coveralls.io/repos/github/jamowei/senv/badge.svg?branch=master)](https://coveralls.io/github/jamowei/senv?branch=master)

A fast Spring cloud-config-client written in Go.
 
It fetches properties from a Spring cloud-config-server
and make them available via system environment variables
For more information on spring cloud config take a look [here](https://cloud.spring.io/spring-cloud-config/).

# Install

You can get the binary with Go:

`> go get -u github.com/jamowei/senv/senv`

or download latest binary release from [here](https://github.com/jamowei/senv/releases).

# Example

Your spring config server is running on http://127.0.0.1:8888/.

your *application.yml* file in config repo is:
```yaml
# general
description: Spring Config Server
user: admin
workdir: /var/work
---
# test environment
spring:
  profiles:
    active: dev
db:
  user: ${user}
  password: test123
---
# production environment
spring:
  profiles:
    active: prod
db:
  user: ${user}
  password: prod123
```

your own application file *myapp.yml* in config repo:
```yaml
file:
  input: ${workdir}/input.txt
  output: ${workdir}/output.txt
```

then you can start your application *myapp* like the following:
* with development settings:
    ```
    > senv -n myapp -p dev \n
      Fetching config from server at: http://127.0.0.1:8888/myapp/dev/master
      Located environment: name="myapp", profiles=[dev], label="master", version=29374923859338549, state=""
    > echo "$DB_USER:$DB_PASSWORD"             // prints: admin:test123
    > myapp -user $DB_USER -pass $DB_PASSWORD -in $FILE_INPUT -out $FILE_OUTPUT
      ...
    ```
* with production settings:
    ```
    > senv -n myapp -p prod \n
      Fetching config from server at: http://127.0.0.1:8888/myapp/prod/master
      Located environment: name="myapp", profiles=[prod], label="master", version=29374923859338549, state=""
    > echo "$DB_USER:$DB_PASSWORD"             // prints: admin:prod123
    > myapp -user $DB_USER -pass $DB_PASSWORD -in $FILE_INPUT -out $FILE_OUTPUT
      ...
    ```
    
# Help

```
  > senv --help 
    A fast spring config client written in Go for recieving properties
    from a spring cloud config server and
    make them available via system environment variables
    
    Usage:
     senv [flags]
    
    Flags:
     -h, --help               help for senv
         --host string        config-server host (default "127.0.0.1")
     -l, --label string       config-repo label to be used (default "master")
     -n, --name string        spring.application.name (default "application")
     -o, --override           overrides existing environment variables
         --port string        config-server port (default "8888")
     -p, --profiles strings   spring.active.profiles (default [default])
     -v, --verbose            show all received properties
```
# ToDo's

* static file support
* vault support
* https support with own ca
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

Senv is released under the MIT license. See [LICENSE.txt](https://github.com/spf13/cobra/blob/master/LICENSE.txt)