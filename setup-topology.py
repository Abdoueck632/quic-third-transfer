#!/usr/bin/env python
from mininet.net import Mininet
from mininet.cli import CLI
from mininet.link import Link, TCLink, Intf
from mininet.node import Controller
from mininet.topo import Topo
if '__main__' == __name__:
        
    net = Mininet(link=TCLink)
    
    c0 = net.addController("controller")

    sw1 = net.addSwitch('sw1')
    sw2 = net.addSwitch('sw2')
    sw3 = net.addSwitch('sw3')
    sw4 = net.addSwitch('sw4')
    
    Server = net.addHost('S') 
    Relay1 = net.addHost('R1')
    Relay2 = net.addHost('R2')
    Client = net.addHost('C')
    router = net.addHost('router')
    linkopt = {'bw': 10}
    linkopt2 = {'bw': 100}
    
    net.addLink(router, sw1, cls=TCLink, **linkopt)
    net.addLink(router, sw2, cls=TCLink, **linkopt)
    net.addLink(router, sw3, cls=TCLink, **linkopt)
    net.addLink(router, sw4, cls=TCLink, **linkopt2)

    net.addLink(sw1, Server, cls=TCLink, **linkopt)
    net.addLink(sw2, Server, cls=TCLink, **linkopt)
    net.addLink(sw3, Relay1, cls=TCLink, **linkopt)
    net.addLink(sw3, Relay2, cls=TCLink, **linkopt)
    net.addLink(sw4, Client, cls=TCLink, **linkopt2)
    net.build()
    router.cmd("ifconfig router-eth0 0")
    router.cmd("ifconfig router-eth1 0")
    router.cmd("ifconfig router-eth2 0")

    Server.cmd("ifconfig S-eth0 0")
    Server.cmd("ifconfig S-eth1 0")
    
    Relay1.cmd("ifconfig R1-eth0 0")
    
    Relay2.cmd("ifconfig R2-eth0 0")

    Client.cmd("ifconfig C-eth0 0")
    
    router.cmd("echo 1 > /proc/sys/net/ipv4/ip_forward")
    router.cmd("ifconfig router-eth0 10.0.0.1 netmask 255.255.255.0")
    router.cmd("ifconfig router-eth1 10.0.1.1 netmask 255.255.255.0")
    router.cmd("ifconfig router-eth2 10.0.2.1 netmask 255.255.255.0")
    router.cmd("ifconfig router-eth3 10.0.3.1 netmask 255.255.255.0")
    
    Server.cmd("ifconfig S-eth0 10.0.0.2 netmask 255.255.255.0")
    Server.cmd("ifconfig S-eth1 10.0.1.2 netmask 255.255.255.0")
    
    Relay1.cmd("ifconfig R1-eth0 10.0.2.2 netmask 255.255.255.0")

    Relay2.cmd("ifconfig R2-eth0 10.0.2.3 netmask 255.255.255.0")
    
    Client.cmd("ifconfig C-eth0 10.0.3.2 netmask 255.255.255.0")
    
    Server.cmd("ip rule add from 10.0.0.2 table 1")
    Server.cmd("ip rule add from 10.0.1.2 table 2")
    Server.cmd("ip route add 10.0.0.0/24 dev S-eth0 scope link table 1")
    Server.cmd("ip route add default via 10.0.0.1 dev S-eth0 table 1")
    Server.cmd("ip route add 10.0.1.0/24 dev S-eth1 scope link table 2")
    Server.cmd("ip route add default via 10.0.1.1 dev S-eth1 table 2")
    Server.cmd("ip route add default scope global nexthop via 10.0.0.1 dev S-eth0")
    
    Relay1.cmd("ip rule add from 10.0.2.2 table 1")
    Relay1.cmd("ip route add 10.0.2.0/24 dev R1-eth0 scope link table 1")
    Relay1.cmd("ip route add default via 10.0.2.1 dev R1-eth0 table 1")
    Relay1.cmd("ip route add default scope global nexthop via 10.0.2.1 dev R1-eth0")

    Relay2.cmd("ip rule add from 10.0.2.3 table 1")
    Relay2.cmd("ip route add 10.0.2.0/24 dev R2-eth0 scope link table 1")
    Relay2.cmd("ip route add default via 10.0.2.1 dev R2-eth0 table 1")
    Relay2.cmd("ip route add default scope global nexthop via 10.0.2.1 dev R2-eth0")
    
    Client.cmd("ip rule add from 10.0.3.2 table 1")
    Client.cmd("ip route add 10.0.3.0/24 dev C-eth0 scope link table 1")
    Client.cmd("ip route add default via 10.0.3.1 dev C-eth0 table 1")
    Client.cmd("ip route add default scope global nexthop via 10.0.3.1 dev C-eth0")
    
    controller = net.controllers[0]
    controller.start()
    
    sw1.start([controller])
    sw2.start([controller])
    sw3.start([controller])
    sw4.start([controller])

    CLI(net)
    net.stop()
