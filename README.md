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
   https://github.com/Abdoueck632/quic-third-transfer.git
   ```
3. Install NPM packages
   ```sh
   npm install
   ```



<!-- USAGE EXAMPLES -->
## Usage

first of all it will be necessary to start the Mininet virtual machine by executing the program setup-topologiy.py containing our basic topology
  ```sh
   sudo python3 setup-topology.py
 ```
Once our topology is well set up, we will have to turn on our different machines (Server, client and relays)

  ```sh
   xterm S C R1 R2
 ```
 Apres nous pouvons verifier notre syteme de transfert de fichier sur nos differentes machines
 
 * On the server host
  ```sh
   go run server.go
 ```
  * On the relay host
  ```sh
   go run relay.go ./
 ```
  * On the client host
  ```sh
   go run server.go ./ test.pdf
 ```
 
