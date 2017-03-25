#/bin/bash
curl -XDELETE http://localhost:9200/developerq_article
curl -XDELETE http://localhost:9200/keyword
rm var/master.info
./go-mysql-elasticsearch
