# How to perform bulk imports

To use bulk imports feature in go-mssqldb, you need to import the sql and go-mssqldb packages.

```
import (
    "database/sql"
    mssqldb "github.com/denisenkom/go-mssqldb"
)
```

The `mssqldb.CopyIn` function creates a string which can be prepared by passing it to `DB.Prepare` or `Tx.Prepare`. The string returned contains information such as the name of the table and columns to bulk import data into, and bulk options.

```
bulkImportStr := mssqldb.CopyIn("tablename", mssql.BulkOptions{}, "column1", "column2", "column3")
stmt, err := db.Prepare(bulkImportStr)
```

Bulk options can be specified using the `mssqldb.BulkOptions` type. The following is how the `BulkOptions` type is defined:

```
type BulkOptions struct {
    CheckConstraints  bool
    FireTriggers      bool
    KeepNulls         bool
    KilobytesPerBatch int
    RowsPerBatch      int
    Order             []string
    Tablock           bool
}
```

The statement can be executed many times to copy data into the table specified.

```
for i := 0; i < 10; i++ {
	_, err = stmt.Exec(col1Data[i], col2Data[i], col3Data[i])
}
```

After all the data is processed, call `Exec` once with no arguments to flush all the buffered data.

```
_, err = stmt.Exec()
```

## Example
[Bulk import example](../examples/bulk/bulk.go)