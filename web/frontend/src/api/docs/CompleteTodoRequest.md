# CompleteTodoRequest

categoryId, hours and projectId fill or override the todo\'s.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**date** | **string** | The client\&#39;s today, used only when the todo has no date (else the server\&#39;s today). | [optional] [default to undefined]
**categoryId** | **string** |  | [optional] [default to undefined]
**hours** | **number** |  | [optional] [default to undefined]
**projectId** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { CompleteTodoRequest } from './api';

const instance: CompleteTodoRequest = {
    date,
    categoryId,
    hours,
    projectId,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
