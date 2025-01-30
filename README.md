
# MultiDialogo - MailCulator server

## Provisioning

This Dockerfile is designed to build and deploy the `mailculator server` application using three distinct stages:
1. Builder Stage
2. Development Stage
3. Production Stage

Each stage serves a specific purpose, and you can use them based on your needs.

### Stage 1: Builder

Purpose:
This stage is responsible for building the Go application, running tests, and preparing it for further stages (Development or Production).

Description:
- The base image used is `golang:1.23`.
- The `mailculator server` project is copied into the container and the necessary dependencies are downloaded using `go mod tidy` and `go mod download`.
- The tests are run with `go test ./...` to ensure everything is correct.
- The application is built with `go build` and the resulting binary is copied to `/usr/local/bin/mailculator-server`.
- Finally, the binary is made executable with `chmod +x`.

To build the image:
```bash
 docker build --no-cache -t mailculators-builder --target mailculators-builder .
 ```

To introspect the builder image:

```bash
docker run -ti --rm mailculators-builder bash
```

### Stage 2: Development

Purpose: This stage is used for local development. It allows you to run the mailculator application with live-reload using air.

Description:

The base image used is golang:1.23.
The binary generated in the builder stage is copied into this container.
The air tool is installed, which enables live-reload for Go projects during development.
The container exposes port 8080, which can be used for development and debugging.
To build the image:
```bash
docker build -t mailculators-dev --target mailculators-dev .
```

To run the development container:
```bash
docker run -v$(pwd)/data:/var/lib/mailculator-server -p 8080:8080 mailculators-dev
```

### Stage 3: Production

Purpose: This stage is optimized for production deployment. It creates a minimal container to run the mailculator server binary in a secure and efficient environment.

Description:

The base image used is gcr.io/distroless/base-debian12, which is a minimal image without unnecessary tools or packages.
The binary from the builder stage is copied into the container.
Port 8080 is exposed for production use.
The container is configured to run the mailculator binary.
To build the image:
```bash
docker build -t mailculators-prod --target mailculators-prod .
```

### Example clients

```php
<?php

function validateEmailData(stdClass $payload): bool
{
    // Validate that the 'data' property exists and is an array
    if (empty($payload->data) || !is_array($payload->data)) {
        throw new InvalidArgumentException("Invalid or missing 'data'. It must be an array of email objects.");
    }

    // Validate each email in the 'data' array
    foreach ($payload->data as $email) {
        if (empty($email->id) || !is_string($email->id)) {
            throw new InvalidArgumentException("Invalid or missing 'id'. It must be a string.");
        }
        if (empty($email->type) || $email->type !== "email") {
            throw new InvalidArgumentException("Invalid or missing 'type'. It must be 'email'.");
        }
        if (empty($email->attributes) || !is_object($email->attributes)) {
            throw new InvalidArgumentException("Invalid or missing 'attributes'. It must be an object.");
        }

        $attributes = $email->attributes;

        // Validate email fields
        if (empty($attributes->from) || !filter_var($attributes->from, FILTER_VALIDATE_EMAIL)) {
            throw new InvalidArgumentException("Invalid or missing 'from' email address.");
        }
        if (!empty($attributes->replyTo) && !filter_var($attributes->replyTo, FILTER_VALIDATE_EMAIL)) {
            throw new InvalidArgumentException("Invalid 'replyTo' email address.");
        }
        if (empty($attributes->to) || !filter_var($attributes->to, FILTER_VALIDATE_EMAIL)) {
            throw new InvalidArgumentException("Invalid or missing 'to' email address.");
        }
        if (empty($attributes->subject) || !is_string($attributes->subject)) {
            throw new InvalidArgumentException("Invalid or missing 'subject'. It must be a string.");
        }
        if (empty($attributes->bodyHTML) && empty($attributes->bodyText)) {
            throw new InvalidArgumentException("At least one of 'bodyHTML' or 'bodyText' must be provided.");
        }

        // Attachments validation (if provided)
        if (!empty($attributes->attachments) && !is_array($attributes->attachments)) {
            throw new InvalidArgumentException("'attachments' must be an array of file paths.");
        }

        // Custom headers validation (if provided)
        if (!empty($attributes->customHeaders) && !is_object($attributes->customHeaders)) {
            throw new InvalidArgumentException("'customHeaders' must be an object.");
        }
    }

    return true; // Validation passed for all emails
}

function createEmailQueue(string $apiUrl, stdClass $payload): array
{
    // Validate payload
    try {
        validateEmailData($payload);
    } catch (InvalidArgumentException $e) {
        throw $e; // Rethrow invalid argument exceptions
    }

    // Initialize cURL
    $ch = curl_init();

    // Set cURL options
    curl_setopt($ch, CURLOPT_URL, $apiUrl . "/email-queues");
    curl_setopt($ch, CURLOPT_POST, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, [
        "Content-Type: application/json"
    ]);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));

    // Execute the request and fetch response
    $response = curl_exec($ch);

    // Check for cURL errors
    if (curl_errno($ch)) {
        throw new RuntimeException("cURL error: " . curl_error($ch));
    }

    // Get the HTTP response code
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);

    // Close cURL
    curl_close($ch);

    // Handle response
    if ($httpCode === 201) {
        return json_decode($response, true); // Success
    } elseif ($httpCode === 400) {
        throw new InvalidArgumentException("Invalid request: " . $response);
    } elseif ($httpCode === 405) {
        throw new RuntimeException("Invalid HTTP method");
    } elseif ($httpCode === 500) {
        throw new RuntimeException("Internal server error");
    } else {
        throw new RuntimeException("Unexpected response code: $httpCode, response: $response");
    }
}

// Example usage
try {
    $apiUrl = "https://api.mailculator.com"; // Replace with your actual API URL

    // Create payload as an object with 'data' as an array of stdClass objects
    $payload = (object)[
        "data" => [
            (object)[
                "id" => "user123:queue456:message789",
                "type" => "email",
                "attributes" => (object)[
                    "from" => "sender@example.com",
                    "replyTo" => "replyto@example.com",
                    "to" => "recipient@example.com",
                    "subject" => "Test Email 1",
                    "bodyHTML" => "<p>This is a test email 1.</p>",
                    "bodyText" => "This is a test email 1.",
                    "attachments" => ["/path/to/attachment1.txt", "/path/to/attachment2.jpg"],
                    "customHeaders" => (object)[
                        "X-Custom-Header" => "CustomValue1"
                    ]
                ]
            ],
            (object)[
                "id" => "user123:queue456:message790",
                "type" => "email",
                "attributes" => (object)[
                    "from" => "sender2@example.com",
                    "replyTo" => "replyto2@example.com",
                    "to" => "recipient2@example.com",
                    "subject" => "Test Email 2",
                    "bodyHTML" => "<p>This is a test email 2.</p>",
                    "bodyText" => "This is a test email 2.",
                    "attachments" => ["/path/to/attachment3.pdf"],
                    "customHeaders" => (object)[
                        "X-Custom-Header" => "CustomValue2"
                    ]
                ]
            ]
        ]
    ];

    $result = createEmailQueue($apiUrl, $payload);
    echo "Email queue created successfully:\n";
    print_r($result);
} catch (InvalidArgumentException $e) {
    echo "Validation Error: " . $e->getMessage();
} catch (RuntimeException $e) {
    echo "Runtime Error: " . $e->getMessage();
} catch (Exception $e) {
    echo "General Error: " . $e->getMessage();
}

```
