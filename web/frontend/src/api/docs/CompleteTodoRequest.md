# CompleteTodoRequest

categoryId, hours and projectId fill or override the todo\'s; the resulting record needs a category and hours.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**date** | **string** |  | [default to undefined]
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
