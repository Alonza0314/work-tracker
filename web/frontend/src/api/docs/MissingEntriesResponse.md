# MissingEntriesResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**message** | **string** |  | [default to undefined]
**from** | **string** |  | [default to undefined]
**to** | **string** |  | [default to undefined]
**workdays** | **number** | Workdays in the range. | [default to undefined]
**checkedCount** | **number** | Accounts checked (everyone but the system admin). | [default to undefined]
**members** | [**Array&lt;MissingMember&gt;**](MissingMember.md) |  | [default to undefined]

## Example

```typescript
import { MissingEntriesResponse } from './api';

const instance: MissingEntriesResponse = {
    message,
    from,
    to,
    workdays,
    checkedCount,
    members,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
