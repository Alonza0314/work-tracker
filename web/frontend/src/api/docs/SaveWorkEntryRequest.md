# SaveWorkEntryRequest

Work records require categoryId and hours; todos only require date and description. projectId is always optional.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**date** | **string** |  | [default to undefined]
**categoryId** | **string** |  | [optional] [default to undefined]
**description** | **string** |  | [default to undefined]
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
