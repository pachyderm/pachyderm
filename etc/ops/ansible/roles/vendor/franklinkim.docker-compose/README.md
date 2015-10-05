# Ansible Docker Compose Role

[![Build Status](https://img.shields.io/travis/weareinteractive/ansible-docker.svg)](https://travis-ci.org/weareinteractive/ansible-docker)
[![Galaxy](http://img.shields.io/badge/galaxy-franklinkim.docker-blue.svg)](https://galaxy.ansible.com/list#/roles/3276)
[![GitHub Tags](https://img.shields.io/github/tag/weareinteractive/ansible-docker.svg)](https://github.com/weareinteractive/ansible-docker)
[![GitHub Stars](https://img.shields.io/github/stars/weareinteractive/ansible-docker.svg)](https://github.com/weareinteractive/ansible-docker)

> `docker-compose` is an [ansible](http://www.ansible.com) role which:
>
> * installs docker-compose

## Installation

Using `ansible-galaxy`:

```
$ ansible-galaxy install franklinkim.docker-compose
```

Using `requirements.yml`:

```
- src: franklinkim.docker-compose
```

Using `git`:

```
$ git clone https://github.com/weareinteractive/ansible-docker-compose.git franklinkim.docker-compose
```

## Variables

Here is a list of all the default variables for this role, which are also available in `defaults/main.yml`.

```
# version
docker_compose_version:
```

## Example playbook

```

- hosts: all
  sudo: yes
  roles:
    - franklinkim.docker
    - franklinkim.docker-compose
  vars:
    docker_compose_version: 1.4.0
```

## Testing

```
$ git clone https://github.com/weareinteractive/ansible-docker-compose.git
$ cd ansible-docker-compose
$ vagrant up
```

## Contributing
In lieu of a formal styleguide, take care to maintain the existing coding style. Add unit tests and examples for any new or changed functionality.

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request

## License
Copyright (c) We Are Interactive under the MIT license.
