# hmsv2api

## Introduction

This API is different from traditional metastore API. It separates all
metadata-only operations and does not include any filesystem operations.
The assumption is that some other service or the client deals with file system
operations.
 
The API also introduces the notion of an object ID - globally unique string
associated with the object. While object name can change (e.g. with
rename operation), identity of the object never changes.

[metastore gRpc spec](protobuf/metastore.proto)

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
        string error; // detailed error message
    }

### Namespace

Every object belongs to a namespace.

    message Namespace {
        string name;
    }

### Cookie

Cookie is supposed to be used to associate multiple requests to a single session.
The value of the cookie is likely to be printed in logs so it shouldn't contain
any sensitive information.

Metastore service does not interpret the cookie but may print it in its logs.

    message Cookie {
        string cookie;
    }

### Id

Objects belong to a specific namespace and have unique name and unique ID
in the namespace.

    message Id {
        Namespace namespace;
        string name;
        string id;
    }

### Database

Database is a namespace for tables.

Original Metastore Database object also had location and owner information.
These can be represented using parameters if needed since the current
metastore service does n;t interpret either Location or Owner info.

    message Database {
        Id id;
        map<string, string> parameters; // Database parameters
    }

## Requests

### CreateDatabaseRequest

Create a new database.

Request should have the namespace and the name of the database.
The name should be unique within the namespace
unique ID is assigned by the service.

    message CreateDatabaseRequest {
        Id id;
        Cookie cookie;
        map<string, string> parameters;
    }
    
### GetDatabaseRequest

Request to get database by its ID.

    message GetDatabaseRequest {
        Id id;
        Cookie cookie;
    }

### GetDatabaseResponse

    message GetDatabaseResponse {
        Database database;
        RequestStatus status;
    }

### ListDatabasesRequest

Request to get list of databases

    message ListDatabasesRequest {
        Namespace namespace;
        Cookie cookie;
        string name_pattern;
    }

### DropDatabaseRequest

Request to drop a database.
Dropping a database also drops all objects contained in the database

    message DropDatabaseRequest {
        Id id;
        Cookie cookie;
    }
