Implement server of github.com/gwaylib/log

# Build
```
. env.sh
cd cmd/web
go build # or sup build
```

# Deploy
Install supd(Debian system, others need build from source)
```
wget https://github.com/gwaycc/supd/releases/download/v1.0.4/supd-v1.0.4-linux-amd64.tar.gz
tar -xzf supd-v1.0.4-linux-amd64.tar.gz
cd supd
./setup.sh install

. env.sh
sup build all
sup install all
```

# License

TODO: MIT
