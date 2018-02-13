# Kubernetes Credentials

[Homepage](https://manifold.co) |
[Twitter](https://twitter.com/manifoldco) |
[Code of Conduct](./.github/CODE_OF_CONDUCT.md) |
[Contribution Guidelines](./.github/CONTRIBUTING.md)

[![GitHub release](https://img.shields.io/github/tag/manifoldco/kubernetes-credentials.svg?label=latest)](https://github.com/manifoldco/kubernetes-credentials/releases)
[![Build Status](https://travis-ci.org/manifoldco/kubernetes-credentials.svg?branch=master)](https://travis-ci.org/manifoldco/kubernetes-credentials)
[![Go Report Card](https://goreportcard.com/badge/github.com/manifoldco/kubernetes-credentials)](https://goreportcard.com/report/github.com/manifoldco/kubernetes-credentials)
[![License](https://img.shields.io/badge/license-BSD-blue.svg)](./LICENSE)

![Kubernetes Manifold](./banner.png)

Manifold gives you a single account to purchase and manage cloud services from
multiple providers, giving you managed logging, email, MySQL, Postgres,
Memcache, Redis, and more. Manifold also lets you register configurations for
your services external to Manifold's marketplace, giving you a single location
to organize and manage the credentials for your environments.

This package allows you to load [Manifold](https://www.manifold.co/) credentials
into your Kubernetes cluster. These credentials will be stored as a Kubernetes
secrets so you can use them as such in your deployments.

## Usage

We've utilised [Kubernetes' CRD](https://kubernetes.io/docs/concepts/api-extension/custom-resources/)
implementation to allow you to define fine grained control over your
credentials. As with our [Terraform provider](https://github.com/manifoldco/terraform-provider-manifold/),
we allow you to filter out projects, resources and specific credentials you want
to load.

### Getting started

To use this, you'll need [an account with
Manifold]((https://dashboard.manifold.co/register)) and some configuration
stored there. You can provision free or paid plans, or insert and manage your
own configuration values.

You'll also want to [install the Manifold CLI](https://www.manifold.co/cli)
so that you can generate auth tokens.

### Defining credentials

Defining credentials happens through our Custom Resource Definition. We have
two ways of defining how you would like to get credentials.

**Note:** Currently, Kubernetes does not provide a way to validate CRDs. Because
of this, we advise you to double check your definitions and monitor the output
of the controller if you experience issues. Adding validation
[is a work in progress](https://github.com/kubernetes/community/pull/708). Once
this is added to the Kubernetes core, we'll look at providing this as well.

**Note:** The minimum requirement to define a specific credential is its key.
If you provide a name, this name will be used as a key reference in the k8s
secret. A default value can also be provided. If a default value is provided for
a key that does not exist in the Manifold credentials list, this default value
will be used to populate the credential.

#### Project

You can load multiple credentials at once for a specific project, [as described
in this manifest file](_examples/project/manifest.yml).

#### Resource

If you only want to get the credentials from a specific resource, you can do
this [as described in this manifest file](_examples/resource/manifest.yml).

### Referencing the credentials

Once you've set up the controller (see [setting up the controller](#setting-up-the-controller)),
the controller will start looking for the resources defined earlier and write
the values from Manifold to the respective Kubernetes secret. This means that
when a credential changes, the secret will also be updated automatically with
the new value.

By using exsiting Kubernetes secrets, we allow you to use the Manifold
credentials as secrets. We've [provided an example manifest file](_examples/secrets-usage/manifest.yml).

### Defining secret types

Kubernetes allows you to set up different types of secrets, such as Opaque,
Docker Registry, TLS, ….

The Manifold CRD allows you to create Opaque and Docker Registry types. The
Opaque type is the default and is transparant, meaning that all credentials
that are available through your custom resource will be loaded as a secret.

#### Docker Registry

Using the Docker Registry type it's possible to create a secret which will make
it possible to pull images from a private registry. This secret type requires
you to have the following credentials available:

- `DOCKER_USERNAME`
- `DOCKER_EMAIL`
- `DOCKER_PASSWORD`

There is the optional `DOCKER_SERVER` if your registry is anything other than
Docker Hub.

We've provided [an example](_examples/docker-registry/manifest.yml) on how to use the `docker-registry` secret type.

## Installation

### Setting up the Manifold Auth Token to retrieve the credentials

Once the controller is installed, you'll also want to enable access to the
Manifold API. First, you'll need to create a new Auth Token:

```
$ manifold tokens create
```

Once you have the token, you'll want to create a new Kubernetes Secret:

```
$ kubectl create namespace manifold-system
$ kubectl create --namespace=manifold-system secret generic manifold-api-secrets --from-literal=api_token=<AUTH_TOKEN> --from-literal=team=<MANIFOLD_TEAM>
```

**Note:** The team value is optional. If a team is provided in the controller
(see below), only resources that define this team will be picked up and used
to load the credentials. If no team is defined, this is ignored.

### Setting up the controller

First, you'll need to set up the controller. The controller takes care of
monitoring your Resource Definitions and populating the correct Kubernetes
Secrets with Manifold Credentials. Without it, nothing will happen.

```
$ kubectl create -f https://raw.githubusercontent.com/manifoldco/kubernetes-credentials/master/credentials-controller.yml
```

**Note:** You can customise this credentials-controller file. This is a general
purpose Deployment. `MANIFOLD_API_TOKEN` is a required environment variable for
the controller.

#### With RBAC installed

To use RBAC, we'll add additional ClusterRoles to allow managing CRDs and
secrets.

```
$ kubectl create -f https://raw.githubusercontent.com/manifoldco/kubernetes-credentials/master/rbac.yml
```

## Releasing

To release a new version of this package, use the Make target `release`:

```
$ VERSION=0.1.2 make release
```

This will update the Version in `version.go`, commit the changes and set up a
new tag.
