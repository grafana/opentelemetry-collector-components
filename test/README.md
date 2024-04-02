# Tests

The tests are a bunch of shell scripts that, for each distribution:

- prepare the environment, such as installing tools like `tracegen`
- start the distribution
- create a data point (a trace, for now)
- check whether the data point was received
- stop the distribution

In the future, we might add more scenarios to the tests.

In the output, pay attention to the general outcome of each distribution's test. While you might see warnings or errors in the logs, they are not indicative of a failure. If you see a "PASS: sidecar", it means that the tests (eventually) passed. If you do NOT see this, check the logs for clues.
