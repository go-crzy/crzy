> Developers were waiting for code to compile, they are now waiting for tests
> to pass!

Even with containers or lambda, running test environments can take minutes. It
might require:

- deploying infrastructure components
- installing compilers, tools, databases
- managing secrets and configurations
- waiting for other tests to succeed or fail
- building, sharing and deploying artifacts

## core principles

- `simplify` even when it is hard
- `don't` spend time on useless operations
- `speed up` by optimizing steps alone and together
- `anticipate` by preparing as much as possible
- `cache` dependencies and previous tests
- `parallelize` operation AND environments
- `share` resources and tests
- `postpone` non-vital operations

`crzy` relies on those core principles to provide fast tests
