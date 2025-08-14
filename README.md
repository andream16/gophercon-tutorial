# GopherCon Tutorial

This is the code used in the workshop about "Code generation for 10x productivity - no AI"
at GopherCon UK 2025.

You can check the slides [here](./deck).

## Setting up

Run `make tidy-vendor-all` to download necessary dependencies.

## First part

We talk about code generation tools in go. You can check out
the [tools](./tools) to play around with the `handson`
exercises.

## Second part

Extend a real code generation tool in [httptestgen](./httptestgen).

This can be achieved by re-applying [these changes](https://github.com/andream16/gophercon-tutorial/commit/264304b4b8dcbdc093b57c58d78d798589859c84)
but, I challenge you not to look at them before trying to do ti on your own!