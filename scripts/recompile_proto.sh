cd ../src
export GOPATH=$HOME/go
PATH=$PATH:$GOPATH/bin

protoc --go_out=. --go_opt=paths=source_relative post.proto
protoc --go_out=. --go_opt=paths=source_relative message.proto