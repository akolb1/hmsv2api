#!/usr/bin/env python3

from __future__ import print_function

import argparse
import logging

import grpc

from protobuf.metastore_pb2 import GetDatabaseRequest, Namespace, Id, Cookie, ListDatabasesRequest
from protobuf.metastore_pb2_grpc import MetastoreStub

_default_host = 'localhost'
_default_port = 50051

parser = argparse.ArgumentParser(description='Hive Metastore client')
parser.add_argument('-H', '--host', dest='host', default=_default_host, help='HMS server address')
parser.add_argument('-d', '--db', help='database name')
parser.add_argument('-P', '--port', dest='port', default=_default_port, type=int, help='HMS thrift port')

args = parser.parse_args()

host = args.host
port = str(args.port)
hostport = host+':'+port
print (host, hostport)
channel = grpc.insecure_channel(hostport)
stub = MetastoreStub(channel)

ns = Namespace(name='ns1')
dbid = Id(name='db1', namespace=ns)
cookie = Cookie(cookie='c1')
req = GetDatabaseRequest(id=dbid, cookie=cookie)
db = stub.GetDatabase(req)
print(db)
ldr = ListDatabasesRequest()
dbs = [d for d in stub.ListDatabases(ListDatabasesRequest(namespace=ns, cookie=cookie, name_pattern='*'))]
for db in dbs:
    print(db)


names = []
for db in dbs:
    names.append(db.id.name)

print (names)