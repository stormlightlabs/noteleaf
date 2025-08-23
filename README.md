# Noteleaf

[![codecov](https://codecov.io/gh/stormlightlabs/noteleaf/branch/main/graph/badge.svg)](https://codecov.io/gh/stormlightlabs/noteleaf)
[![Go Report Card](https://goreportcard.com/badge/github.com/stormlightlabs/noteleaf)](https://goreportcard.com/report/github.com/stormlightlabs/noteleaf)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/stormlightlabs/noteleaf)](go.mod)

```sh
                                    ,,                       ,...
`7MN.   `7MF'         mm          `7MM                     .d' ""
  MMN.    M           MM            MM                     dM`
  M YMb   M  ,pW"Wq.mmMMmm .gP"Ya   MM  .gP"Ya   ,6"Yb.   mMMmm
  M  `MN. M 6W'   `Wb MM  ,M'   Yb  MM ,M'   Yb 8)   MM    MM
  M   `MM.M 8M     M8 MM  8M""""""  MM 8M""""""  ,pm9MM    MM
  M     YMM YA.   ,A9 MM  YM.    ,  MM YM.    , 8M   MM    MM
.JML.    YM  `Ybmd9'  `Mbmo`Mbmmd'.JMML.`Mbmmd' `Moo9^Yo..JMML.
```

A note, task & time management CLI built with Golang & Charm.sh libs. Inspired by TaskWarrior & todo.txt CLI applications.

## Development

Requires Go v1.24+

### Testing

#### Handlers

The command handlers (`cmd/handlers/`) use a multi-layered testing approach for happy and error paths:

- Environment Isolation
    - Tests manipulate environment variables to simulate configuration failures
- File System Simulation
    - By creating temporary directories with controlled permissions, tests verify that handlers properly handle file system errors like read-only directories, missing files, and permission denied scenarios.
- Data Corruption Testing
    - Tests intentionally corrupt database schemas and configuration files to ensure handlers detect and report data integrity issues.
- Table-Driven Error Testing
    - Systematic testing of multiple error scenarios using structured test tables
