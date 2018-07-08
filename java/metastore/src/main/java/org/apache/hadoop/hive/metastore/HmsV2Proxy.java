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

// TODO: Do not store partition's field schema
// TODO: Implement dropPartitions

package org.apache.hadoop.hive.metastore;

import com.akolb.metastore.Id;
import com.akolb.metastore.InputFormat;
import com.akolb.metastore.OutputFormat;
import com.akolb.metastore.SerdeType;
import com.akolb.metastore.SerializationLib;
import org.apache.hadoop.hive.metastore.api.Database;
import org.apache.hadoop.hive.metastore.api.FieldSchema;
import org.apache.hadoop.hive.metastore.api.MetaException;
import org.apache.hadoop.hive.metastore.api.Order;
import org.apache.hadoop.hive.metastore.api.Partition;
import org.apache.hadoop.hive.metastore.api.PrincipalType;
import org.apache.hadoop.hive.metastore.api.SerDeInfo;
import org.apache.hadoop.hive.metastore.api.SkewedInfo;
import org.apache.hadoop.hive.metastore.api.StorageDescriptor;
import org.apache.hadoop.hive.metastore.api.Table;
import org.apache.hadoop.hive.metastore.model.MDatabase;
import org.apache.hadoop.hive.metastore.model.MTable;

import java.util.AbstractList;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import static com.akolb.metastore.InputFormat.IF_CUSTOM;
import static com.akolb.metastore.InputFormat.IF_PARQUET;
import static com.akolb.metastore.OutputFormat.OF_CUSTOM;
import static com.akolb.metastore.OutputFormat.OF_IGNORE_KEY;
import static com.akolb.metastore.OutputFormat.OF_PARQUET;
import static com.akolb.metastore.OutputFormat.OF_SEQUENCE;
import static com.akolb.metastore.SerializationLib.SL_CUSTOM;
import static com.akolb.metastore.SerializationLib.SL_PARQUET;
import static org.apache.hadoop.hive.metastore.HmsV2Proxy.Thrift2Grpc.DESCRIPTION_KEY;
import static org.apache.hadoop.hive.metastore.HmsV2Proxy.Thrift2Grpc.OWNER_KEY;
import static org.apache.hadoop.hive.metastore.HmsV2Proxy.Thrift2Grpc.tableToThrift;
import static org.apache.hadoop.hive.metastore.TableType.EXTERNAL_TABLE;
import static org.apache.hadoop.hive.metastore.TableType.INDEX_TABLE;
import static org.apache.hadoop.hive.metastore.TableType.MANAGED_TABLE;

// TODO: Normaalize db and table names

final class HmsV2Proxy {
  // Filter to return name of objects only
  private static final List<String> NAME_FIELDS = Collections.singletonList("id.name");

  private final HmsV2Client v2client;
  private final String v2namespace;

  HmsV2Proxy(HmsV2Client v2client, String v2namespace) {
    this.v2client = v2client;
    this.v2namespace = v2namespace;
  }

  /**
   * Create database using V2 API
   *
   * @param db
   */
  void createDatabase(Database db) throws MetaException {
    try {
      v2client.createDatabase(v2namespace, Thrift2Grpc.dbFromThrift(db));
    } catch (MetaException e) {
      throw e;
    }
  }

  Database getDatabase(String dbName) {
    System.out.printf("Get database %s\n", dbName);
    try {
      return Thrift2Grpc.dbToThrift(v2client.getDatabase(v2namespace, dbName));
    } catch (MetaException e) {
      e.printStackTrace();
      return null;
    }
  }

  MDatabase getMDatabase(String dbName) {
    System.out.printf("Get Mdatabase %s\n", dbName);
    try {
      return Model2Grpc.dbModelFromGrpc(v2client.getDatabase(v2namespace, dbName));
    } catch (MetaException ignored) {
      return null;
    }
  }

  List<String> getDatabases(String pattern) {
    System.out.printf("Get database %s\n", pattern);
    try {
      return Thrift2Grpc.dbListToStringList(v2client.listDatabases(v2namespace, pattern, NAME_FIELDS));
    } catch (MetaException ignored) {
      return null;
    }
  }

