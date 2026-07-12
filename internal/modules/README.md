# Module template

Each bounded context uses `domain`, `port`, `app`, `adapter/http`, and `adapter/postgres` packages. Domain code remains independent of adapters. Cross-module communication uses published ports or domain events.
