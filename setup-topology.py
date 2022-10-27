#!/usr/bin/env python
from mininet.net import Mininet
from mininet.cli import CLI
from mininet.link import Link, TCLink, Intf
from subprocess import Popen, PIPE
from mininet.log import setLogLevel
from mininet.node import Controller
from mininet.topo import Topo
if '__main__' == __name__:
    setLogLevel('info')
    net = Mininet(link=TCLink)
    key = "net.mptcp.mptcp_enabled"
    value = 1
    p = Popen("sysctl -w %s=%s" % (key, value),
            shell=True, stdout=PIPE, stderr=PIPE)
    stdout, stderr = p.communicate()
    print ("stdout=", stdout, "stderr=", stderr)
    
    
    c0 = net.addController("controller")

    s1 = net.addSwitch('s1')
    s2 = net.addSwitch('s2')
    s3 = net.addSwitch('s3')
    s4 = net.addSwitch('s4')
    
    h1 = net.addHost('S1')
    h2 = net.addHost('R1')
    h4 = net.addHost('R2')
    h3 = net.addHost('C1')
    r1 = net.addHost('r1')
    linkopt = {'bw': 10}
    linkopt2 = {'bw': 100}
    
    net.addLink(r1, s1, cls=TCLink, **linkopt)
    net.addLink(r1, s2, cls=TCLink, **linkopt)
    net.addLink(r1, s3, cls=TCLink, **linkopt)
    net.addLink(r1, s4, cls=TCLink, **linkopt2)

    net.addLink(s1, h1, cls=TCLink, **linkopt)
    net.addLink(s2, h1, cls=TCLink, **linkopt)
    net.addLink(s3, h2, cls=TCLink, **linkopt)
    net.addLink(s3, h4, cls=TCLink, **linkopt)
    net.addLink(s4, h3, cls=TCLink, **linkopt2)
    net.build()
    r1.cmd("ifconfig r1-eth0 0")
    r1.cmd("ifconfig r1-eth1 0")
    r1.cmd("ifconfig r1-eth2 0")

    h1.cmd("ifconfig S1-eth0 0")
    h1.cmd("ifconfig S1-eth1 0")
    
    h2.cmd("ifconfig R1-eth0 0")
    
    h4.cmd("ifconfig R2-eth0 0")

    h3.cmd("ifconfig C1-eth0 0")
    
    r1.cmd("echo 1 > /proc/sys/net/ipv4/ip_forward")
    r1.cmd("ifconfig r1-eth0 10.0.0.1 netmask 255.255.255.0")
    r1.cmd("ifconfig r1-eth1 10.0.1.1 netmask 255.255.255.0")
    r1.cmd("ifconfig r1-eth2 10.0.2.1 netmask 255.255.255.0")
    r1.cmd("ifconfig r1-eth3 10.0.3.1 netmask 255.255.255.0")
    
    h1.cmd("ifconfig S1-eth0 10.0.0.2 netmask 255.255.255.0")
    h1.cmd("ifconfig S1-eth1 10.0.1.2 netmask 255.255.255.0")
    
    h2.cmd("ifconfig R1-eth0 10.0.2.2 netmask 255.255.255.0")

    h4.cmd("ifconfig R2-eth0 10.0.2.3 netmask 255.255.255.0")
    
    h3.cmd("ifconfig C1-eth0 10.0.3.2 netmask 255.255.255.0")
    
    h1.cmd("ip rule add from 10.0.0.2 table 1")
    h1.cmd("ip rule add from 10.0.1.2 table 2")
    h1.cmd("ip route add 10.0.0.0/24 dev S1-eth0 scope link table 1")
    h1.cmd("ip route add default via 10.0.0.1 dev S1-eth0 table 1")
    h1.cmd("ip route add 10.0.1.0/24 dev S1-eth1 scope link table 2")
    h1.cmd("ip route add default via 10.0.1.1 dev S1-eth1 table 2")
    h1.cmd("ip route add default scope global nexthop via 10.0.0.1 dev S1-eth0")
    
    h2.cmd("ip rule add from 10.0.2.2 table 1")
    h2.cmd("ip route add 10.0.2.0/24 dev R1-eth0 scope link table 1")
    h2.cmd("ip route add default via 10.0.2.1 dev R1-eth0 table 1")
    h2.cmd("ip route add default scope global nexthop via 10.0.2.1 dev R1-eth0")

    h4.cmd("ip rule add from 10.0.2.3 table 1")
    h4.cmd("ip route add 10.0.2.0/24 dev R2-eth0 scope link table 1")
    h4.cmd("ip route add default via 10.0.2.1 dev R2-eth0 table 1")
    h4.cmd("ip route add default scope global nexthop via 10.0.2.1 dev R2-eth0")
    
    h3.cmd("ip rule add from 10.0.3.2 table 1")
    h3.cmd("ip route add 10.0.3.0/24 dev C1-eth0 scope link table 1")
    h3.cmd("ip route add default via 10.0.3.1 dev C1-eth0 table 1")
    h3.cmd("ip route add default scope global nexthop via 10.0.3.1 dev C1-eth0")
    
    controller = net.controllers[0]
    controller.start()
    
    s1.start([controller])
    s2.start([controller])
    s3.start([controller])
    s4.start([controller])

    CLI(net)
    net.stop()
