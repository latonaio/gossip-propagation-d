# -*- mode: ruby -*-
# vi: set ft=ruby :

$num_instances = 4
$vm_cpus = 1
$vm_gui = false
$instance_name_prefix = "vm"
$subnet = "192.168.33"

Vagrant.configure("2") do |config|
    config.vm.box = "bento/ubuntu-18.04"

    (1..$num_instances).each do |i|
        config.vm.define vm_name = "%s-%01d" % [$instance_name_prefix, i] do |node|
            node.vm.hostname = vm_name

            node.vm.provider :virtualbox do |vb|
                vb.cpus = $vm_cpus
                vb.gui = $vm_gui
                vb.memory = 516
                vb.linked_clone = true
                vb.customize ["modifyvm", :id, "--vram", "8"] # ubuntu defaults to 256 MB which is a waste of precious RAM
                vb.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
            end

            ip = "#{$subnet}.#{i+100}"
            node.vm.network :private_network, ip: ip

            # for docker daemon
            node.vm.network "forwarded_port", guest: 2375, host: "#{i+12375}", host_ip: "127.0.0.1"
            # for etcd
            node.vm.network "forwarded_port", guest: 2379, host: "#{i+13379}", host_ip: "127.0.0.1"
            node.vm.network "forwarded_port", guest: 22, host: "#{i+10022}", host_ip: "127.0.0.1"

            node.vm.provision "shell", inline: <<-SHELL
                apt-get update
                apt-get install -y docker.io
                sed -i -e "s/ExecStart/#ExecStart/g" /lib/systemd/system/docker.service
                sed -i -e '14a ExecStart=/usr/bin/dockerd -H tcp://0.0.0.0:2375 -H fd:// --containerd=/run/containerd/containerd.sock' /lib/systemd/system/docker.service
                gpasswd -a vagrant docker
                systemctl daemon-reload
                systemctl start docker
                systemctl enable docker
                timedatectl set-timezone Asia/Tokyo
            SHELL

            node.vm.provision "shell", run: "always", inline: <<-SHELL
                mkdir -p /tmp/etcd.tmp
            SHELL
        end
    end
end