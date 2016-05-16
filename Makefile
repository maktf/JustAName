all:
	export GOPATH=//home/yuanhui/Downloads/JustAName-dev && go install kademlia && ./bin/kademlia localhost:7890 localhost:7890
test:
	export GOPATH=//home/yuanhui/Downloads/JustAName-dev && go test libkademlia
