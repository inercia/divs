# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

# Provisioning script
$PROVISION_SCRIPT = <<SCRIPT
  # Install Go
  sudo apt-get update
  sudo apt-get install -y build-essential mercurial git-core golang-go

  # Setup the GOPATH
  sudo mkdir -p /opt/gopath
  sudo chown -R vagrant:vagrant /opt/

  :> /tmp/gopath.sh
  echo 'export GOPATH="/opt/gopath"'                        >> /etc/profile.d/gopath.sh
  echo 'export PATH="/opt/gopath/bin:\$GOPATH/bin:\$PATH"'  >> /etc/profile.d/gopath.sh
  sudo chmod 0755 /etc/profile.d/gopath.sh

SCRIPT

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.provision "shell", inline: $PROVISION_SCRIPT

  ["virtualbox", "vmware_fusion", "vmware_workstation"].each do |p|
    config.vm.provider "p" do |v|
      v.vmx["memsize"] = "2048"
      v.vmx["numvcpus"] = "2"
      v.vmx["cpuid.coresPerSocket"] = "1"
    end
  end

  config.vm.define "64bit" do |n1|
      n1.vm.box = "ubuntu/trusty64"
  end

  config.vm.define "32bit" do |n2|
      n2.vm.box = "ubuntu/trusty32"
  end
end
