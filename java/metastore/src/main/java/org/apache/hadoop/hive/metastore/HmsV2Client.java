/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package org.apache.hadoop.hive.metastore;

import com.akolb.metastore.AddPartitionRequest;
import com.akolb.metastore.AddPartitionResponse;
import com.akolb.metastore.AlterDatabaseRequest;
import com.akolb.metastore.CreateDatabaseRequest;
import com.akolb.metastore.CreateTableRequest;
import com.akolb.metastore.Database;
import com.akolb.metastore.DropDatabaseRequest;
import com.akolb.metastore.DropPartitionsRequest;
import com.akolb.metastore.DropTableRequest;
import com.akolb.metastore.GetDatabaseRequest;
import com.akolb.metastore.GetDatabaseResponse;
import com.akolb.metastore.GetPartitionRequest;
import com.akolb.metastore.GetPartitionResponse;
import com.akolb.metastore.GetTableRequest;
import com.akolb.metastore.GetTableResponse;
import com.akolb.metastore.Id;
import com.akolb.metastore.ListDatabasesRequest;
import com.akolb.metastore.ListPartitionsRequest;
import com.akolb.metastore.ListTablesRequest;
import com.akolb.metastore.MetastoreGrpc;
import com.akolb.metastore.Partition;
import com.akolb.metastore.PartitionValues;
import com.akolb.metastore.RequestStatus;
import com.akolb.metastore.Table;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.StatusRuntimeException;
import org.apache.hadoop.hive.metastore.api.MetaException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Collections;
import java.util.Iterator;
import java.util.List;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

class HmsV2Client {
  private static final Logger LOG = LoggerFactory.getLogger(HmsV2Client.class.getName());
  private final ManagedChannel channel;
  private final MetastoreGrpc.MetastoreBlockingStub blockingStub;

  /**
   * Construct client connecting to HelloWorld server at {@code host:port}.
   */
  HmsV2Client(String host, int port) {
    this(ManagedChannelBuilder.forAddress(host, port)
        // Channels are secure by default (via SSL/TLS). For the example we disable TLS to avoid
        // needing certificates.
        .usePlaintext(true)
        .build());
  }

  /**
   * Construct client for accessing RouteGuide server using the existing channel.
   */
  HmsV2Client(ManagedChannel channel) {
    this.channel = channel;
    blockingStub = MetastoreGrpc.newBlockingStub(channel);
  }

  public void shutdown() throws InterruptedException {
    channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
  }

  void createDatabase(String catalog, Database database) throws MetaException {
    LOG.info("createDatabase({}, {})", catalog, database.getId().getName());
    CreateDatabaseRequest request = CreateDatabaseRequest
        .newBuilder()
        .setCatalog(catalog)
        .setDatabase(database)
        .build();
    GetDatabaseResponse response;
    try {
      response = blockingStub.createDabatase(request);
    } catch (StatusRuntimeException e) {
      LOG.error("failed to create database", e);
      throw new MetaException("failed to create database: " + e.getMessage());
    }
    if (response.getStatus().getStatus() != RequestStatus.Status.STATUS_OK) {
      throw new MetaException(response.getStatus().getError());
    }
  }

  /**
   * Get database parameters.
   */
  Database getDatabase(String catalog, String name) throws MetaException {
    LOG.info("getDatabase({}, {})", catalog, name);
    GetDatabaseRequest req = GetDatabaseRequest
        .newBuilder()
        .setCatalog(catalog)
        .setId(Id.newBuilder().setName(name).build())
        .build();
    GetDatabaseResponse response;
    try {
      response = blockingStub.getDatabase(req);
    } catch (StatusRuntimeException e) {
      LOG.error("failed to get database", e);
      throw new MetaException("failed to get database: " + e.getMessage());
    }
    if (response.getStatus().getStatus() != RequestStatus.Status.STATUS_OK) {
      throw new MetaException(response.getStatus().getError());
    }
    return response.getDatabase();
  }

