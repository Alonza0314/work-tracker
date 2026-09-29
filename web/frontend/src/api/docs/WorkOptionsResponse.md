# WorkOptionsResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**message** | **string** |  | [default to undefined]
**categories** | [**Array&lt;WorkOption&gt;**](WorkOption.md) |  | [default to undefined]
**projects** | [**Array&lt;WorkOption&gt;**](WorkOption.md) |  | [default to undefined]
**allowViewAll** | **boolean** |  | [default to undefined]
**startDate** | **string** | Missing when not set. | [optional] [default to undefined]

## Example

```typescript
import { WorkOptionsResponse } from './api';

const instance: WorkOptionsResponse = {
    message,
    categories,
    projects,
    allowViewAll,
    startDate,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
