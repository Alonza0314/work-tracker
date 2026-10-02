# WorkRecord

Optional fields are empty (\"\" / 0) when not set. Every field is optional, but records always have a date; todos may not.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** |  | [default to undefined]
**account** | **string** |  | [default to undefined]
**date** | **string** |  | [default to undefined]
**categoryId** | **string** |  | [default to undefined]
**description** | **string** |  | [default to undefined]
**hours** | **number** |  | [default to undefined]
**projectId** | **string** |  | [default to undefined]
**createdAt** | **string** |  | [default to undefined]

## Example

```typescript
import { WorkRecord } from './api';

const instance: WorkRecord = {
    id,
    account,
    date,
    categoryId,
    description,
    hours,
    projectId,
    createdAt,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
