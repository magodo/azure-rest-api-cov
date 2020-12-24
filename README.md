# azure-rest-api-cov

A tool consumes [schema input](##schema-input-format) and compare against [azure-rest-api-specs](https://github.com/Azure/azure-rest-api-specs) to get property covereage report.

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

```shell
azure-rest-api-cov -i <schema input> <options>
  -a    whether to export the RP level coverage
  -c string
        The azure-rest-api-spec repo commit hash to checkout
  -f    whether to force create the workspace
  -i string
        The path to the schema input file, which is expected to hold a JSON array of input schemas
  -r    whether to force regenerate the repository index in the workspace
  -w string
        The path to the workspace for caching the azure-rest-api repository and other stuffs (default "$HOME/.cache/azure-rest-api-cov")
```

The tool will setup a workspace (controlled by `-w` option, which by default is the [user cache directory](https://golang.org/pkg/os/#UserCacheDir)) that will hold the [azure-rest-api-specs](https://github.com/Azure/azure-rest-api-specs) repository and some index files for accelaration.

Note that you should ensure that the commit of the [azure-rest-api-specs](https://github.com/Azure/azure-rest-api-specs) from which you generate the [schema input](##schema-input-format) is the same as the one you run against this tool, via `-c` option.

If `-a` is set, the output will contain all existing resource providers in the  [azure-rest-api-specs](https://github.com/Azure/azure-rest-api-specs), together with its API paths, versions and operations, no matter whether they appears in the  [schema input](##schema-input-format).

## Schema Input Format

The schema input is a json array, where each element is a json object of following properties:

- `api_version` (REQUIRED): The API version defined in the `version` field of [Info Object](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#infoObject)

    E.g.
    
    ```json
    "api_version": "2019-11-01",
    ```
    
- `api_path` (REQUIRED): The path in the [Paths Object](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#pathsObject)

    E.g. 
    
    ```json
    "api_path": "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.DataShare/accounts/{accountName}/shares/{shareName}/dataSets/{dataSetName}",
    ```

- `operation` (REQUIRED): One of `GET`, `PUT`, `POST`, `PATCH`, `DELETE`.

- `variant` (OPTIONAL): The discriminator value of the variant schema given the root schema is a [discriminator](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#fixed-fields-13). The discriminator value is the variant schema's name, except [`x-ms-discriminator-value`](http://azure.github.io/autorest/extensions/#x-ms-discriminator-value) is set, which takes higher priority.

    E.g.

    ```json
    "variant": "AdlsGen1File",
    ```

- `properties` (OPTIONAL): This is only optional for `DELETE`, but is required for the other operations. This is an array of [properties](###property-format) indicating those properties are covered

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