  /**
   * Get all databases
   */
  Iterator<Database> listDatabases(String catalog,
                                   String pattern,
                                   List<String> fields) throws MetaException {
    LOG.info("listDatabases({}, {})", catalog, pattern);
    ListDatabasesRequest req = ListDatabasesRequest.newBuilder()
        .setCatalog(catalog)
        .setNamePattern(pattern)
        .addAllFields(fields)
        .build();
    try {
      return blockingStub.listDatabases(req);
    } catch (StatusRuntimeException e) {
      LOG.error("failed to create database", e);
      throw new MetaException("failed to get database: " + e.getMessage());
    }
  }

  void dropDatabase(String catalog, String name) throws MetaException {
    LOG.info("dropDatabase({}, {})", catalog, name);
    Id id = Id.newBuilder().setName(name).build();
    DropDatabaseRequest req = DropDatabaseRequest.newBuilder().setCatalog(catalog).setId(id).build();
    RequestStatus response;
    try {
      response = blockingStub.dropDatabase(req);
    } catch (StatusRuntimeException e) {
      LOG.error("failed to drop database", e);
      throw new MetaException("failed to drop database: " + e.getMessage());
    }
    if (response.getStatus() != RequestStatus.Status.STATUS_OK) {
      throw new MetaException(response.getError());
    }
  }

  void createTable(String catalog, String dbName, Table table) throws MetaException {
    LOG.info("createTable({}.{})", dbName, table.getId().getName());
    CreateTableRequest request = CreateTableRequest
        .newBuilder()
        .setCatalog(catalog)
        .setDbId(Id.newBuilder().setName(dbName).build())
        .setTable(table)
        .build();
    GetTableResponse response;
    try {
      response = blockingStub.createTable(request);
    } catch (StatusRuntimeException e) {
      LOG.error("failed to create table", e);
      throw new MetaException("failed to create table: " + e.getMessage());
    }
    if (response.getStatus().getStatus() != RequestStatus.Status.STATUS_OK) {
      throw new MetaException(response.getStatus().getError());
    }
  }

  Table getTable(String catalog, String dbName, String tableName) throws MetaException {
    LOG.info("getTable({}.{})", dbName, tableName);
    GetTableRequest request = GetTableRequest.newBuilder()
        .setCatalog(catalog)
        .setId(Id.newBuilder().setName(tableName).build())
        .setDbId(Id.newBuilder().setName(dbName).build())
        .build();
    GetTableResponse response;
    try {
      response = blockingStub.getTable(request);
    } catch (StatusRuntimeException e) {
      LOG.error("failed to get table", e);
      throw new MetaException("failed to create table: " + e.getMessage());
    }
    if (response.getStatus().getStatus() != RequestStatus.Status.STATUS_OK) {
      throw new MetaException(response.getStatus().getError());
    }
    return response.getTable();
  }

  Iterator<Table> listTables(String catalog, String dbName, List<String> fields)
      throws MetaException {
    LOG.info("getTables({})", dbName);
    ListTablesRequest request = ListTablesRequest
        .newBuilder()
        .setCatalog(catalog)
        .setDbId(Id.newBuilder().setName(dbName).build())
        .addAllFields(fields == null ? Collections.emptyList() : fields)
        .build();

    try {
      return blockingStub.listTables(request);
    } catch (StatusRuntimeException e) {
      LOG.error("failed to list tables", e);
      throw new MetaException("failed list tables: " + e.getMessage());
    }
  }

  void dropTable(String catalog, String dbName, String tableName) throws MetaException {
    LOG.info("dropTable({}.{})", dbName, tableName);
    DropTableRequest req = DropTableRequest
        .newBuilder()
        .setCatalog(catalog)
        .setId(Id.newBuilder().setName(tableName).build())
        .setDbId(Id.newBuilder().setName(dbName).build())
        .build();
    RequestStatus response;
    try {
      response = blockingStub.dropTable(req);
    } catch (StatusRuntimeException e) {
      LOG.error("failed to drop table", e);
      throw new MetaException("failed to drop table: " + e.getMessage());
    }
    if (response.getStatus() != RequestStatus.Status.STATUS_OK) {
      throw new MetaException(response.getError());
    }
  }

