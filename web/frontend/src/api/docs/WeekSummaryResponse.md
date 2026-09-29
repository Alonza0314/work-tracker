# WeekSummaryResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**message** | **string** |  | [default to undefined]
**from** | **string** |  | [default to undefined]
**to** | **string** |  | [default to undefined]
**recordCount** | **number** |  | [default to undefined]
**loggedHours** | **number** |  | [default to undefined]
**requiredHours** | **number** |  | [default to undefined]
**remainingHours** | **number** | requiredHours - loggedHours, never below 0. | [default to undefined]
**daysOff** | **number** | Weekdays off in the week. | [default to undefined]
**days** | [**Array&lt;WeekDay&gt;**](WeekDay.md) |  | [default to undefined]

## Example

```typescript
import { WeekSummaryResponse } from './api';

const instance: WeekSummaryResponse = {
    message,
    from,
    to,
    recordCount,
    loggedHours,
    requiredHours,
    remainingHours,
    daysOff,
    days,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
