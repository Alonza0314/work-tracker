# ApiTokenInfo


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** |  | [default to undefined]
**name** | **string** |  | [default to undefined]
**prefix** | **string** | The first characters of the token, to recognize it. | [default to undefined]
**createdAt** | **string** |  | [default to undefined]
**expiresAt** | **string** |  | [default to undefined]
**lastUsedAt** | **string** | Missing until the token is first used; updated at most once a minute. | [optional] [default to undefined]

## Example

```typescript
import { ApiTokenInfo } from './api';

const instance: ApiTokenInfo = {
    id,
    name,
    prefix,
    createdAt,
    expiresAt,
    lastUsedAt,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
