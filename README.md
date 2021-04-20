# OAS3-Agg

OAS3-Agg is an aggregator utility that searches through and collects OpenAPI Specifications (OAS3) across source code files and generates a single specification document.

The program expects OAS3 specifications in YAML format that are enclosed in a block comment and start with `OAS3-Specification`. Hence, a specification block recognized by the program looks as follows:
```golang
...

/* OAS3-Specification
<YOUR API SPECIFICATION IN YAML FORMAT HERE>
*/

...

```

Currently, only block comments  of syntax `/*...*/` are supported (e.g. golang, C/C++, java).

## Installation
Simply go-get and install OAS3-Agg with
```bash
go get github.com/kahefi/oas3-agg
```

## Usage
In order to aggregate the specifications from all files in the current directory and all of its subdirectories, run
```bash
oas3-agg generate ./ -o "./oas3_spec.yaml"
```
The program will generate a unified specification and write it to the file `oas3_spec.yaml`.