  void dropDatabase(String dbName) {
    System.out.printf("Drop database %s\n", dbName);
    try {
      v2client.dropDatabase(v2namespace, dbName);
    } catch (MetaException e) {
      e.printStackTrace();
    }
  }

  void createTable(Table tbl) {
    System.out.printf("create table %s.%s\n", tbl.getDbName(), tbl.getTableName());
    try {
      v2client.createTable(v2namespace, tbl.getDbName(), Thrift2Grpc.tableFromThrift(tbl));
    } catch (MetaException e) {
      e.printStackTrace();
    }
  }

  Table getTable(String dbName, String tableName) {
    System.out.printf("get table %s.%s\n", dbName, tableName);
    try {
      return Thrift2Grpc.tableToThrift(dbName, v2client.getTable(v2namespace, dbName, tableName));
    } catch (MetaException e) {
      e.printStackTrace();
    }
    return null;
  }

  List<String> getTables(String dbName) {
    System.out.printf("Get Tables %s\n", dbName);
    try {
      return Thrift2Grpc.tableListToStringList(v2client.listTables(v2namespace, dbName, NAME_FIELDS));
    } catch (MetaException e) {
      e.printStackTrace();
      return null;
    }
  }

  void dropTable(String dbName, String tableName) {
    System.out.printf("Drop table %s.%s\n", dbName, tableName);
    try {
      v2client.dropTable(v2namespace, dbName, tableName);
    } catch (MetaException e) {
      e.printStackTrace();
    }
  }

  void addPartition(Partition partition) {
    try {
      v2client.addPartition(v2namespace, partition.getDbName(),
          partition.getTableName(), Thrift2Grpc.partitionFromThrift(partition));
    } catch (MetaException e) {
      e.printStackTrace();
    }
  }

  void addPartitions(String dbName, String tblName, List<Partition> partitions) {
    for (Partition p : partitions) {
      addPartition(p);
    }
  }

  Partition getPartition(String dbName, String tableName, List<String> values) {
    try {
      com.akolb.metastore.Partition partition =
          v2client.getPartition(v2namespace, dbName, tableName, values);
      if (partition == null) {
        System.out.println("no such partition " + values);
        return null;
      }
      com.akolb.metastore.Table gTable = partition.getTable();
      if (gTable == null) {
        System.out.println("missing table info from partition " + partition);
      }
      Table table = tableToThrift(dbName, gTable);
      return Thrift2Grpc.partitionToThrift(table,
          v2client.getPartition(v2namespace, dbName, tableName, values));
    } catch (MetaException e) {
      e.printStackTrace();
    }
    return null;
  }

  List<Partition> getPartitions(String dbName, String tableName) {
    try {
      Iterator<com.akolb.metastore.Partition> partitions =
          v2client.listPartitions(v2namespace, dbName, tableName, null);
      if (!partitions.hasNext()) {
        return Collections.emptyList();
      }
      com.akolb.metastore.Partition first = partitions.next();
      final Table table = tableToThrift(dbName, first.getTable());
      if (table == null) {
        System.out.println("Missing table info from partition list");
        return Collections.emptyList();
      }
      System.out.println("getPartitions(): got table " + table);
      List<Partition> result = new ArrayList<>();
      result.add(Thrift2Grpc.partitionToThrift(table, first));

      partitions.forEachRemaining(p -> result.add(Thrift2Grpc.partitionToThrift(table, p)));
      return result;
    } catch (MetaException e) {
      e.printStackTrace();
      return Collections.emptyList();
    }
  }

  public List<String> listPartitionNames(String dbName, String tableName) {
    try {
      Iterator<com.akolb.metastore.Partition> partitions =
          v2client.listPartitions(v2namespace, dbName, tableName,
              Arrays.asList("values", "table"));
      if (!partitions.hasNext()) {
        return Collections.emptyList();
      }
      com.akolb.metastore.Partition first = partitions.next();
      List<String> result = new ArrayList<>();
      com.akolb.metastore.Table gTable = first.getTable();
      if (gTable == null) {
        System.out.println("Missing table info from partition list");
        return Collections.emptyList();
      }
      final Table table = tableToThrift(dbName, gTable);
      result.add(Warehouse.makePartName(table.getPartitionKeys(), first.getValuesList()));
      partitions.forEachRemaining(p -> {
        try {
          result.add(Warehouse.makePartName(table.getPartitionKeys(), p.getValuesList()));
        } catch (MetaException e) {
          e.printStackTrace();
        }
      });
      return result;
    } catch (MetaException e) {
      e.printStackTrace();
      return null;
    }
  }

