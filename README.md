# bs_router

Beanstalkd driven CBSD

# Installation

setenv GOPATH $( realpath . )

go get

go build

pkg update -f

pkg install -y beanstalkd

service beanstalkd enable

service beanstalkd start

./bs_router

