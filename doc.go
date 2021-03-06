//Package goutil is a collection of utilities for working with the go/* packages and
//the go tool.
//
//Goutil makes it easier to find, parse, and analyze Go code. It also enables
//easy use of the go tool for running generated code and creating executables
//from it. There are many miscellaneous utilities for easing
//the use of the go/* packages.
//
//DocParse has been extracted from the go/doc package as this functionality
//is not exported.
//
//Importing
//
//There are five ways to import packages with goutil:
//Import, ImportTree, ImportAll, ImportRec, and the ImportDeps method
//on *Package. The latter are wrappers around Import for common tasks. Import's
//documentation applies to all of them, unless otherwise specified.
//
//Imported packages are cached with a pointer to its build.Context as part
//of the key. The same pointer is also stored on each Package. This means you
//should never modify a build.Context after using it with this library.
//If the ctx parameter to any Import function is nil, a pointer to the go/build
//default context is used.
//
//Packages
//
//A *Package always has its go/build Context and Package set. It has methods
//to parse the files designated by its build.Package with go/ast and go/doc.
//
//With the exception of Import, the other Import functions all return
//Packages, a []*Package with methods for filter and map applications.
//
//Some methods of Package and Packages require that certain parsing actions
//be taken first, but these are always documented.
//
//Gostrap
//
//Gostrap is a utility for running the go(1) command in a temporary directory.
package goutil
