# Gleece Changelog

## Gleece v2.0.0

### Summary

*Gleece* 2 is a major milestone that includes a complete overhaul of the internal code analysis and validation facilities
as well as a multitude of small bug fixes.

These changes aim to allow us to better expand and maintain the project and provide the groundwork for powerful and unique features down the road like live OAS preview, LSP support and more.

### Features

* Added a rich, LSP oriented diagnostics system. Issues will be reporter with far greater detail and clarity

* Added many validation previously available only via the IDE extension

* Added facilities necessary to generate full project dependency graphs (`SymbolGraph.ToDot`)

* Created a `GleecePipeline` to orchestrate execution and lifecycle.
	  This allows re-using caches and previous analysis results to expedite subsequent operations.

### Enhancements

* Improved analysis speed by up to 50% via code optimization and introduction of package, file and node caches

* Adjusted most processes to yield sorted results for more consistent builds results

* Re-structured the project to provide a much clearer separation of concerns and allow for easier maintenance

* Improved test coverage

### Bugfixes 
* Fixed several cases of panic due to mis-configuration or invalid commands

* Fixed cases where documentation was not properly siphoned from some types of entities

* Fixed several issues with complex, nested type layers (*e.g*.map[string][][]int) resolution

* Fixed several issues with complex type resolution

* Fixed several issues with import detection resulting in resolution failures

* Fixed an issue that could cause type information to be emitted with incorrect `PkgPath`

* Fixed an OAS validation error caused by `[]byte` and `time` fields in returned structs
