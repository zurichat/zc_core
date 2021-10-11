ZC_CORE [![Go Reference](https://zuri.chat/b73fbbfa9db45fc3c22e.svg)](https://docs.zuri.chat)
===============

# Zuri Chat Core Api <img align='right' src="https://zuri.chat/b73fbbfa9db45fc3c22e.svg" width='100"'>

See the docs for everything:https://docs.zuri.chat/

![GitHub repo size](https://img.shields.io/github/repo-size/zurichat/zc_core?style=flat-square)

## Description

The Zuri Chat Core (zc_core) is an open source API that serves as the backend and backbone of Zuri Chat - an open source slack clone developed during the HNG8 internship program 

Zuri Chat implements a plug-in system that allows for much more functionality because different plugins can be developed by different creators. These plugins can then rest on the Zuri Chat Core backbone (zc_core) which is written in **Golang**

## Language

zc_core was written in Go because it is a fast, compiled language and it is easily scaled. Concurrency in Go is also ideal for implementing a plugin structure. This makes it easy to build on the project and allows for future expansion.
* [Golang](https://golang.org)

## API Documentation

A list of the endpoints and the functions they implement can be found in the API folder that contains YAML files for each of the functionalities

## Getting Started

This is an example of how you can setup your project locally.
To get a local copy up and running follow these simple example steps.

### Prerequisites

1. **Go 1.16** or  **lastest version** already installed on your local machine.
2. MongoDB

### Installation
1. Run an instance of MongoDB
```bash
$ mongod
```


2. Create and populate a `.env` file with its keys corresponding values as listed in `example.env`
3. Run zc_core from project root directory
```bash
$ cd /path/to/zc_core
$ go run main.go
```
### Testing
1. Lint checks are done with [golangci-lint](https://golangci-lint.run/) - an aggregator of linters
```bash
$ cd /path/to/zc_core
$ golangci-lint run
```
2. Automated unit and integration tests done with golang's builtin [`testing`](https://pkg.go.dev/testing) package. Single test files, test files in a package, 
```bash
WIP
```

## Features
- [x] Blog: Content management package for the CRUD of blog posts and associated services
- [x] Contact: Service to process all user-related inquiries, complaints, and service-related suggestions
- [x] Data: Data management for plugins
- [x] Docs: OpenAPI specification for all resources
- [x] External: Desktop application download, email subscriptions, etc...
- [x] Marketplace: Plugins marketplace
- [x] Messaging: Realtime messaging service with SocketIO
- [x] Organizations: Organization management system
- [x] Plugin: Plugin management system
- [x] Realtime Communication (RTC): Connection authentication, subcription and pulishing service with centrifugo proxy server
- [x] Report: Report generation and viewing service
- [x] Services: Mail and file upload service package
- [x] Users: User management system
- [x] Utils: Utility package of shared helper functions and data types

## Contributing

Pull requests are welcome from the zc_core_main team. Please review the issues created before making a pull request or create an issue if the issue you solve is not already listed.

1. Fork/Clone the Project
2. Create your Feature Branch (`git checkout -b [branch name]`)
3. Commit your Changes (`git commit -m Add some amazing features`)
4. Lint your Changes
4. Push to the Branch (`git push origin [branch name]`)
5. Open a Pull Request
