Buildpack Runtime Acceptance Tests
---

### Functionality

Test that the compiled binaries of the buildpacks are working as expected.

### Usage

BRATs use the [Cutlass](https://github.com/cloudfoundry/libbuildpack/cutlass) framework for running integration tests.

Example of testing the Apt buildpack:

1. Source the .envrc file in the buildpack directory.

   ```bash
   source .envrc
   ```
   To simplify the process in the future, install [direnv](https://direnv.net/) which will automatically source .envrc when you change directories.

1. Run integration tests

    ```bash
    ./scripts/integration.sh apt develop
    ```

More information can be found on Github [cutlass](https://github.com/cloudfoundry/libbuildpack/cutlass).

Note that the appropriate language tag is required to run the full BRATS suite for the specified buildpack.
The interpreter matrix tests will not execute unless the tag for the appropriate interpreter is passed into the rspec arguments.

It is required to specify a git branch of the buildpack to test against. In the example above it is develop
