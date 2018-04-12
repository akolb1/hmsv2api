# Protocol Documentation
<a name="top"/>

## Table of Contents

- [metastore.proto](#metastore.proto)
    - [CreateDatabaseRequest](#metastore.CreateDatabaseRequest)
    - [CreateTableRequest](#metastore.CreateTableRequest)
    - [Database](#metastore.Database)
    - [Database.ParametersEntry](#metastore.Database.ParametersEntry)
    - [Database.SystemParametersEntry](#metastore.Database.SystemParametersEntry)
    - [DropDatabaseRequest](#metastore.DropDatabaseRequest)
    - [DropTableRequest](#metastore.DropTableRequest)
    - [FieldSchema](#metastore.FieldSchema)
    - [GetDatabaseRequest](#metastore.GetDatabaseRequest)
    - [GetDatabaseResponse](#metastore.GetDatabaseResponse)
    - [GetTableRequest](#metastore.GetTableRequest)
    - [GetTableResponse](#metastore.GetTableResponse)
    - [Id](#metastore.Id)
    - [ListDatabasesRequest](#metastore.ListDatabasesRequest)
    - [ListTablesRequest](#metastore.ListTablesRequest)
    - [Order](#metastore.Order)
    - [RequestStatus](#metastore.RequestStatus)
    - [SerDeInfo](#metastore.SerDeInfo)
    - [SerDeInfo.ParametersEntry](#metastore.SerDeInfo.ParametersEntry)
    - [StorageDescriptor](#metastore.StorageDescriptor)
    - [StorageDescriptor.ParametersEntry](#metastore.StorageDescriptor.ParametersEntry)
    - [StorageDescriptor.SystemParametersEntry](#metastore.StorageDescriptor.SystemParametersEntry)
    - [Table](#metastore.Table)
    - [Table.ParametersEntry](#metastore.Table.ParametersEntry)
    - [Table.SystemParametersEntry](#metastore.Table.SystemParametersEntry)
  
    - [InputFormat](#metastore.InputFormat)
    - [OutputFormat](#metastore.OutputFormat)
    - [RequestStatus.Status](#metastore.RequestStatus.Status)
    - [SerdeType](#metastore.SerdeType)
    - [TableType](#metastore.TableType)
  
  
    - [Metastore](#metastore.Metastore)
  

- [Scalar Value Types](#scalar-value-types)



<a name="metastore.proto"/>
<p align="right"><a href="#top">Top</a></p>

## metastore.proto



<a name="metastore.CreateDatabaseRequest"/>

### CreateDatabaseRequest
Create a new database.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| catalog | [string](#string) |  |  |
| database | [Database](#metastore.Database) |  |  |
| cookie | [string](#string) |  |  |






<a name="metastore.CreateTableRequest"/>

### CreateTableRequest
Create a new table.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| catalog | [string](#string) |  |  |
| db_id | [Id](#metastore.Id) |  |  |
| table | [Table](#metastore.Table) |  |  |
| cookie | [string](#string) |  |  |






<a name="metastore.Database"/>

### Database
Database is a container for tables.

Database object has two sets of parameters:
 - User parameters are intended for user and are just transparently passed around
 - System parameters are intended to be used by Hive for its internal purposes
seq_id is a numeric ID which is unique within a catalog. It can be used to track
new databases in the catalog
Original Metastore Database object also had owner information.
These can be represented using system parameters if needed since the current
metastore service does not interpret Owner info.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [Id](#metastore.Id) |  |  |
| seq_id | [uint64](#uint64) |  | Unique sequence ID within calalog |
| location | [string](#string) |  | Default location of database objects |
| parameters | [Database.ParametersEntry](#metastore.Database.ParametersEntry) | repeated | Database user parameters |
| system_parameters | [Database.SystemParametersEntry](#metastore.Database.SystemParametersEntry) | repeated | System parameters |






<a name="metastore.Database.ParametersEntry"/>

### Database.ParametersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="metastore.Database.SystemParametersEntry"/>

### Database.SystemParametersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="metastore.DropDatabaseRequest"/>

### DropDatabaseRequest
Request to drop a database.
Dropping a database also drops all objects contained in the database.
TODO: Add flag to prohibit dropping non-empty databases


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| catalog | [string](#string) |  |  |
| id | [Id](#metastore.Id) |  |  |
| cookie | [string](#string) |  |  |






<a name="metastore.DropTableRequest"/>

### DropTableRequest
Request to drop a table.
Dropping a table also drops all objects contained in the table
TODO: Add flag to prohibit dropping of non-empty table


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| catalog | [string](#string) |  |  |
| db_id | [Id](#metastore.Id) |  |  |
| id | [Id](#metastore.Id) |  |  |
| cookie | [string](#string) |  |  |






<a name="metastore.FieldSchema"/>

### FieldSchema
FieldSchema defines name and type for each column.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | name of the field |
| type | [string](#string) |  | type of the field. |
| comment | [string](#string) |  | User description |






<a name="metastore.GetDatabaseRequest"/>

### GetDatabaseRequest
Request to get database by its ID.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| catalog | [string](#string) |  |  |
| id | [Id](#metastore.Id) |  |  |
| cookie | [string](#string) |  |  |






<a name="metastore.GetDatabaseResponse"/>

### GetDatabaseResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| database | [Database](#metastore.Database) |  |  |
| status | [RequestStatus](#metastore.RequestStatus) |  |  |






<a name="metastore.GetTableRequest"/>

### GetTableRequest
Request to get table by its ID.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| catalog | [string](#string) |  |  |
| db_id | [Id](#metastore.Id) |  | Database ID |
| id | [Id](#metastore.Id) |  |  |
| cookie | [string](#string) |  |  |






<a name="metastore.GetTableResponse"/>

### GetTableResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| table | [Table](#metastore.Table) |  |  |
| status | [RequestStatus](#metastore.RequestStatus) |  |  |






<a name="metastore.Id"/>

### Id
Objects have unique name and unique ID.

Both name and ID are just sequence of bytes - there are no
assumptions about encoding or length.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| id | [string](#string) |  |  |






<a name="metastore.ListDatabasesRequest"/>

### ListDatabasesRequest
Request to get list of databases
If exclude_params is set, result may omit parameters


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| catalog | [string](#string) |  |  |
| cookie | [string](#string) |  |  |
| name_pattern | [string](#string) |  |  |
| exclude_params | [bool](#bool) |  |  |






<a name="metastore.ListTablesRequest"/>

### ListTablesRequest
Request to get list of databases
If exclude_params is set, result may omit parameters


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| catalog | [string](#string) |  |  |
| db_id | [Id](#metastore.Id) |  |  |
| cookie | [string](#string) |  |  |






<a name="metastore.Order"/>

### Order
sort order of a column (column name along with asc/desc)


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| col | [string](#string) |  | sort column name |
| ascending | [bool](#bool) |  | asc(1) or desc(0) |






<a name="metastore.RequestStatus"/>

### RequestStatus
General status for results


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | [RequestStatus.Status](#metastore.RequestStatus.Status) |  |  |
| error | [string](#string) |  | detailed error message |
| cookie | [string](#string) |  | copied from request |






<a name="metastore.SerDeInfo"/>

### SerDeInfo
Serialization/Deserialization information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [SerdeType](#metastore.SerdeType) |  | Serde type. If CUSTOM, use the name |
| name | [string](#string) |  | name of the serde, table name by default |
| serializationLib | [string](#string) |  | usually the class that implements the extractor &amp; loader |
| parameters | [SerDeInfo.ParametersEntry](#metastore.SerDeInfo.ParametersEntry) | repeated | NOTE: Should we enum this as well?

initialization parameters |






<a name="metastore.SerDeInfo.ParametersEntry"/>

### SerDeInfo.ParametersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="metastore.StorageDescriptor"/>

### StorageDescriptor
StorageDescriptor holds all the information about physical storage of the data belonging to a table


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cols | [FieldSchema](#metastore.FieldSchema) | repeated | required (refer to types defined above) |
| location | [string](#string) |  | Object location |
| inputFormat | [InputFormat](#metastore.InputFormat) |  | Specification of input format. If custom, use inputFormatName. |
| inputFormatName | [string](#string) |  | Name of input format if custom |
| outputFormat | [OutputFormat](#metastore.OutputFormat) |  | Specification of input format. If custom, use outputFormatName. |
| outputFormatName | [string](#string) |  | Name of output format if custom |
| numBuckets | [int32](#int32) |  | this must be specified if there are any dimension columns |
| serdeInfo | [SerDeInfo](#metastore.SerDeInfo) |  | serialization and deserialization information |
| bucketCols | [string](#string) | repeated | reducer grouping columns and clustering columns and bucketing columns` |
| sortCols | [Order](#metastore.Order) | repeated | sort order of the data in each bucket |
| parameters | [StorageDescriptor.ParametersEntry](#metastore.StorageDescriptor.ParametersEntry) | repeated | any user supplied key value hash |
| system_parameters | [StorageDescriptor.SystemParametersEntry](#metastore.StorageDescriptor.SystemParametersEntry) | repeated | Internal parameters not settable by user |






<a name="metastore.StorageDescriptor.ParametersEntry"/>

### StorageDescriptor.ParametersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="metastore.StorageDescriptor.SystemParametersEntry"/>

### StorageDescriptor.SystemParametersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="metastore.Table"/>

### Table
Table information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [Id](#metastore.Id) |  |  |
| seq_id | [uint64](#uint64) |  | Sequential ID within database |
| sd | [StorageDescriptor](#metastore.StorageDescriptor) |  | storage descriptor of the table |
| partitionKeys | [FieldSchema](#metastore.FieldSchema) | repeated | partition keys of the table. only primitive types are supported |
| tableType | [TableType](#metastore.TableType) |  | table type enum, e.g. EXTERNAL_TABLE |
| parameters | [Table.ParametersEntry](#metastore.Table.ParametersEntry) | repeated | User-settable parameters |
| system_parameters | [Table.SystemParametersEntry](#metastore.Table.SystemParametersEntry) | repeated | Internal parameters |






<a name="metastore.Table.ParametersEntry"/>

### Table.ParametersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |






<a name="metastore.Table.SystemParametersEntry"/>

### Table.SystemParametersEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | [string](#string) |  |  |
| value | [string](#string) |  |  |





 


<a name="metastore.InputFormat"/>

### InputFormat
Known Input Formats. CUSTOM means that it should be specified as a string.

| Name | Number | Description |
| ---- | ------ | ----------- |
| IF_CUSTOM | 0 |  |
| IF_SEQUENCE | 1 |  |
| IF_TEXT | 2 |  |
| IF_HIVE | 3 |  |



<a name="metastore.OutputFormat"/>

### OutputFormat
Known Output Formats. CUSTOM means that it should be specified as a string.

| Name | Number | Description |
| ---- | ------ | ----------- |
| OF_CUSTOM | 0 |  |
| OF_SEQUENCE | 2 |  |
| OF_IGNORE_KEY | 3 |  |
| OF_HIVE | 4 |  |



<a name="metastore.RequestStatus.Status"/>

### RequestStatus.Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| OK | 0 | successful request |
| ERROR | 1 | General error |
| NOTFOUND | 2 | Requested object not found |
| CONFLICT | 3 | Object already exists |
| BUSY | 4 | Object is busy/used and can&#39;t be accessed/destroyed |
| INTERNAL | 5 | Internal server error |



<a name="metastore.SerdeType"/>

### SerdeType
Known SerDes are represented using enum. Unknown ones are represented using strings.

| Name | Number | Description |
| ---- | ------ | ----------- |
| SERDE_CUSTOM | 0 |  |
| SERDE_LAZY_SIMPLE | 1 |  |
| SERDE_AVRO | 2 |  |
| SERDE_JSON | 3 |  |
| SERDE_ORC | 4 |  |
| SERDE_REGEX | 5 |  |
| SERDE_THRIFT | 6 |  |
| SERDE_PARQUET | 7 |  |
| SERDE_CSV | 8 |  |



<a name="metastore.TableType"/>

### TableType


| Name | Number | Description |
| ---- | ------ | ----------- |
| TTYPE_MANAGED | 0 |  |
| TTYPE_EXTERNAL | 1 |  |
| TTYPE_INDEX | 2 |  |


 

 


<a name="metastore.Metastore"/>

### Metastore
Metastore service

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateDabatase | [CreateDatabaseRequest](#metastore.CreateDatabaseRequest) | [GetDatabaseResponse](#metastore.CreateDatabaseRequest) | Create a new database |
| GetDatabase | [GetDatabaseRequest](#metastore.GetDatabaseRequest) | [GetDatabaseResponse](#metastore.GetDatabaseRequest) | Get database information |
| ListDatabases | [ListDatabasesRequest](#metastore.ListDatabasesRequest) | [Database](#metastore.ListDatabasesRequest) | Get collection of databases |
| DropDatabase | [DropDatabaseRequest](#metastore.DropDatabaseRequest) | [RequestStatus](#metastore.DropDatabaseRequest) | Destroy a database |
| CreateTable | [CreateTableRequest](#metastore.CreateTableRequest) | [GetTableResponse](#metastore.CreateTableRequest) | Create a new table |
| GetTable | [GetTableRequest](#metastore.GetTableRequest) | [GetTableResponse](#metastore.GetTableRequest) | Get table information |
| ListTables | [ListTablesRequest](#metastore.ListTablesRequest) | [Table](#metastore.ListTablesRequest) | Get collection of tables |
| DropTable | [DropTableRequest](#metastore.DropTableRequest) | [RequestStatus](#metastore.DropTableRequest) | Destroy a table |

 



## Scalar Value Types

| .proto Type | Notes | C++ Type | Java Type | Python Type |
| ----------- | ----- | -------- | --------- | ----------- |
| <a name="double" /> double |  | double | double | float |
| <a name="float" /> float |  | float | float | float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long |
| <a name="bool" /> bool |  | bool | boolean | boolean |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str |

