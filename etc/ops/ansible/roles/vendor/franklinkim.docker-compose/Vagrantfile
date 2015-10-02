# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vbguest.no_remote = true
  config.vbguest.auto_update = false

  config.vm.define 'trusty' do |instance|
    instance.vm.box = 'ubuntu/trusty64'
  end

  config.vm.define 'precise' do |instance|
    instance.vm.box = 'ubuntu/precise64'
  end

  # View the documentation for the provider you're using for more
  # information on available options.
  config.vm.provision "ansible" do |ansible|
    ansible.playbook = "test.yml"
    ansible.verbose = 'vv'
    ansible.sudo = true
  end

end
