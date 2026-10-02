# SaveWorkEntryRequest

Every field is optional. A work record without a date is dated the server\'s today; a todo may have no date. Unset hours count as 0.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**date** | **string** |  | [optional] [default to undefined]
**categoryId** | **string** |  | [optional] [default to undefined]
**description** | **string** |  | [optional] [default to undefined]
**hours** | **number** |  | [optional] [default to undefined]
**projectId** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { SaveWorkEntryRequest } from './api';

const instance: SaveWorkEntryRequest = {
    date,
    categoryId,
    description,
    hours,
    projectId,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
