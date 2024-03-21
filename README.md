# quic-third-transfer



<!-- ABOUT THE PROJECT -->
## About The Project
It is a file transfer module with QUIC mutipath.

### Built With

Golang

<!-- GETTING STARTED -->
## Getting Started

You will need to install some tools and dependencies.

### Prerequisites

You need to install mininet to create a realistic virtual network with hosts, switchs,routers,... .
* Install Mininet <a href="http://mininet.org/download/#option-2-native-installation-from-source">Here</a>
 
* Install Golang <a href="https://go.dev/doc/install">Here</a> or use Goland IDE for Golang. download <a href="https://www.jetbrains.com/go/download/#section=linuxhttps://www.jetbrains.com/go/download/#section=linux">here</a>
 
### Installation

_Below is an example of how you can instruct your audience on installing and setting up your app. This template doesn't rely on any external dependencies or services._

1. Clone the repo
   ```sh
    git clone https://github.com/Abdoueck632/quic-third-transfer.git
   ```
2. Or go install for last version golang
   ```sh
    go install https://github.com/Abdoueck632/quic-third-transfer@latest
   ```
3. Download all dependency packages
   ```sh
   go get -t -u ./...
   
   ```


<!-- USAGE EXAMPLES -->
## Usage with the real address of the servers
* Don't forget do run the root mode
 ```sh
  sudo -i 
 ```
* And then we can go to the directory. For me, it's 
 ```sh
  cd go/src/github.com/Abdoueck632/quic-third-transfer
 ```
* We may sometimes need git pull to retrieve the latest version of the project
  
first of all it will be necessary to start the Mininet virtual machine by executing the program setup-topologiy.py containing our basic topology if you use mininet
  ```sh
   sudo python3 setup-topology.py
 ```
Once our topology is well set up, we will have to turn on our different machines (Server, client and relays) if you use mininet

  ```sh
   xterm S C R1 R2
 ```
 After we will check the file transfer system on our different machines
 
 * On the server host 
  ```sh
   go run server.go IPAddressRelay1:4242 
  ```
  * On the relay1 host
  ```sh
   go run relay.go ./storage-server/ IPAddressRelay2:4242 
  ```
  * On the relay2 host
  ```sh
   go run relay.go ./storage-server/
  ```
  * On the client host
  ```sh
   go run server.go ./ test.pdf ./download/ IPAddressServer:4242 
 ```
 ## Demo, we put the IP addresses of the servers
  *Go environment 
  
  ![Capture d’écran du 2023-12-19 03-37-43](https://github.com/Abdoueck632/quic-third-transfer/assets/50526469/6d70f5d4-af85-4214-80cd-c205dc74a767)

  *The commands for starting the transfer
  
  ![Capture d’écran du 2023-12-18 17-26-52](https://github.com/Abdoueck632/quic-third-transfer/assets/50526469/31ffa774-364f-43c7-996a-c04a2ae33f11)

  *The end of the transfer
  
  ![Capture d’écran du 2023-12-19 02-50-30](https://github.com/Abdoueck632/quic-third-transfer/assets/50526469/d88a680c-d31a-49cc-b532-72dac23b82e1)
