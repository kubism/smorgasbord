# Smorgasbord

[![Go Documentation](https://img.shields.io/badge/go-doc-blue.svg?style=flat)](https://pkg.go.dev/github.com/kubism/smorgasbord/pkg)
[![Build Status](https://travis-ci.org/kubism/smorgasbord.svg?branch=master)](https://travis-ci.org/kubism/smorgasbord)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubism/smorgasbord)](https://goreportcard.com/report/github.com/kubism/smorgasbord)
[![Coverage Status](https://coveralls.io/repos/github/kubism/smorgasbord/badge.svg?branch=master)](https://coveralls.io/github/kubism/smorgasbord?branch=master)
[![Maintainability](https://api.codeclimate.com/v1/badges/b6fbe93e1b95f6b7f5e3/maintainability)](https://codeclimate.com/github/kubism/smorgasbord/maintainability)

> a range of open sandwiches and delicacies served as hors d'oeuvres or a buffet

Smorgasbord purpose is to ease up the administration of a wireguard-based VPN.  
It creates, stores and distributes client configurations for its users and can
derive server configuration using the provided agent.
Users can self-service their public keys after authenticating via OpenID Connect.
Rather than using a database the public keys and metadata are commited to a
git repository, which is used as storage endpoint.

Smorgasbord primary goal is to provide a minimalistic environment to manage
users across multiple wireguard servers applicable to embedded systems as well
as more complex installments.

![Concept of Smorgasbord](./docs/concept.svg)

## About the name

This project started a late night project and the name was essentially what
came up first after googling "synonym self-service".
It might therefore be subject to change.


