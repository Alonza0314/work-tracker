# UpdateWorkSettingRequest

Only the given fields are changed.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**allowViewAll** | **boolean** |  | [optional] [default to undefined]
**startDate** | **string** | When the team started logging (YYYY-MM-DD); missed entries are not checked before it. An empty string clears it. | [optional] [default to undefined]

## Example

```typescript
import { UpdateWorkSettingRequest } from './api';

const instance: UpdateWorkSettingRequest = {
    allowViewAll,
    startDate,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