  void dropPartition(String dbName, String tableName, List<String> values) {
    try {
      v2client.dropPartitions(v2namespace, dbName, tableName, Collections.singletonList(values));
    } catch (MetaException e) {
      e.printStackTrace();
    }
  }

  void dropPartitions(String dbName, String tableName, List<String> names) {
    List<List<String>> partitionVaues = new ArrayList<>();
    try {
      for (String name : names) {
        AbstractList<String> values = null;
        values = Warehouse.makeValsFromName(name, values);
        partitionVaues.add(values);
      }
      v2client.dropPartitions(v2namespace, dbName, tableName, partitionVaues);
    } catch (MetaException e) {
      e.printStackTrace();
    }
  }

  void alterDatabase(String dbName, Database db) throws MetaException {
    v2client.alterDatabase(v2namespace, dbName, Thrift2Grpc.dbFromThrift(db));
  }

  // Conversion between GRPC and Thrift schemas
  static final class Thrift2Grpc {
    static final String ID_KEY = "Id";
    static final String SEQ_ID_KEY = "SeqId";
    static final String OWNER_KEY = "Owner";
    static final String OWNER_TYPE_KEY = "OwnerType";
    static final String DESCRIPTION_KEY = "Description";
    static final String CREATE_TIME = "createTime";
    static final String EXTERNAL = "EXTERNAL";
    static final String TRUE_VAL = "true";

    private Thrift2Grpc() {
    }

    static Database dbToThrift(com.akolb.metastore.Database gDb) {
      Database db = new Database();
      Map<String, String> params = gDb.getParametersMap();
      String ownerType = gDb.getParametersOrDefault(OWNER_TYPE_KEY, (PrincipalType.USER.toString()));
      try {
        db.setOwnerType(PrincipalType.findByValue(Integer.valueOf(ownerType)));
      } catch (NumberFormatException ignored) {
        db.setOwnerType(PrincipalType.USER);
      }


      db.setName(gDb.getId().getName());

      String locationURI = gDb.getLocation();
      db.setLocationUri(locationURI);
      db.setDescription(gDb.getSystemParametersOrDefault(DESCRIPTION_KEY, ""));
      db.setOwnerName(gDb.getSystemParametersOrDefault(OWNER_KEY, ""));
      Map<String, String> parameters = new HashMap<>(params);
      parameters.put(ID_KEY, gDb.getId().getId());
      parameters.put(SEQ_ID_KEY, String.valueOf(gDb.getSeqId()));
      db.setParameters(parameters);

      return db;
    }

    static com.akolb.metastore.Database dbFromThrift(Database db) {
      Map<String, String> parameters = new HashMap<>();
      if (db.isSetDescription()) {
        parameters.put(DESCRIPTION_KEY, db.getDescription());
      }
      if (db.isSetOwnerName()) {
        parameters.put(OWNER_KEY, db.getOwnerName());
      }
      if (db.isSetOwnerType()) {
        parameters.put(OWNER_TYPE_KEY, String.valueOf(db.getOwnerType().getValue()));
      } else {
        parameters.put(OWNER_TYPE_KEY, PrincipalType.USER.name());
      }
      com.akolb.metastore.Database.Builder dbBuilder = com.akolb.metastore.Database.newBuilder();
      dbBuilder.putAllSystemParameters(parameters);
      if (db.isSetParameters()) {
        dbBuilder.putAllParameters(db.getParameters());
      }
      dbBuilder.setId(Id.newBuilder().setName(db.getName()))
          .setLocation(db.getLocationUri())
          .build();

      return dbBuilder.build();
    }

    static List<String> dbListToStringList(Iterator<com.akolb.metastore.Database> dbList) {
      List<String> dbNames = new ArrayList<>();
      dbList.forEachRemaining(db -> dbNames.add(db.getId().getName()));
      return dbNames.stream().sorted().collect(Collectors.toList());
    }

    static List<String> tableListToStringList(Iterator<com.akolb.metastore.Table> tables) {
      List<String> names = new ArrayList<>();
      tables.forEachRemaining(t -> names.add(t.getId().getName()));
      return names.stream().sorted().collect(Collectors.toList());
    }

