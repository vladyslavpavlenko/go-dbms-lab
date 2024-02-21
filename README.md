# Go DBMS Lab

This lab focuses on managing structured files without using a DBMS. It is implemented on two objects linked by a 1:N relationship. As a result, two types of files are created: master and slave, which can be accessed through the user interface. The program supports operations such as reading, deleting, updating, and inserting records and subrecords.

The files use the `*.fl` format for data, the `*.ind` and `*.jk` index table for storing unused addresses. The slave file forms a linked list for sub-records, where each record in the main file is linked to the initial sub-record, and each sub-record is linked to the next and previous ones.

Deletion is accomplished by "garbage collection", where records are marked as logically deleted but not deleted immediately. In case of large data fragmentation, the files are compacted and garbage collected.
## Usage

Next command are supported:

### Inserting
`insert-m`, `insert-s`: Add new records or sub-records.

**Examples:**
```shell
$ insert-m 1 'Go Course' 'Go' 'Gopher'
```

```shell
$ insert-s 1 1 'Robert Griesemer'
```

### Reading
`get-m`, `get-s`: Access specific records and sub-records directly.

**Examples:**
```shell
$ get-m all
```

```shell
$ get-m 1
```

```shell
$ get-m 1 'category'
```

```shell
$ get-m 1 'category' 'title'
```

```shell
$ get-s all
```

### Updating
`update-m`, `update-s`: Modify specific fields of records or sub-records.

**Examples:**
```shell
$ update-m 1 '-' '-' 'Rob Pike'
```
```shell
$ update-s 1 'Ken Thompson'
```

### Deleting
`del-m`, `del-s`: Remove records or sub-records, automatically deleting all sub-records of a record when it is deleted.

**Examples:**
```shell
$ del-m 1
```

```shell
$ del-s 1
```

### Counting
`calc-m`, `calc-s`: Tally total records, sub-records, and sub-records per record.

**Examples:**

```shell
$ calc-m
```

```shell
$ calc-s
```

```shell
$ calc-s 1
```

### Utilities
`ut-m`, `ut-s`: Display all fields of master and slave files, including service fields.

## Dependencies
* [kballard/go-shellquote](https://github.com/kballard/go-shellquote)
* [olekukonko/tablewriter](https://github.com/olekukonko/tablewriter)
* [spf13/cobra](https://github.com/spf13/cobra)