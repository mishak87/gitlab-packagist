# GitLab Packagist

Simple no dependency in memory cached packages.json server.

## Install

`go get github.com/mishak87/gitlab-packagist`

## Launch

```
Usage of gitlab-packagist:
  -addr=":7070": HTTP server address
  -help=false: Show usage
  -interval=5: Interval between updates
  -token="": GitLab API token
  -url="": GitLab API url including version string ie. https://gitlab.com/api/v3/
  -verify-ssl=true: GitLab API verify SSL
```

## Consume

`http://localhost:7070/packages.json`


## Docker

### Parameters (ENV)

* `URL` `https://gitlab.com/api/v3/`
* `TOKEN`

* `INTERVAL` `5` (minutes)
* `PORT` `7070`

* `VERIFY_SSL` `1` (bool) 

### GitLab in Local Container

```sh
docker run --rm --name gitlab-packagist \
        -e 'URL=https://gitlab/api/v3/' \
        -e 'TOKEN=BOOBIES' \
        -p 7070:7070 \
        --link gitlab:gitlab \
        mishak/gitlab-packagist
```

### GitLab on The Internets

```sh
docker run --rm --name gitlab-packagist \
        -e 'URL=https://gitlab.com/api/v3/' \
        -e 'TOKEN=KITTENS' \
        -p 7070:7070 \
        mishak/gitlab-packagist
```