    static com.akolb.metastore.Table tableFromThrift(Table table) {
      Map<String, String> parameters = new HashMap<>();
      if (table.isSetOwner()) {
        parameters.put(OWNER_KEY, table.getOwner());
      }
      com.akolb.metastore.Table.Builder tableBuilder = com.akolb.metastore.Table.newBuilder();
      if (table.isSetOwner()) {
        tableBuilder.putSystemParameters(OWNER_KEY, table.getOwner());
      }
      if (table.isSetParameters()) {
        tableBuilder.putAllParameters(table.getParameters());
      }
      tableBuilder.setId(Id.newBuilder().setName(table.getTableName()));

      tableBuilder.putSystemParameters(CREATE_TIME, String.valueOf(table.getCreateTime()));
      tableBuilder.setSd(sdFromThrift(table.getSd()));
      tableBuilder.setLocation(table.getSd().getLocation());

      // Copy Partition keys information
      if (table.isSetPartitionKeys()) {
        for (FieldSchema fs : table.getPartitionKeys()) {
          tableBuilder.addPartitionKeys(com.akolb.metastore.FieldSchema
              .newBuilder()
              .setName(fs.getName())
              .setType(fs.getType())
              .setComment(fs.getComment())
              .build());
        }
      }

      boolean isExternal = table.getTableType().equals(EXTERNAL_TABLE.toString()) ||
          Boolean.parseBoolean(table.getParameters().get("EXTERNAL"));
      if (isExternal) {
        tableBuilder.setTableType(com.akolb.metastore.TableType.TTYPE_EXTERNAL);
      } else {
        tableBuilder.setTableType(com.akolb.metastore.TableType.TTYPE_MANAGED);
      }

      // TODO Add partition keys

      return tableBuilder.build();
    }

    static Table tableToThrift(String dbName, com.akolb.metastore.Table gTable) {
      Table table = new Table();
      table.setDbName(dbName);
      // TODO: check for id != null
      table.setTableName(gTable.getId().getName());
      boolean isExternalTable = false;
      switch (gTable.getTableType()) {
        case TTYPE_INDEX:
          table.setTableType(INDEX_TABLE.toString());
          break;
        case TTYPE_EXTERNAL:
          table.setTableType(EXTERNAL_TABLE.toString());
          isExternalTable = true;
          break;
        default:
          table.setTableType(MANAGED_TABLE.toString());
          break;
      }

      table.setOwner(gTable.getSystemParametersOrDefault(OWNER_KEY, ""));
      table.setCreateTime(Integer.valueOf(gTable.getSystemParametersOrDefault(CREATE_TIME,
          "0")));
      table.setSd(sdToThrift(gTable.getSd()));
      table.getSd().setLocation(gTable.getLocation());

      // Copy partition keys if they are present
      if (gTable.getPartitionKeysCount() != 0) {
        table.setPartitionKeys(gTable.getPartitionKeysList()
            .stream()
            .map(fs -> new FieldSchema(fs.getName(), fs.getType(), fs.getComment()))
            .collect(Collectors.toList()));
      } else {
        table.setPartitionKeys(Collections.emptyList());
      }

      Map<String, String> parameters = new HashMap<>(gTable.getParametersMap());
      parameters.put(ID_KEY, gTable.getId().getId());
      parameters.put(SEQ_ID_KEY, String.valueOf(gTable.getSeqId()));
      if (isExternalTable && !parameters.containsKey(EXTERNAL)) {
        parameters.put(EXTERNAL, TRUE_VAL);
      }
      table.setParameters(parameters);

      return table;
    }

