---
title: Releases

---



# Gitopssets Controller Releases ~ENTERPRISE~

## v0.16.1
2023-09-06

- Bump client-go to 0.26.8 - avoids a buggy version of the upstream client
  package

## v0.16.0
2023-09-05

- Fix partial-apply resources bug - errors generating resources could lead to
  incomplete inventories and errors when regenerating resources
- Bump the memory limits for the Helm chart and document that these may need to
  be increased.

## v0.15.3
2023-08-17

- Fix bug when a Matrix generator doesn't generate any elements.

## v0.15.2
2023-08-17

- Update the ImagePolicy generator to add the image by splitting the image from
  the tag.

## v0.15.1
2023-08-17

- Fix bug in the processing of empty artifacts in GitRepositories and
  OCIRepositories - the directory listing will also return the special empty
  marker when the Repository is empty.

## v0.15.0
2023-08-10

- ClusterGenerator - return labels as generic maps - this makes it easier to
  query for labels in a map.

## v0.14.1
2023-07-26

- When a GitRepository or OCIRepository artifact is empty, handle this as a
  special case that doesn't mean "no resources" this prevents removal of
  resources when the Flux resource hasn't populated yet.

## v0.14.0
2023-07-14

- Support multidoc when rendering via the CLI tool
- Allow custom CAs for the APIGenerator HTTPClient
- Single element Matrix generation - compress multiple Matrix elements into a
  single element
- Implement element index and repeat index
- Local GitRepository generation from the filesystem in the CLI tool
- Implement OCIGenerator - functionally very similar to the GitRepository
  generator

## v0.13.3
2023-06-26

- Secrets are now provided in Elements as strings rather than byte slices

## v0.13.1
2023-06-21

- Expose the latest tag not just the latest image in the ImageRepository

## v0.13.0
2023-06-20

- Fix bug in matrix generator when updating GitRepository resources
- Config generator - track Secrets and ConfigMaps and generate from them
- CLI tool for rendering GitOpsSets

## v0.12.0
2023-05-24

- Allow altering the delimiters
- Imagerepository generator by @bigkevmcd in #71
- Allow cross-namespace resources
- Upgrade the matrix to support "unlimited" numbers of generators
- Add support for Flux annotation triggered rereconciliation

## v0.11.0
2023-05-10

- Support for using the `repeat` mechanism within maps not just arrays

## v0.10.0
2023-04-28

- Bump to support Flux v2

## v0.9.0
2023-04-27

- Fail if we cannot find a relevant generator
- Suppress caching of Secrets and ConfigMaps
- Improve APIClient error message
- Support correctly templating numbers - insertion of numeric values is improved

## v0.8.0
2023-04-13

- Add events recording to GitOpsSets
- Fix updating of ConfigMaps

## v0.7.0
2023-03-30

- Implement custom delimiters

## v0.6.1
2023-03-20

- Implement optional list expansion