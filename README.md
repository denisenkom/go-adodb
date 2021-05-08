# A pure Go MSSQL driver for Go's database/sql package

[![Go Reference](https://pkg.go.dev/badge/github.com/denisenkom/go-mssqldb.svg)](https://pkg.go.dev/github.com/denisenkom/go-mssqldb)
[![Build status](https://ci.appveyor.com/api/projects/status/jrln8cs62wj9i0a2?svg=true)](https://ci.appveyor.com/project/denisenkom/go-mssqldb)
[![codecov](https://codecov.io/gh/denisenkom/go-mssqldb/branch/master/graph/badge.svg)](https://codecov.io/gh/denisenkom/go-mssqldb)

## Install

Requires Go 1.8 or above.

Install with `go get github.com/denisenkom/go-mssqldb` .

## Connection Parameters and DSN

The recommended connection string uses a URL format:
`sqlserver://username:password@host/instance?param1=value&param2=value`
Other supported formats are listed below.

### Common parameters:

* `user id` - enter the SQL Server Authentication user id or the Windows Authentication user id in the DOMAIN\User format. On Windows, if user id is empty or missing Single-Sign-On is used. The user domain sensitive to the case which is defined in the connection string.
* `password`
* `database`
* `connection timeout` - in seconds (default is 0 for no timeout), set to 0 for no timeout. Recommended to set to 0 and use context to manage query and connection timeouts.
* `dial timeout` - in seconds (default is 15), set to 0 for no timeout
* `encrypt`
  * `disable` - Data send between client and server is not encrypted.
  * `false` - Data sent between client and server is not encrypted beyond the login packet. (Default)
  * `true` - Data sent between client and server is encrypted.
* `app name` - The application name (default is go-mssqldb)

### Connection parameters for ODBC and ADO style connection strings:

* `server` - host or host\instance (default localhost)
* `port` - used only when there is no instance in server (default 1433)

### Less common parameters:

* `keepAlive` - in seconds; 0 to disable (default is 30)
* `failoverpartner` - host or host\instance (default is no partner). 
* `failoverport` - used only when there is no instance in failoverpartner (default 1433)
* `packet size` - in bytes; 512 to 32767 (default is 4096)
  * Encrypted connections have a maximum packet size of 16383 bytes
  * Further information on usage: https://docs.microsoft.com/en-us/sql/database-engine/configure-windows/configure-the-network-packet-size-server-configuration-option
* `log` - logging flags (default 0/no logging, 63 for full logging)
  *  1 log errors
  *  2 log messages
  *  4 log rows affected
  *  8 trace sql statements
  * 16 log statement parameters
  * 32 log transaction begin/end
* `TrustServerCertificate`
  * false - Server certificate is checked. Default is false if encypt is specified.
  * true - Server certificate is not checked. Default is true if encrypt is not specified. If trust server certificate is true, driver accepts any certificate presented by the server and any host name in that certificate. In this mode, TLS is susceptible to man-in-the-middle attacks. This should be used only for testing.
* `certificate` - The file that contains the public key certificate of the CA that signed the SQL Server certificate. The specified certificate overrides the go platform specific CA certificates.
* `hostNameInCertificate` - Specifies the Common Name (CN) in the server certificate. Default value is the server host.
* `ServerSPN` - The Kerberos SPN (Service Principal Name) for the server. Default is MSSQLSvc/host:port.
* `Workstation ID` - The workstation name (default is the host name)
* `ApplicationIntent` - Can be given the value `ReadOnly` to initiate a read-only connection to an Availability Group listener. The `database` must be specified when connecting with `Application Intent` set to `ReadOnly`. 

  
### The connection string can be specified in one of three formats:


1. URL: with `sqlserver` scheme. username and password appears before the host. Any instance appears as
    the first segment in the path. All other options are query parameters. Examples:

  * `sqlserver://username:password@host/instance?param1=value&param2=value`
  * `sqlserver://username:password@host:port?param1=value&param2=value`
  * `sqlserver://sa@localhost/SQLExpress?database=master&connection+timeout=30` // `SQLExpress instance.
  * `sqlserver://sa:mypass@localhost?database=master&connection+timeout=30`     // username=sa, password=mypass.
  * `sqlserver://sa:mypass@localhost:1234?database=master&connection+timeout=30` // port 1234 on localhost.
  * `sqlserver://sa:my%7Bpass@somehost?connection+timeout=30` // password is "my{pass"

  A string of this format can be constructed using the `URL` type in the `net/url` package.

```go
  query := url.Values{}
  query.Add("app name", "MyAppName")

  u := &url.URL{
      Scheme:   "sqlserver",
      User:     url.UserPassword(username, password),
      Host:     fmt.Sprintf("%s:%d", hostname, port),
      // Path:  instance, // if connecting to an instance instead of a port
      RawQuery: query.Encode(),
  }
  db, err := sql.Open("sqlserver", u.String())
```

2. ADO: `key=value` pairs separated by `;`. Values may not contain `;`, leading and trailing whitespace is ignored.
     Examples:
	
  * `server=localhost\\SQLExpress;user id=sa;database=master;app name=MyAppName`
  * `server=localhost;user id=sa;database=master;app name=MyAppName`

3. ODBC: Prefix with `odbc`, `key=value` pairs separated by `;`. Allow `;` by wrapping
    values in `{}`. Examples:
	
  * `odbc:server=localhost\\SQLExpress;user id=sa;database=master;app name=MyAppName`
  * `odbc:server=localhost;user id=sa;database=master;app name=MyAppName`
  * `odbc:server=localhost;user id=sa;password={foo;bar}` // Value marked with `{}`, password is "foo;bar"
  * `odbc:server=localhost;user id=sa;password={foo{bar}` // Value marked with `{}`, password is "foo{bar"
  * `odbc:server=localhost;user id=sa;password={foobar }` // Value marked with `{}`, password is "foobar "
  * `odbc:server=localhost;user id=sa;password=foo{bar`   // Literal `{`, password is "foo{bar"
  * `odbc:server=localhost;user id=sa;password=foo}bar`   // Literal `}`, password is "foo}bar"
  * `odbc:server=localhost;user id=sa;password={foo{bar}` // Literal `{`, password is "foo{bar"
  * `odbc:server=localhost;user id=sa;password={foo}}bar}` // Escaped `} with `}}`, password is "foo}bar"

### Azure Active Directory authentication

Azure Active Directory authentication uses temporary authentication tokens to authenticate.
The `mssql` package does not provide an implementation to obtain tokens: instead, import the
`azuread` package and use driver name `azuresql`. This driver uses the
[Active Directory Authentication Library for Go](https://github.com/Azure/go-autorest/tree/master/autorest/adal)
to obtain Azure Active Directory authentication tokens.

Authentication using Active Directory is enabled using the `fedauth` connection parameter,
in combination with the `user id` and `password` (or URL username and password).

  * `fedauth=ActiveDirectoryApplication` - authenticates using an Azure Active Directory application client ID and client secret or certificate.
    
    Set the `user id` to `clientID@tenantID` for your service principal. If using a client secret, set the `password` to the client secret. If using client certificates, provide the path to the PEM file containing the certificate concatenated with the RSA private key in the `clientcertpath` parameter, and set the `password` to the passphrase needed to decrypt the RSA private key (omit or leave blank if unencrypted).

  * `fedauth=ActiveDirectoryMSI` - authenticates using the managed service identity (MSI) attached to the VM (system identity), or a specific user-assigned identity.
    
    To select a user-assigned identity, specify a client ID in the `user id` parameter. For the system-assigned identity, leave the `user id` empty.

  * `fedauth=ActiveDirectoryPassword` - authenticates an Azure Active Directory user account.
    
    Set the `user id` to `user@domain.com` and the `password`. This method is not recommended for general use and does not support multi-factor authentication for accounts.


```golang
import (
  "database/sql"
  "net/url"

  // Import the Azure AD driver module (also imports the regular driver package)
  "github.com/denisenkom/go-mssqldb/azuread"
)

func ConnectWithMSI() (*sql.DB, error) {
  return sql.Open(azuread.DriverName, "sqlserver://azuresql.database.windows.net?database=yourdb&fedauth=ActiveDirectoryMSI")
}
```

As an alternative, you can select the federated authentication library and Active Directory
using the connection string parameters, but then implement your own routine for obtaining
tokens. The second example shows how this could be used to add in a token for the Azure AD
Integrated authentication scenario.

```golang
import (
  "context"
  "database/sql"
  "net/url"

  // Import the driver
  "github.com/denisenkom/go-mssqldb"
)

func ConnectWithSecurityToken() (*sql.DB, error) {
  conn, err := mssql.NewSecurityTokenConnector(
    "sqlserver://azuresql.database.windows.net?database=yourdb",
    func(ctx context.Context) (string, error) {
      return "the token", nil
    },
  )
  if err != nil {
    // handle errors in DSN
  }

  return sql.OpenDB(conn), nil
}

func ConnectWithADIntegrated() (*sql.DB, error) {
  conn, err := mssql.NewActiveDirectoryTokenConnector(
    "sqlserver://azuresq;.database.windows.net?database=yourdb",
    2, // Active Directory workflow: 1 = user/password, 2 = integrated, 3 = MSI
    func(ctx context.Context, serverSPN, stsURL string) (string, error) {
      return "the token", nil
    },
  )
  if err != nil {
    // handle errors in DSN
  }

  return sql.OpenDB(conn), nil
}
```

## Executing Stored Procedures

To run a stored procedure, set the query text to the procedure name:
```go
var account = "abc"
_, err := db.ExecContext(ctx, "sp_RunMe",
	sql.Named("ID", 123),
	sql.Named("Account", sql.Out{Dest: &account}),
)
```

## Reading Output Parameters from a Stored Procedure with Resultset

To read output parameters from a stored procedure with resultset, make sure you read all the rows before reading the output parameters:
```go
sqltextcreate := `
CREATE PROCEDURE spwithoutputandrows
	@bitparam BIT OUTPUT
AS BEGIN
	SET @bitparam = 1
	SELECT 'Row 1'
END
`
var bitout int64
rows, err := db.QueryContext(ctx, "spwithoutputandrows", sql.Named("bitparam", sql.Out{Dest: &bitout}))
var strrow string
for rows.Next() {
	err = rows.Scan(&strrow)
}
fmt.Printf("bitparam is %d", bitout)
```

## Caveat for local temporary tables

Due to protocol limitations, temporary tables will only be allocated on the connection
as a result of executing a query with zero parameters. The following query
will, due to the use of a parameter, execute in its own session,
and `#mytemp` will be de-allocated right away:

```go
conn, err := pool.Conn(ctx)
defer conn.Close()
_, err := conn.ExecContext(ctx, "select @p1 as x into #mytemp", 1)
// at this point #mytemp is already dropped again as the session of the ExecContext is over
```

To work around this, always explicitly create the local temporary
table in a query without any parameters. As a special case, the driver
will then be able to execute the query directly on the
connection-scoped session. The following example works:

```go
conn, err := pool.Conn(ctx)

// Set us up so that temp table is always cleaned up, since conn.Close()
// merely returns conn to pool, rather than actually closing the connection.
defer func() {
	_, _ = conn.ExecContext(ctx, "drop table #mytemp")  // always clean up
	conn.Close() // merely returns conn to pool
}()


// Since we not pass any parameters below, the query will execute on the scope of
// the connection and succeed in creating the table.
_, err := conn.ExecContext(ctx, "create table #mytemp ( x int )")

// #mytemp is now available even if you pass parameters
_, err := conn.ExecContext(ctx, "insert into #mytemp (x) values (@p1)", 1)

```

## Return Status

To get the procedure return status, pass into the parameters a
`*mssql.ReturnStatus`. For example:
```
var rs mssql.ReturnStatus
_, err := db.ExecContext(ctx, "theproc", &rs)
log.Printf("status=%d", rs)
```

or

```
var rs mssql.ReturnStatus
_, err := db.QueryContext(ctx, "theproc", &rs)
for rows.Next() {
	err = rows.Scan(&val)
}
log.Printf("status=%d", rs)
```

Limitation: ReturnStatus cannot be retrieved using `QueryRow`.

## Parameters

The `sqlserver` driver uses normal MS SQL Server syntax and expects parameters in
the sql query to be in the form of either `@Name` or `@p1` to `@pN` (ordinal position).

```go
db.QueryContext(ctx, `select * from t where ID = @ID and Name = @p2;`, sql.Named("ID", 6), "Bob")
```

### Parameter Types

To pass specific types to the query parameters, say `varchar` or `date` types,
you must convert the types to the type before passing in. The following types
are supported:

 * string -> nvarchar
 * mssql.VarChar -> varchar
 * time.Time -> datetimeoffset or datetime (TDS version dependent)
 * mssql.DateTime1 -> datetime
 * mssql.DateTimeOffset -> datetimeoffset
 * "github.com/golang-sql/civil".Date -> date
 * "github.com/golang-sql/civil".DateTime -> datetime2
 * "github.com/golang-sql/civil".Time -> time
 * mssql.TVP -> Table Value Parameter (TDS version dependent)

## Important Notes

 * [LastInsertId](https://golang.org/pkg/database/sql/#Result.LastInsertId) should
    not be used with this driver (or SQL Server) due to how the TDS protocol
	works. Please use the [OUTPUT Clause](https://docs.microsoft.com/en-us/sql/t-sql/queries/output-clause-transact-sql)
	or add a `select ID = convert(bigint, SCOPE_IDENTITY());` to the end of your
	query (ref [SCOPE_IDENTITY](https://docs.microsoft.com/en-us/sql/t-sql/functions/scope-identity-transact-sql)).
	This will ensure you are getting the correct ID and will prevent a network round trip.
 * [NewConnector](https://godoc.org/github.com/denisenkom/go-mssqldb#NewConnector)
    may be used with [OpenDB](https://golang.org/pkg/database/sql/#OpenDB).
 * [Connector.SessionInitSQL](https://godoc.org/github.com/denisenkom/go-mssqldb#Connector.SessionInitSQL)
	may be set to set any driver specific session settings after the session
	has been reset. If empty the session will still be reset but use the database
	defaults in Go1.10+.

## Features

* Can be used with SQL Server 2005 or newer
* Can be used with Microsoft Azure SQL Database
* Can be used on all go supported platforms (e.g. Linux, Mac OS X and Windows)
* Supports new date/time types: date, time, datetime2, datetimeoffset
* Supports string parameters longer than 8000 characters
* Supports encryption using SSL/TLS
* Supports SQL Server and Windows Authentication
* Supports Single-Sign-On on Windows
* Supports connections to AlwaysOn Availability Group listeners, including re-direction to read-only replicas.
* Supports query notifications

## Tests

`go test` is used for testing. A running instance of MSSQL server is required.
Environment variables are used to pass login information.

Example:

    env SQLSERVER_DSN=sqlserver://user:pass@hostname/instance?database=test1 go test

## Deprecated

These features still exist in the driver, but they are are deprecated.

### Query Parameter Token Replace (driver "mssql")

If you use the driver name "mssql" (rather then "sqlserver") the SQL text
will be loosly parsed and an attempt to extract identifiers using one of

* ?
* ?nnn
* :nnn
* $nnn

will be used. This is not recommended with SQL Server. 
There is at least one existing `won't fix` issue with the query parsing.

Use the native "@Name" parameters instead with the "sqlserver" driver name.

## Known Issues

* SQL Server 2008 and 2008 R2 engine cannot handle login records when SSL encryption is not disabled.
To fix SQL Server 2008 R2 issue, install SQL Server 2008 R2 Service Pack 2.
To fix SQL Server 2008 issue, install Microsoft SQL Server 2008 Service Pack 3 and Cumulative update package 3 for SQL Server 2008 SP3.
More information: http://support.microsoft.com/kb/2653857