    /**
     * Convert gRRPC storage descriptor to Thrift storage descriptor
     *
     * @param gSd
     * @return
     */
    private static StorageDescriptor sdToThrift(com.akolb.metastore.StorageDescriptor gSd) {
      StorageDescriptor sd = new StorageDescriptor();
      // Copy table field schema
      List<FieldSchema> fsList = gSd.getColsList()
          .stream()
          .map(fs -> new FieldSchema(fs.getName(), fs.getType(), fs.getComment()))
          .collect(Collectors.toList());
      sd.setCols(fsList);

      // TODO: deal with format enums
      sd.setOutputFormat(getOutputFormatStr(gSd.getOutputFormat(), gSd.getOutputFormatName()));
      sd.setInputFormat(getInputFormatStr(gSd.getInputFormat(), gSd.getInputFormatName()));
      sd.setNumBuckets(gSd.getNumBuckets());
      // TODO: Implement skewed info
      sd.setSkewedInfo(new SkewedInfo(Collections.emptyList(),
          Collections.emptyList(), Collections.emptyMap()));
      // TODO: implement StoredAsSubDirectories
      sd.setStoredAsSubDirectories(false);
      com.akolb.metastore.SerDeInfo gSdi = gSd.getSerdeInfo();
      if (gSdi != null) {
        SerDeInfo sdi = new SerDeInfo();
        sdi.setName(gSdi.getName());
        sdi.setSerializationLib(gSdi.getSerializationLib());
        if (gSdi.getParametersCount() != 0) {
          sdi.setParameters(gSdi.getParametersMap());
        }
        sd.setSerdeInfo(sdi);
      }
      if (gSd.getBucketColsCount() != 0) {
        sd.setBucketCols(gSd.getBucketColsList());
      } else {
        sd.setBucketCols(Collections.emptyList());
      }

      if (gSd.getSortColsCount() != 0) {
        sd.setSortCols(gSd.getSortColsList()
            .stream()
            .map(sc -> new Order(sc.getCol(), sc.getAscending() ? 1 : 0))
            .collect(Collectors.toList()));
      } else {
        sd.setSortCols(Collections.emptyList());
      }
      sd.setParameters(gSd.getParametersMap());

      return sd;
    }

    static com.akolb.metastore.Partition partitionFromThrift(Partition partition) {
      com.akolb.metastore.Partition.Builder pb = com.akolb.metastore.Partition.newBuilder();
      StorageDescriptor sd = partition.getSd();
      sd.setCols(null);
      pb.setSd(sdFromThrift(sd));
      pb.setLocation(partition.getSd().getLocation());
      if (partition.isSetParameters()) {
        pb.putAllParameters(partition.getParameters());
      }
      pb.addAllValues(partition.getValues());
      return pb.build();
    }

    static Partition partitionToThrift(Table table, com.akolb.metastore.Partition gPartition) {
      Partition partition = new Partition();
      partition.setDbName(table.getDbName());
      partition.setTableName(table.getTableName());
      partition.setSd(sdToThrift(gPartition.getSd()));
      partition.getSd().setCols(table.getSd().getCols());
      partition.getSd().setLocation(gPartition.getLocation());
      partition.setValues(gPartition.getValuesList());
      Map<String, String> parameters = new HashMap<>(gPartition.getParametersMap());
      parameters.put(SEQ_ID_KEY, String.valueOf(gPartition.getSeqId()));
      partition.setParameters(parameters);
      return partition;
    }

    private static com.akolb.metastore.StorageDescriptor sdFromThrift(StorageDescriptor sd) {
      com.akolb.metastore.StorageDescriptor.Builder sdBuilder =
          com.akolb.metastore.StorageDescriptor.newBuilder();

      SerDeInfo sdi = sd.getSerdeInfo();

      com.akolb.metastore.SerDeInfo.Builder si = com.akolb.metastore.SerDeInfo
          .newBuilder()
          .setName(sdi.getName())
          .setSerializationLib(sdi.getSerializationLib());
      if (sdi.isSetParameters()) {
        si.putAllParameters(sdi.getParameters());
      }

      String outputFormat = sd.getOutputFormat();
      OutputFormat of = getOutputFormat(outputFormat);
      String inputFormat = sd.getInputFormat();
      InputFormat iff = getInputFormat(inputFormat);

      sdBuilder
          .setOutputFormat(of)
          .setInputFormat(iff)
          .setNumBuckets(sd.getNumBuckets())
          .setSerdeInfo(si.build());

      if (sd.getBucketCols() != null) {
        sdBuilder.addAllBucketCols(sd.getBucketCols());
      }

      if (of == OF_CUSTOM) {
        sdBuilder.setOutputFormatName(outputFormat);
      }
      if (iff == IF_CUSTOM) {
        sdBuilder.setInputFormatName(inputFormat);
      }

      if (sd.isSetSortCols()) {
        for (Order o : sd.getSortCols()) {
          com.akolb.metastore.Order.Builder ob = com.akolb.metastore.Order.newBuilder();
          sdBuilder.addSortCols(ob.setAscending(o.getOrder() == 1).setCol(o.getCol()).build());
        }
      }

      if (sd.getCols() != null) {
        // Copy field schema for the table
        for (FieldSchema fs : sd.getCols()) {
          com.akolb.metastore.FieldSchema.Builder b =
              com.akolb.metastore.FieldSchema.newBuilder();
          sdBuilder.addCols(com.akolb.metastore.FieldSchema
              .newBuilder()
              .setName(fs.getName())
              .setType(fs.getType())
              .setComment(fs.getComment())
              .build()
          );
        }
      }

      return sdBuilder.build();
    }

