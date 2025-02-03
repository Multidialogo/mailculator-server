# OpenAPI\Client\DefaultApi

All URIs are relative to http://localhost, except if the operation defines another base path.

| Method | HTTP request | Description |
| ------------- | ------------- | ------------- |
| [**createEmailQueue()**](DefaultApi.md#createEmailQueue) | **POST** /email-queues | Create a new email queue |


## `createEmailQueue()`

```php
createEmailQueue($create_email_queue_request): \OpenAPI\Client\Model\CreateEmailQueue201Response
```

Create a new email queue

Receives email data and saves it to the mail queue.

### Example

```php
<?php
require_once(__DIR__ . '/vendor/autoload.php');



$apiInstance = new OpenAPI\Client\Api\DefaultApi(
    // If you want use custom http client, pass your client which implements `GuzzleHttp\ClientInterface`.
    // This is optional, `GuzzleHttp\Client` will be used as default.
    new GuzzleHttp\Client()
);
$create_email_queue_request = new \OpenAPI\Client\Model\CreateEmailQueueRequest(); // \OpenAPI\Client\Model\CreateEmailQueueRequest

try {
    $result = $apiInstance->createEmailQueue($create_email_queue_request);
    print_r($result);
} catch (Exception $e) {
    echo 'Exception when calling DefaultApi->createEmailQueue: ', $e->getMessage(), PHP_EOL;
}
```

### Parameters

| Name | Type | Description  | Notes |
| ------------- | ------------- | ------------- | ------------- |
| **create_email_queue_request** | [**\OpenAPI\Client\Model\CreateEmailQueueRequest**](../Model/CreateEmailQueueRequest.md)|  | |

### Return type

[**\OpenAPI\Client\Model\CreateEmailQueue201Response**](../Model/CreateEmailQueue201Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/vnd.api+json`

[[Back to top]](#) [[Back to API list]](../../README.md#endpoints)
[[Back to Model list]](../../README.md#models)
[[Back to README]](../../README.md)
