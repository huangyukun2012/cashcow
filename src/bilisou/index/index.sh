#/bin/bash
curl -XDELETE http://localhost:9200/uinfo
curl -XDELETE http://localhost:9200/sharedata
rm var/master.info
./go-mysql-elasticsearch
