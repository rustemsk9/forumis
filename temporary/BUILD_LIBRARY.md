# Build data package as dynamic library

## Step 1: Create C-compatible interface

# Create a file data/c_interface.go

package data

/_
#include <stdlib.h>
_/
import "C"
import "unsafe"

var dbManager \*DatabaseManager

//export InitDB
func InitDB(dbPath \*C.char) C.int {
path := C.GoString(dbPath)
var err error
dbManager, err = NewDatabaseManager(path)
if err != nil {
return -1
}
return 0
}

//export CreateUser
func CreateUser(name, email, password \*C.char) C.int {
user := &User{
Name: C.GoString(name),
Email: C.GoString(email),
Password: C.GoString(password),
}
err := dbManager.CreateUser(user)
if err != nil {
return -1
}
return C.int(user.Id)
}

//export CloseDB
func CloseDB() {
if dbManager != nil {
dbManager.Close()
}
}

func main() {} // Required for building as library

## Step 2: Build commands

# Build as shared library (.so on Linux/Mac, .dll on Windows):

go build -buildmode=c-shared -o libdata.so data/

# Or build as static library:

go build -buildmode=c-archive -o libdata.a data/

## Step 3: Usage in main.go

/_
#cgo LDFLAGS: -L. -ldata
#include "libdata.h"
_/
import "C"

func main() {
dbPath := C.CString("mydb.db")
defer C.free(unsafe.Pointer(dbPath))

    result := C.InitDB(dbPath)
    if result != 0 {
        panic("Failed to initialize database")
    }
    defer C.CloseDB()

    // Your application code here

}
