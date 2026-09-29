# DefaultApi

All URIs are relative to *http://127.0.0.1:5000*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**changeMyPassword**](#changemypassword) | **PUT** /api/me/password | Change the logged-in user\&#39;s password|
|[**createUser**](#createuser) | **POST** /api/users | Create a user (admin)|
|[**deleteUser**](#deleteuser) | **DELETE** /api/users/{account} | Delete a user (admin)|
|[**getMe**](#getme) | **GET** /api/me | Get the logged-in user|
|[**listUsers**](#listusers) | **GET** /api/users | List users (admin)|
|[**login**](#login) | **POST** /api/login | Login|
|[**logout**](#logout) | **POST** /api/logout | Logout|
|[**updateMe**](#updateme) | **PUT** /api/me | Update the logged-in user\&#39;s preferences|
|[**updateUser**](#updateuser) | **PUT** /api/users/{account} | Update a user (admin)|

# **changeMyPassword**
> MessageResponse changeMyPassword(changeMyPasswordRequest)

Not allowed for the system (config) admin.

### Example

```typescript
import {
    DefaultApi,
    Configuration,
    ChangeMyPasswordRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let changeMyPasswordRequest: ChangeMyPasswordRequest; //

const { status, data } = await apiInstance.changeMyPassword(
    changeMyPasswordRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **changeMyPasswordRequest** | **ChangeMyPasswordRequest**|  | |


### Return type

**MessageResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createUser**
> UserResponse createUser(createUserRequest)

The new user\'s initial password is its account.

### Example

```typescript
import {
    DefaultApi,
    Configuration,
    CreateUserRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let createUserRequest: CreateUserRequest; //

const { status, data } = await apiInstance.createUser(
    createUserRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createUserRequest** | **CreateUserRequest**|  | |


### Return type

**UserResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**409** | Conflict |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteUser**
> MessageResponse deleteUser()

The system admin and the caller themself cannot be deleted.

### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let account: string; // (default to undefined)

const { status, data } = await apiInstance.deleteUser(
    account
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **account** | [**string**] |  | defaults to undefined|


### Return type

**MessageResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getMe**
> UserResponse getMe()


### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.getMe();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**UserResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**401** | Unauthorized |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listUsers**
> ListUsersResponse listUsers()


### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.listUsers();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**ListUsersResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **login**
> LoginResponse login(loginRequest)

The returned JWT carries `sub` (account), `name`, `role` and `i18n` claims.

### Example

```typescript
import {
    DefaultApi,
    Configuration,
    LoginRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let loginRequest: LoginRequest; //

const { status, data } = await apiInstance.login(
    loginRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **loginRequest** | **LoginRequest**|  | |


### Return type

**LoginResponse**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **logout**
> logout()


### Example

```typescript
import {
    DefaultApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

const { status, data } = await apiInstance.logout();
```

### Parameters
This endpoint does not have any parameters.


### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**204** | No Content |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateMe**
> TokenResponse updateMe(updateMeRequest)

Returns a new JWT carrying the updated `i18n` claim.

### Example

```typescript
import {
    DefaultApi,
    Configuration,
    UpdateMeRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let updateMeRequest: UpdateMeRequest; //

const { status, data } = await apiInstance.updateMe(
    updateMeRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateMeRequest** | **UpdateMeRequest**|  | |


### Return type

**TokenResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateUser**
> UserResponse updateUser(updateUserRequest)

Only the given fields are changed. The system admin cannot be updated.

### Example

```typescript
import {
    DefaultApi,
    Configuration,
    UpdateUserRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new DefaultApi(configuration);

let account: string; // (default to undefined)
let updateUserRequest: UpdateUserRequest; //

const { status, data } = await apiInstance.updateUser(
    account,
    updateUserRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateUserRequest** | **UpdateUserRequest**|  | |
| **account** | [**string**] |  | defaults to undefined|


### Return type

**UserResponse**

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | OK |  -  |
|**400** | Bad Request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not Found |  -  |
|**500** | Internal Server Error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

