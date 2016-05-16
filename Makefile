all:
	export GOPATH=/home/yuanhui/GitHub/JustAName && go install kademlia && ./bin/kademlia localhost:7890 localhost:7890
test:
	export GOPATH=/home/yuanhui/GitHub/JustAName && go test libkademlia