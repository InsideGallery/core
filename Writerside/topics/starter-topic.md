Of course. The next logical step is the **"Getting Started"** section.

This part is crucial for new users as it provides their first hands-on experience with your framework. It should be clear, concise, and lead to a quick success.

Here is a draft in markdown format.

## Getting Started

This guide will walk you through the basic steps to install the `core` framework and run a minimal example.

------

### 1. Installation

To get started, use `go get` to add the framework to your project's dependencies.

```bash
go get github.com/InsideGallery/core
```

### 2. Quick Start: A Simple Service

Let's create a small application to see the framework in action. The goal is to initialize a service and demonstrate how easily you can integrate a core component, such as a logger from the observability package.

Create a new file named `main.go` and add the following code:

```go
package main

import (
	"context
	
	_ "github.com/InsideGallery/core/fastlog/handlers/otel"
	_ "github.com/InsideGallery/core/fastlog/handlers/stderr"

	"github.com/InsideGallery/core/app"
	"github.com/InsideGallery/core/fastlog/metrics"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func main() {
	cfg, err := domain.GetConfigFromEnv(domain.EnvPrefix)
	if err != nil {
		log.Fatal("Error getting config", "err", err)
	}

	ctx := context.Background()

	app.WebMain(ctx, cfg.Port, cfg.ServerName, func(
		ctx context.Context,
		app *fiber.App,
		met *metrics.OTLPMetric,
	) error {
		// Initialize and run business logic modules
		return nil
	})
}
```

### Deconstructing the Template

This template demonstrates the framework's powerful design pattern:

1. **Handler Registration**: The `_ "..."` imports are used for their **side effects**. They automatically register OpenTelemetry (`otel`) and standard error (`stderr`) log handlers. The developer doesn't need to write any code to wire them up.
2. **Application Runner**: The `app` package is the imported runner from the `core` framework. It contains the logic for hosting the service.
3. **Configuration**: This step is where you would call your application's code to load configuration from environment variables or files.
4. **`app.WebMain`**: This is the heart of the framework. It's the main entry point that accepts your business logic as a **callback function**. It performs all the heavy lifting:
    - Creates and configures a `fiber.App` instance.
    - Sets up the `OTLPMetric` collector for observability.
    - Starts the web server on the configured port.
    - Listens for OS signals to trigger a **graceful shutdown**.
5. **Business Logic Injection**: The callback function is the designated place for your code. The framework **passes the fully initialized `fiber.App` and `metrics.OTLPMetric` objects to you**. Here, you would define your API routes, initialize services, and connect to databases.

## Core Concepts

The `InsideGallery/core` framework is built upon a few key architectural concepts. Understanding them will help you use the library effectively and build robust, maintainable applications.

### The Application Runner (`app`)

The **`app` package** is the heart of the framework. Its primary role is to act as a **host** for your application, managing its entire lifecycle. The `app.WebMain` function you used in the "Getting Started" guide is a concrete implementation of this concept.

The runner is responsible for:

- **Boilerplate Reduction**: It handles repetitive setup tasks like creating a web server, configuring ports, and setting up listeners.
- **Lifecycle Management**: It correctly starts the application and, most importantly, manages **graceful shutdown**. It listens for operating system signals (like `Ctrl+C`) to ensure that in-flight requests are completed before the application exits.
- **Dependency Injection**: It follows the **Inversion of Control (IoC)** principle by preparing essential objects (like the `fiber.App` and `metrics.OTLPMetric` collector) and passing them to your business logic via a callback. This cleanly separates the framework's concerns from your application's concerns.

### Integrated Observability (`fastlog`)

In any distributed system, understanding what's happening inside your services is critical. The `fastlog` package provides a powerful, integrated observability solution out of the box.

- **Structured Logging**: All logs are structured (e.g., JSON format), which is essential for modern log analysis tools.
- **Pluggable Handlers**: The logging system uses a **plugin-style architecture**. By simply importing a package for its side effects (e.g., `_ "github.com/InsideGallery/core/fastlog/handlers/otel"`), you automatically register a new output for the logger. This makes it trivial to send logs to multiple destinations (like the console, OpenTelemetry, etc.) without changing your application code.
- **Metrics**: The framework also initializes and provides a metrics collector (`metrics.OTLPMetric`), allowing your business logic to record key performance indicators for monitoring and alerting.

### The Swiss Army Knife: Common Packages

Beyond the application runner and observability, `core` provides a suite of common, reusable packages designed to be a **"Swiss army knife"** for Go development. These are utilities and wrappers that solve common problems, so you don't have to build them from scratch.

This includes, but is not limited to:

- **Common Algorithms**: A collection of optimized functions for frequent tasks.
- **Database Wrappers**: Simplified clients or helpers for interacting with various databases.
- **Configuration Helpers**: Tools to streamline loading configuration from files or the environment.