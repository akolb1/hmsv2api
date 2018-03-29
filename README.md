# hmsv2api

## Introduction

This API is different from traditional metastore API. It separates all
metadata-only operations and does not include any filesystem operations.
The assumption is that some other service or the client deals with file system
operations.
 
The API also introduces the notion of an object ID - globally unique string
associated with the object. While object name can change (e.g. with
rename operation), identity of the object never changes. The lookup can be done
either by name or by ID. The intention is to have cacheable objects which do not
change on renames.

Cookie is supposed to be used to associate multiple requests to a single session.
The value of the cookie is likely to be printed in logs so it shouldn't contain
any sensitive information.
Metastore service does not interpret the cookie but may print it in its logs.
We could call it SessionId but callers may decide to use it for whatever other
purposes, so using generic term here.

Every object belongs to a catalog. The idea is that operations across
catalogs are completely independent. They can be forwarded to different storage
engines.

- [metastore gRpc spec](protobuf/metastore.proto)
- [Swagger spec](swagger/metastore.swagger.json)

## Service

    service Metastore {
        // Create a new database
        rpc CreateDabatase (CreateDatabaseRequest) returns (GetDatabaseResponse);
        // Get database information
        rpc GetDatabase (GetDatabaseRequest) returns  (GetDatabaseResponse);
        // Get collection of databases
        rpc ListDatabases (ListDatabasesRequest) returns  (stream Database);
        // Destroy a database
        rpc DropDatabase (DropDatabaseRequest) returns (RequestStatus);
    }

## Objects

### Request Status

    message RequestStatus {
        enum Status {
            OK = 0;       // successful request
            ERROR = 1;    // General error
            NOTFOUND = 2; // Requested object not found
            CONFLICT = 3; // Object already exists
        }
        Status status;
        string error;  // detailed error message
        string cookie; // copied from request
    }

### Id

Objects belong to a specific catalog and have unique name and unique ID
in the catalog. Both name and ID are just sequence of bytes - there are no
assumptions about encoding or length.

*NOTE*: _May be we should limit the name length to prevent DOS type attacks_ 

    message Id {
        string catalog;
        string name;
        string id;
    }

### Database

Database is a container for tables.

Database object has two sets of parameters:
- User parameters are intended for user and are just transparently passed around
- System parameters are intended to be used by Hive for its internal purposes

seq_id is a numeric ID which is unique within a catalog. It can be used to track
new databases in the catalog

Original Metastore Database object also had location and owner information.
These can be represented using parameters if needed since the current
metastore service does n;t interpret either Location or Owner info.

    message Database {
        Id id;
        uint64 seq_id = 2;              // Unique sequence ID within calalog
	    string location                 // Location of Database data
        map<string, string> parameters; // Database parameters
        map<string, string> system_parameters;
    }

## Requests

### CreateDatabaseRequest

Create a new database.

Request should have the catalog and the name of the database.
The name should be unique within the catalog
unique ID is assigned by the service.

    message CreateDatabaseRequest {
        Database database;
        string cookie;
    }
    
### GetDatabaseRequest

Request to get database by its ID.

    message GetDatabaseRequest {
        Id id;
        string cookie;
    }

### GetDatabaseResponse

    message GetDatabaseResponse {
        Database database;
        RequestStatus status;
    }

### ListDatabasesRequest

Request to get list of databases. If exclude_params is set, result may omit parameters
.

    message ListDatabasesRequest {
        string catalog;
        string cookie;
        string name_pattern;
        bool   exclude_params;
    }

### DropDatabaseRequest

Request to drop a database.
Dropping a database also drops all objects contained in the database

    message DropDatabaseRequest {
        Id id;
        string cookie;
    }
