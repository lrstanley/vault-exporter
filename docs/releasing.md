# Releasing
To release code off the master branch, increment the version in the Makefile and push up the changes.

On the next Jenkins run, accept the prompt to deploy the code which will prompt Jenkins to run the `make release` steps. Currently this includes creating a GitHub release using goreleaser as well as pushing a tagged version to Docker Hub grapeshot/vault_exporter.
