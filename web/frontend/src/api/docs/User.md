# User


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**account** | **string** |  | [default to undefined]
**name** | **string** |  | [default to undefined]
**role** | [**Role**](Role.md) |  | [default to undefined]
**i18n** | [**I18n**](I18n.md) |  | [default to undefined]
**isSystem** | **boolean** | True for the admin defined in the backend config file. | [default to undefined]

## Example

```typescript
import { User } from './api';

const instance: User = {
    account,
    name,
    role,
    i18n,
    isSystem,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
