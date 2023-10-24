# azure-rest-api-cov

A tool consumes [schema input](#schema-input-format) and compare against [azure-rest-api-specs](https://github.com/Azure/azure-rest-api-specs) to get property covereage report.

## Install

Since this repo is private, it is not possible to install it via a single `go get` command.

### Pre-built

Download the pre-built binary from releases.

### Local Build

```shell
$ git clone https://github.com/magodo/azure-rest-api-cov
$ cd azure-rest-api-cov
$ go install
```

## Usage

The tool will setup a workspace (controlled by `-w` option, which by default is the [user cache directory](https://golang.org/pkg/os/#UserCacheDir)) that will hold the [azure-rest-api-specs](https://github.com/Azure/azure-rest-api-specs) repository and some index files for accelaration.

Note that you should ensure that the commit of the [azure-rest-api-specs](https://github.com/Azure/azure-rest-api-specs) from which you generate the [schema input](#schema-input-format) is the same as the one you run against this tool, via `-c` option.

Typically, one follow the process below to use this tool:

- Generate the schema input file (see the "Schema Input Format" section below for its format) by some means, e.g. using `azure-rest-api-cov-terraform` for terraform provider to generate the input file
- Run `azure-rest-api-cov [option] coverage [option]` to consume the schema input file above, and generate the coverage file, which contains the covered/uncovered expanded properties for each API model (optionally plus predicting models in newer API versions)
- Run `azure-rest-api-cov [option] feeddb [option]` to consume the coverage file above, which will generate a report file that can contains not only the API operations covered, also include the other uncovered API models among all the available RPs. This report file is then expected to be upload to a backend cosmos db, to be consumed by the frontend portal, which is handled by [another project](https://github.com/magodo/azure-rest-api-viewer2).

## Schema Input Format

The schema input is a json array, where each element is a json object of following properties:

- `api_version` (REQUIRED): The API version defined in the `version` field of [Info Object](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#infoObject)

    E.g.
    
    ```json
    "api_version": "2019-11-01",
    ```
    
- `api_path` (REQUIRED): The path in the [Paths Object](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#pathsObject). The path is normalized by uppercase the invariant path segments and remove the variant segments around the `{}`.

    E.g. 
    
    ```json
    "api_path": "/SUBSCRIPTIONS/{}/RESOURCEGROUPS/{}/PROVIDERS/MICROSOFT.DATASHARE/ACCOUNTS/{}/SHARES/{}/DATASETS/{}",
    ```

- `operation` (REQUIRED): One of `GET`, `PUT`, `POST`, `PATCH`, `DELETE`.

- `variant` (OPTIONAL): The discriminator value of the variant schema given the root schema is a [discriminator](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#fixed-fields-13). The discriminator value is the variant schema's name, except [`x-ms-discriminator-value`](http://azure.github.io/autorest/extensions/#x-ms-discriminator-value) is set, which takes higher priority.

    E.g.

    ```json
    "variant": "AdlsGen1File",
    ```

- `properties` (OPTIONAL): This is only optional for `DELETE`, but is required for the other operations. This is an array of [properties](#property-format) indicating those properties are covered

    E.g.

    ```json
    "properties": [
      "id",
      "kind",
      "properties/accountName",
      "properties/dataSetId",
      "properties/fileName",
      "properties/folderPath",
      "properties/resourceGroup",
      "properties/subscriptionId"
    ]
    ```

### Property Format

The properties are flattend to segments till the most deep level, whose type is a primary type. Two property segments are separated by `/`. If the segment contains `/`, one can escape it by prefixing a `\`.

For property segment whose type is an array, we will only continue to flatten it if the type of its item is not a primary type.

In case the schema of a property segment is a discriminator, we will flatten all its variants schemas. The variant value of each one is annotated by surrounding the discriminator value between `{` and `}` right after the property name.

E.g.

```json
    "p1/p_discriminator{v1}/x",
    "p1/p_discriminator{v2}/y",
```

If the discriminatr value contains `}`, it should be escaped by prefixing a `\`.

## Limitations

Currently, only management plane APIs are handled, no data plane support yet.