  void addPartition(String catalog, String dbName, String tableName, Partition partition)
      throws MetaException {
    AddPartitionRequest req = AddPartitionRequest
        .newBuilder()
        .setCatalog(catalog)
        .setDbId(Id.newBuilder().setName(dbName).build())
        .setTableId(Id.newBuilder().setName(tableName).build())
        .setPartition(partition)
        .build();
    AddPartitionResponse response;
    try {
      response = blockingStub.addPartition(req);
    } catch (StatusRuntimeException e) {
      LOG.error("failed to add partition", e);
      throw new MetaException("failed to add partition: " + e.getMessage());
    }
    if (response.getStatus().getStatus() != RequestStatus.Status.STATUS_OK) {
      throw new MetaException(response.getStatus().getError());
    }
  }

  Partition getPartition(String catalog, String dbName,
                         String tableName, List<String> values) throws MetaException {
    GetPartitionRequest req = GetPartitionRequest
        .newBuilder()
        .setCatalog(catalog)
        .setDbId(Id.newBuilder().setName(dbName).build())
        .setTableId(Id.newBuilder().setName(tableName).build())
        .addAllValues(values)
        .build();
    GetPartitionResponse response;
    try {
      response = blockingStub.getPartition(req);
    } catch (StatusRuntimeException e) {
      LOG.error("failed to get partition", e);
      throw new MetaException("failed get partitions " + e.getMessage());
    }
    if (response.getStatus().getStatus() != RequestStatus.Status.STATUS_OK) {
      throw new MetaException(response.getStatus().getError());
    }
    return response.getPartition();
  }

  Iterator<Partition> listPartitions(String catalog, String dbName,
                                     String tableName, List<String> fields)
      throws MetaException {
    ListPartitionsRequest req = ListPartitionsRequest
        .newBuilder()
        .setCatalog(catalog)
        .setDbId(Id.newBuilder().setName(dbName).build())
        .setTableId(Id.newBuilder().setName(tableName).build())
        .addAllFields(fields == null ? Collections.emptyList() : fields)
        .build();
    try {
      return blockingStub.listPartitions(req);
    } catch (StatusRuntimeException e) {
      LOG.error("failed to list partitions", e);
      throw new MetaException("failed list partitions: " + e.getMessage());
    }
  }

  void dropPartitions(String catalog, String dbName, String tableName, List<List<String>> values)
      throws MetaException {
    List<PartitionValues> valuesList =
        values.stream()
            .map(v -> PartitionValues
                .newBuilder()
                .addAllValue(v)
                .build())
            .collect(Collectors.toList());

    DropPartitionsRequest req = DropPartitionsRequest
        .newBuilder()
        .setCatalog(catalog)
        .setDbId(Id.newBuilder().setName(dbName).build())
        .setTableId(Id.newBuilder().setName(tableName).build())
        .addAllValues(valuesList)
        .build();
    RequestStatus response;
    try {
      response = blockingStub.dropPartitions(req);
    } catch (StatusRuntimeException e) {
      LOG.error("failed to drop table", e);
      throw new MetaException("failed to drop partition: " + e.getMessage());
    }
    if (response.getStatus() != RequestStatus.Status.STATUS_OK) {
      throw new MetaException(response.getError());
    }
  }

  Database alterDatabase(String catalog, String dbName, Database db) throws MetaException {
    AlterDatabaseRequest req = AlterDatabaseRequest
        .newBuilder()
        .setCatalog(catalog)
        .setId(Id.newBuilder().setName(dbName).build())
        .setDatabase(db)
        .build();
    GetDatabaseResponse response;
    try {
      response = blockingStub.alterDatabase(req);
    } catch (StatusRuntimeException e) {
      LOG.error("failed to alter table", e);
      throw new MetaException("failed to alter database: " + e.getMessage());
    }

    if (response.getStatus().getStatus() != RequestStatus.Status.STATUS_OK) {
      throw new MetaException(response.getStatus().getError());
    }
    return response.getDatabase();
  }
}
