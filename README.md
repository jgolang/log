# Log v1

Simple, fast, structured and level registration in Go.

## Overview

This package offers several functions that allow you to trace controlled log errors in the files and functions where they occur, as well as useful programmer log comments in your code. This package will help you identify and segment the different types of errors log or unique circumstances during the execution of your program using criteria according to the level of impact on the business rules of its development. The package prints the useful information for the programmer in a human readable format.

## Installation

`go get -u https://github.com/jgolang/log`

## Quick Start

```go
package main

import "github.com/jgolang/log"

func main(){
    log.Println("My info....")
}
```

### Terminal output:

```terminal
2020/07/05 13:32:03     INFO    /dir/file.go:10 (function) My info...
```

## Mode

You can configure the package depending on your needs to display certain types of log by defining an environment variable on the system that runs your program.

```
[user@ /home]# export MODE="DEV"
```

### Allowed modes 

| Mode | Description |
| :------ | :--: | 
| PROD | Only prints error log | 
| DEV | Print all | 

Note: Error logs are printed regardless of this setting.

<hr>

Released under the [GPL-3.0](LICENSE.txt).

