# bs_router

Beanstalkd driven CBSD

# Installation

setenv GOPATH $( realpath . )

go get

go build

pkg update -f

pkg install -y beanstalkd

service beanstalkd enable

sysrc beanstalkd_flags="-l 127.0.0.1"

service beanstalkd start

./bs_router