    private static OutputFormat getOutputFormat(String format) {
      switch (format) {
        case "org.apache.hadoop.hive.ql.io.HiveIgnoreKeyTextOutputFormat":
          return OF_IGNORE_KEY;
        case "org.apache.hadoop.hive.ql.io.HiveSequenceFileOutputFormat":
          return OF_SEQUENCE;
        case "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat":
          return OF_PARQUET;
        case "org.apache.hadoop.hive.ql.io.HiveNullValueSequenceFileOutputFormat":
        case "org.apache.hadoop.hive.ql.io.HivePassThroughOutputFormat":
        case "org.apache.hadoop.hive.ql.io.IgnoreKeyTextOutputFormat":
        case "org.apache.hadoop.hive.ql.io.HiveBinaryOutputFormat":
        case "org.apache.hadoop.hive.ql.io.RCFileOutputFormat":
        default:
          return OF_CUSTOM;
      }
    }

    private static String getOutputFormatStr(OutputFormat format, String val) {
      switch (format) {
        case OF_IGNORE_KEY:
          return "org.apache.hadoop.hive.ql.io.HiveIgnoreKeyTextOutputFormat";
        case OF_SEQUENCE:
          return "org.apache.hadoop.hive.ql.io.HiveSequenceFileOutputFormat";
        case OF_PARQUET:
          return "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat";
        default:
          return val;
      }
    }

    private static InputFormat getInputFormat(String format) {
      switch (format) {
        case "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat":
          return IF_PARQUET;
        case "org.apache.hadoop.hive.ql.io.HiveNullValueSequenceFileOutputFormat":
        case "org.apache.hadoop.hive.ql.io.HivePassThroughOutputFormat":
        case "org.apache.hadoop.hive.ql.io.IgnoreKeyTextOutputFormat":
        case "org.apache.hadoop.hive.ql.io.HiveBinaryOutputFormat":
        case "org.apache.hadoop.hive.ql.io.RCFileOutputFormat":
        default:
          return IF_CUSTOM;
      }
    }

      private static String getInputFormatStr(InputFormat format, String val) {
        switch (format) {
          case IF_PARQUET:
            return "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat";
          default:
            return val;
        }
      }

      private static SerializationLib getSerdeType(String serdeName) {
        switch (serdeName) {
          case "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe":
            return SL_PARQUET;
          default:
            return SL_CUSTOM;
        }
      }

      private static String getSerdeStr(SerializationLib sl, String val) {
        switch (sl) {
          case SL_PARQUET:
            return "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe";
          default:
            return val;
        }
      }
  }

  // Conversion between Database model and gRPC
  static final class Model2Grpc {

    // Convert gRPC Database object to MDatabase object
    static MDatabase dbModelFromGrpc(com.akolb.metastore.Database gdb) {
      MDatabase mdb = new MDatabase(gdb.getId().getName(),
          gdb.getLocation(), null, gdb.getParametersMap());
      mdb.setDescription(gdb.getSystemParametersOrDefault(DESCRIPTION_KEY, ""));
      mdb.setOwnerName(gdb.getSystemParametersOrDefault(OWNER_KEY, ""));
      return mdb;
    }

    // TODO: Implement getMtable
    static MTable tableModelFromGrpc(com.akolb.metastore.Table gTable) {
      MTable table = new MTable();
      table.setTableName(gTable.getId().getName());
      return table;
    }
  }
}
